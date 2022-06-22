package server

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/config"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/domain"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/service"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/metrics"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/interceptors"
	kafkaClient "github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/kafka"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/logger"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/middlewares"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/mongodb"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/tracing"
	"github.com/go-playground/validator"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
	v7 "github.com/olivere/elastic/v7"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	maxConnectionIdle = 5
	gRPCTimeout       = 15
	maxConnectionAge  = 5
	gRPCTime          = 10
)

type server struct {
	log           logger.Logger
	cfg           config.Config
	im            interceptors.InterceptorManager
	mw            middlewares.MiddlewareManager
	probesSrv     *http.Server
	v             *validator.Validate
	metrics       *metrics.ESMicroserviceMetrics
	kafkaConn     *kafka.Conn
	pgxConn       *pgxpool.Pool
	mongoClient   *mongo.Client
	doneCh        chan struct{}
	elasticClient *v7.Client
	echo          *echo.Echo
	ps            *http.Server
	bs            *service.BankAccountService
}

func NewServer(log logger.Logger, cfg config.Config) *server {
	return &server{log: log, cfg: cfg, v: validator.New(), doneCh: make(chan struct{}), echo: echo.New()}
}

func (s *server) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	// enable tracing
	if s.cfg.Jaeger.Enable {
		tracer, closer, err := tracing.NewJaegerTracer(s.cfg.Jaeger)
		if err != nil {
			return err
		}
		defer closer.Close() // nolint: errcheck
		opentracing.SetGlobalTracer(tracer)
	}

	s.metrics = metrics.NewESMicroserviceMetrics(s.cfg)
	s.im = interceptors.NewInterceptorManager(s.log, s.getGrpcMetricsCb())
	s.mw = middlewares.NewMiddlewareManager(s.log, s.cfg, s.getHttpMetricsCb())

	// connect postgres
	if err := s.connectPostgres(ctx); err != nil {
		return err
	}
	defer s.pgxConn.Close()

	// connect mongo
	mongoDBConn, err := mongodb.NewMongoDBConn(ctx, s.cfg.Mongo)
	if err != nil {
		return errors.Wrap(err, "NewMongoDBConn")
	}
	s.mongoClient = mongoDBConn
	defer mongoDBConn.Disconnect(ctx) // nolint: errcheck
	s.log.Infof("(Mongo connected) SessionsInProgress: %v", mongoDBConn.NumberSessionsInProgress())

	// connect elastic
	if err := s.initElasticClient(ctx); err != nil {
		s.log.Errorf("(initElasticClient) err: %v", err)
		return err
	}

	// connect kafka brokers
	if err := s.connectKafkaBrokers(ctx); err != nil {
		return errors.Wrap(err, "s.connectKafkaBrokers")
	}
	defer s.kafkaConn.Close() // nolint: errcheck

	// init kafka topics
	if s.cfg.Kafka.InitTopics {
		s.initKafkaTopics(ctx)
	}

	// kafka producer
	kafkaProducer := kafkaClient.NewProducer(s.log, s.cfg.Kafka.Brokers)
	defer kafkaProducer.Close() // nolint: errcheck

	eventBus := es.NewKafkaEventsBus(kafkaProducer, s.cfg.KafkaPublisherConfig)
	eventStore := es.NewPgEventStore(s.log, s.cfg.EventSourcingConfig, s.pgxConn, eventBus, domain.NewEventSerializer())
	s.bs = service.NewBankAccountService(s.log, eventStore)

	closeGrpcServer, grpcServer, err := s.newBankAccountGrpcServer()
	if err != nil {
		cancel()
		return err
	}
	defer closeGrpcServer() // nolint: errcheck

	// run metrics and health check
	s.runMetrics(cancel)
	s.runHealthCheck(ctx)

	go func() {
		if err := s.runHttpServer(); err != nil {
			s.log.Errorf("(s.runHttpServer) err: %v", err)
			cancel()
		}
	}()
	s.log.Infof("%s is listening on PORT: %s", GetMicroserviceName(s.cfg), s.cfg.Http.Port)

	<-ctx.Done()
	s.waitShootDown(waitShotDownDuration)
	grpcServer.GracefulStop()
	if err := s.shutDownHealthCheckServer(ctx); err != nil {
		s.log.Warnf("(shutDownHealthCheckServer) err: %v", err)
	}

	if err := s.echo.Shutdown(ctx); err != nil {
		s.log.Warnf("(Shutdown) err: %v", err)
	}

	<-s.doneCh
	s.log.Infof("%s server exited properly", GetMicroserviceName(s.cfg))
	return nil
}
