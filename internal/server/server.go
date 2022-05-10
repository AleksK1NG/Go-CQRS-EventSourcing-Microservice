package server

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/config"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/metrics"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/interceptors"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/logger"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/middlewares"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/tracing"
	"github.com/go-playground/validator"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
	v7 "github.com/olivere/elastic/v7"
	"github.com/opentracing/opentracing-go"
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
}

func NewServer(log logger.Logger, cfg config.Config) *server {
	return &server{log: log, cfg: cfg, v: validator.New(), doneCh: make(chan struct{})}
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

	<-ctx.Done()
	return nil
}
