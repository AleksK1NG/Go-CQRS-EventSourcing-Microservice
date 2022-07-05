package app

import (
	"fmt"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/config"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/migrations"
	"strings"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	waitShotDownDuration = 3 * time.Second
)

func (a *app) getHttpMetricsCb() func(err error) {
	return func(err error) {
		if err != nil {
			a.metrics.ErrorHttpRequests.Inc()
		} else {
			a.metrics.SuccessHttpRequests.Inc()
		}
	}
}

func (a *app) getGrpcMetricsCb() func(err error) {
	return func(err error) {
		if err != nil {
			a.metrics.ErrorGrpcRequests.Inc()
		} else {
			a.metrics.SuccessGrpcRequests.Inc()
		}
	}
}

func (a *app) waitShootDown(duration time.Duration) {
	go func() {
		time.Sleep(duration)
		a.doneCh <- struct{}{}
	}()
}

func GetMicroserviceName(cfg config.Config) string {
	return fmt.Sprintf("(%s)", strings.ToUpper(cfg.ServiceName))
}

func (a *app) runMigrate() error {
	//m, err := migrate.New(
	//	"file://migrations",
	//	"postgres://postgres:postgres@localhost:5432/bank_accounts?sslmode=disable")
	//if err != nil {
	//	return err
	//}
	//
	//if err := m.Down(); err != nil {
	//	return err
	//}
	//
	//if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
	//	return err
	//}
	a.log.Infof("Run migrations with config: %+v", a.cfg.MigrationsConfig)

	//if err := migrations.RunMigrations(a.cfg.MigrationsConfig); err != nil {
	//	a.log.Errorf("RunMigrations err: %v", err)
	//	return err
	//}

	version, dirty, err := migrations.RunMigrations(a.cfg.MigrationsConfig)
	if err != nil {
		a.log.Errorf("RunMigrations err: %v", err)
		return err
	}

	a.log.Infof("Migrations successfully created: version: %d, dirty: %v", version, dirty)
	return nil
}
