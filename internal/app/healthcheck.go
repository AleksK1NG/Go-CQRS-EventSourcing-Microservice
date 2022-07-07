package app

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/constants"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/esclient"
	"github.com/heptiolabs/healthcheck"
	"net/http"
	"time"
)

func (a *app) runHealthCheck(ctx context.Context) {
	health := healthcheck.NewHandler()

	mux := http.NewServeMux()
	a.probeServer = &http.Server{
		Handler:      mux,
		Addr:         a.cfg.Probes.Port,
		WriteTimeout: writeTimeout,
		ReadTimeout:  readTimeout,
	}
	mux.HandleFunc(a.cfg.Probes.LivenessPath, health.LiveEndpoint)
	mux.HandleFunc(a.cfg.Probes.ReadinessPath, health.ReadyEndpoint)

	a.configureHealthCheckEndpoints(ctx, health)

	go func() {
		a.log.Infof("(%s) Kubernetes probes listening on port: %s", a.cfg.ServiceName, a.cfg.Probes.Port)
		if err := a.probeServer.ListenAndServe(); err != nil {
			a.log.Errorf("(ListenAndServe) err: %v", err)
		}
	}()
}

func (a *app) configureHealthCheckEndpoints(ctx context.Context, health healthcheck.Handler) {

	health.AddReadinessCheck(constants.Postgres, healthcheck.AsyncWithContext(ctx, func() error {
		if err := a.pgxConn.Ping(ctx); err != nil {
			a.log.Warnf("(Postgres Readiness Check) err: %v", err)
			return err
		}
		return nil
	}, time.Duration(a.cfg.Probes.CheckIntervalSeconds)*time.Second))

	health.AddLivenessCheck(constants.Postgres, healthcheck.AsyncWithContext(ctx, func() error {

		if err := a.pgxConn.Ping(ctx); err != nil {
			a.log.Warnf("(Postgres Liveness Check) err: %v", err)
			return err
		}
		return nil
	}, time.Duration(a.cfg.Probes.CheckIntervalSeconds)*time.Second))

	health.AddReadinessCheck(constants.MongoDB, healthcheck.AsyncWithContext(ctx, func() error {
		if err := a.mongoClient.Ping(ctx, nil); err != nil {
			a.log.Warnf("(MongoDB Readiness Check) err: %v", err)
			return err
		}
		return nil
	}, time.Duration(a.cfg.Probes.CheckIntervalSeconds)*time.Second))

	health.AddLivenessCheck(constants.MongoDB, healthcheck.AsyncWithContext(ctx, func() error {
		if err := a.mongoClient.Ping(ctx, nil); err != nil {
			a.log.Warnf("(MongoDB Liveness Check) err: %v", err)
			return err
		}
		return nil
	}, time.Duration(a.cfg.Probes.CheckIntervalSeconds)*time.Second))

	health.AddReadinessCheck(constants.ElasticSearch, healthcheck.AsyncWithContext(ctx, func() error {
		_, err := esclient.Info(ctx, a.elasticClient)
		if err != nil {
			a.log.Warnf("(ElasticSearch Readiness Check) err: %v", err)
			return err
		}
		return nil
	}, time.Duration(a.cfg.Probes.CheckIntervalSeconds)*time.Second))

	health.AddLivenessCheck(constants.ElasticSearch, healthcheck.AsyncWithContext(ctx, func() error {
		_, err := esclient.Info(ctx, a.elasticClient)
		if err != nil {
			a.log.Warnf("(ElasticSearch Liveness Check) err: %v", err)
			return err
		}
		return nil
	}, time.Duration(a.cfg.Probes.CheckIntervalSeconds)*time.Second))

	health.AddReadinessCheck(constants.Kafka, healthcheck.AsyncWithContext(ctx, func() error {
		if _, err := a.kafkaConn.Brokers(); err != nil {
			a.log.Warnf("readiness kafka health check err: %v", err)
			return err
		}
		return nil
	}, time.Duration(a.cfg.Probes.CheckIntervalSeconds)*time.Second))

	health.AddLivenessCheck(constants.Kafka, healthcheck.AsyncWithContext(ctx, func() error {
		if _, err := a.kafkaConn.Brokers(); err != nil {
			a.log.Warnf("kafka health check err: %v", err)
			if a.kafkaConn != nil {
				if err := a.kafkaConn.Close(); err != nil {
					a.log.Warnf("kafkaConn.Close err: %v", err)
				}
				if err := a.connectKafkaBrokers(ctx); err != nil {
					a.log.Warnf("connectKafkaBrokers err: %v", err)
					return err
				}
			}
			return err
		}
		return nil
	}, time.Duration(a.cfg.Probes.CheckIntervalSeconds)*time.Second))
}

func (a *app) shutDownHealthCheckServer(ctx context.Context) error {
	return a.probeServer.Shutdown(ctx)
}
