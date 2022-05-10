package server

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/postgres"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/utils"
	"time"

	"github.com/avast/retry-go"
	"github.com/pkg/errors"
)

func (s *server) connectPostgres(ctx context.Context) error {

	retryOptions := []retry.Option{
		retry.Attempts(s.cfg.Timeouts.PostgresInitRetryCount),
		retry.Delay(time.Duration(s.cfg.Timeouts.PostgresInitMilliseconds) * time.Millisecond),
		retry.DelayType(retry.BackOffDelay),
		retry.LastErrorOnly(true),
		retry.Context(ctx),
		retry.OnRetry(func(n uint, err error) {
			s.log.Errorf("retry connect postgres err: %v", err)
		}),
	}

	return retry.Do(func() error {
		pgxConn, err := postgres.NewPgxConn(s.cfg.Postgresql)
		if err != nil {
			return errors.Wrap(err, "postgresql.NewPgxConn")
		}
		s.pgxConn = pgxConn
		s.log.Infof("(postgres connected) poolStat: %s", utils.GetPostgresStats(s.pgxConn.Stat()))
		return nil
	}, retryOptions...)
}
