package app

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/postgres"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/utils"
	"time"

	"github.com/avast/retry-go"
	"github.com/pkg/errors"
)

func (a *app) connectPostgres(ctx context.Context) error {

	retryOptions := []retry.Option{
		retry.Attempts(a.cfg.Timeouts.PostgresInitRetryCount),
		retry.Delay(time.Duration(a.cfg.Timeouts.PostgresInitMilliseconds) * time.Millisecond),
		retry.DelayType(retry.BackOffDelay),
		retry.LastErrorOnly(true),
		retry.Context(ctx),
		retry.OnRetry(func(n uint, err error) {
			a.log.Errorf("retry connect postgres err: %v", err)
		}),
	}

	return retry.Do(func() error {
		pgxConn, err := postgres.NewPgxConn(a.cfg.Postgresql)
		if err != nil {
			return errors.Wrap(err, "postgresql.NewPgxConn")
		}
		a.pgxConn = pgxConn
		a.log.Infof("(postgres connected) poolStat: %s", utils.GetPostgresStats(a.pgxConn.Stat()))
		return nil
	}, retryOptions...)
}
