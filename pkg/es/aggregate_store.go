package es

import (
	"context"
	"encoding/json"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/tracing"
	"github.com/jackc/pgx/v4"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
)

// Load es.Aggregate events using snapshots with given frequency
func (p *pgEventStore[T]) Load(ctx context.Context, aggregate T) (T, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pgEventStore.Load")
	defer span.Finish()
	span.LogFields(log.String("aggregate", aggregate.String()))

	snapshot, err := p.GetSnapshot(ctx, aggregate.GetID())
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		tracing.TraceErr(span, err)
		return aggregate, err
	}

	if snapshot != nil {
		if err := json.Unmarshal(snapshot.State, aggregate); err != nil {
			tracing.TraceErr(span, err)
			p.log.Errorf("(Load) json.Unmarshal err: %v", err)
			return aggregate, errors.Wrap(err, "json.Unmarshal")
		}

		err := p.loadAggregateEventsByVersion(ctx, aggregate)
		if err != nil {
			return aggregate, err
		}

		p.log.Debugf("(Load Aggregate): aggregate: {%s}", aggregate.String())
		span.LogFields(log.String("aggregate with events", aggregate.String()))
		return aggregate, nil
	}

	err = p.loadEvents(ctx, aggregate)
	if err != nil {
		return aggregate, err
	}

	p.log.Debugf("(Load Aggregate): aggregate: {%s}", aggregate.String())
	span.LogFields(log.String("aggregate with events", aggregate.String()))
	return aggregate, nil
}

// Save es.Aggregate events using snapshots with given frequency
func (p *pgEventStore[T]) Save(ctx context.Context, aggregate T) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pgEventStore.Save")
	defer span.Finish()
	span.LogFields(log.String("aggregate", aggregate.String()))

	if len(aggregate.GetChanges()) == 0 {
		p.log.Debug("(Save) aggregate.GetUncommittedEvents()) == 0")
		span.LogFields(log.Int("events", len(aggregate.GetChanges())))
		return nil
	}

	tx, err := p.db.Begin(ctx)
	if err != nil {
		tracing.TraceErr(span, err)
		p.log.Errorf("(Save) db.Begin err: %v", err)
		return errors.Wrap(err, "db.Begin")
	}

	defer func() {
		if tx != nil {
			if txErr := tx.Rollback(ctx); txErr != nil && !errors.Is(txErr, pgx.ErrTxClosed) {
				err = txErr
				tracing.TraceErr(span, err)
				return
			}
		}
	}()

	events := aggregate.GetChanges()

	if err := p.saveEventsTx(ctx, tx, events); err != nil {
		return err
	}

	if aggregate.GetVersion()%p.cfg.SnapshotFrequency == 0 {
		aggregate.ToSnapshot()
		if err := p.saveSnapshotTx(ctx, tx, aggregate); err != nil {
			return err
		}
	}

	if err := p.processEvents(ctx, events); err != nil {
		tracing.TraceErr(span, err)
		return errors.Wrap(err, "processEvents")
	}

	p.log.Debugf("(Save Aggregate): aggregate: {%s}", aggregate.String())
	span.LogFields(log.String("aggregate with events", aggregate.String()))
	return tx.Commit(ctx)
}
