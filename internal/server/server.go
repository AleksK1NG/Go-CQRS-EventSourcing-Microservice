package server

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/config"
	bankAccountMongoSubscription "github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/delivery/kafka"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/domain"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/projection/mongo_projection"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/repository/mongo_repository"
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

type app struct {
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

func NewApp(log logger.Logger, cfg config.Config) *app {
	return &app{log: log, cfg: cfg, v: validator.New(), doneCh: make(chan struct{}), echo: echo.New()}
}

func (a *app) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	// enable tracing
	if a.cfg.Jaeger.Enable {
		tracer, closer, err := tracing.NewJaegerTracer(a.cfg.Jaeger)
		if err != nil {
			return err
		}
		defer closer.Close() // nolint: errcheck
		opentracing.SetGlobalTracer(tracer)
	}

	a.metrics = metrics.NewESMicroserviceMetrics(a.cfg)
	a.im = interceptors.NewInterceptorManager(a.log, a.getGrpcMetricsCb())
	a.mw = middlewares.NewMiddlewareManager(a.log, a.cfg, a.getHttpMetricsCb())

	// connect postgres
	if err := a.connectPostgres(ctx); err != nil {
		return err
	}
	defer a.pgxConn.Close()

	// connect mongo
	mongoDBConn, err := mongodb.NewMongoDBConn(ctx, a.cfg.Mongo)
	if err != nil {
		return errors.Wrap(err, "NewMongoDBConn")
	}
	a.mongoClient = mongoDBConn
	defer mongoDBConn.Disconnect(ctx) // nolint: errcheck
	a.log.Infof("(Mongo connected) SessionsInProgress: %v", mongoDBConn.NumberSessionsInProgress())

	a.initMongoDBCollections(ctx)

	// connect elastic
	if err := a.initElasticClient(ctx); err != nil {
		a.log.Errorf("(initElasticClient) err: %v", err)
		return err
	}

	// connect kafka brokers
	if err := a.connectKafkaBrokers(ctx); err != nil {
		return errors.Wrap(err, "a.connectKafkaBrokers")
	}
	defer a.kafkaConn.Close() // nolint: errcheck

	// init kafka topics
	if a.cfg.Kafka.InitTopics {
		a.initKafkaTopics(ctx)
	}

	// kafka producer
	kafkaProducer := kafkaClient.NewProducer(a.log, a.cfg.Kafka.Brokers)
	defer kafkaProducer.Close() // nolint: errcheck

	eventSerializer := domain.NewEventSerializer()
	eventBus := es.NewKafkaEventsBus(kafkaProducer, a.cfg.KafkaPublisherConfig)
	eventStore := es.NewPgEventStore(a.log, a.cfg.EventSourcingConfig, a.pgxConn, eventBus, eventSerializer)
	a.bs = service.NewBankAccountService(a.log, eventStore)

	bankAccountMongoRepository := mongo_repository.NewBankAccountMongoRepository(a.log, &a.cfg, a.mongoClient)
	bankAccountMongoProjection := mongo_projection.NewBankAccountMongoProjection(a.log, &a.cfg, eventSerializer, bankAccountMongoRepository)

	mongoSubscription := bankAccountMongoSubscription.NewBankAccountMongoSubscription(a.log, &a.cfg, a.bs, bankAccountMongoProjection, eventSerializer)
	mongoConsumerGroup := kafkaClient.NewConsumerGroup(a.cfg.Kafka.Brokers, a.cfg.Projections.MongoGroup, a.log)
	go func() {
		err := mongoConsumerGroup.ConsumeTopicWithErrGroup(ctx, a.getConsumerGroupTopics(), 10, mongoSubscription.ProcessMessagesErrGroup)
		if err != nil {
			a.log.Errorf("(mongoConsumerGroup ConsumeTopicWithErrGroup) err: %v", err)
			cancel()
			return
		}
	}()

	closeGrpcServer, grpcServer, err := a.newBankAccountGrpcServer()
	if err != nil {
		cancel()
		return err
	}
	defer closeGrpcServer() // nolint: errcheck

	// run metrics and health check
	a.runMetrics(cancel)
	a.runHealthCheck(ctx)

	go func() {
		if err := a.runHttpServer(); err != nil {
			a.log.Errorf("(a.runHttpServer) err: %v", err)
			cancel()
		}
	}()
	a.log.Infof("%a is listening on PORT: %a", GetMicroserviceName(a.cfg), a.cfg.Http.Port)

	<-ctx.Done()
	a.waitShootDown(waitShotDownDuration)
	grpcServer.GracefulStop()
	if err := a.shutDownHealthCheckServer(ctx); err != nil {
		a.log.Warnf("(shutDownHealthCheckServer) err: %v", err)
	}

	if err := a.echo.Shutdown(ctx); err != nil {
		a.log.Warnf("(Shutdown) err: %v", err)
	}

	<-a.doneCh
	a.log.Infof("%a app exited properly", GetMicroserviceName(a.cfg))
	return nil
}
