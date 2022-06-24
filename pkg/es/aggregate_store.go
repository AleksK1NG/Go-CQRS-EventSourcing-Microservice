package es

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es/serializer"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/tracing"
	"github.com/jackc/pgx/v4"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
)

// Load es.Aggregate events using snapshots with given frequency
func (p *pgEventStore) Load(ctx context.Context, aggregate Aggregate) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pgEventStore.Load")
	defer span.Finish()
	span.LogFields(log.String("aggregate", aggregate.String()))

	snapshot, err := p.GetSnapshot(ctx, aggregate.GetID())
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return tracing.TraceWithErr(span, err)
	}

	if snapshot != nil {
		if err := serializer.Unmarshal(snapshot.State, aggregate); err != nil {
			p.log.Errorf("(Load) serializer.Unmarshal err: %v", err)
			return tracing.TraceWithErr(span, errors.Wrap(err, "json.Unmarshal"))
		}

		err := p.loadAggregateEventsByVersion(ctx, aggregate)
		if err != nil {
			return err
		}

		p.log.Debugf("(Load Aggregate By Version) aggregate: %s", aggregate.String())
		span.LogFields(log.String("aggregate with events", aggregate.String()))
		return nil
	}

	err = p.loadEvents(ctx, aggregate)
	if err != nil {
		return err
	}

	p.log.Debugf("(Load Aggregate): aggregate: %s", aggregate.String())
	span.LogFields(log.String("aggregate with events", aggregate.String()))
	return nil
}

// Save es.Aggregate events using snapshots with given frequency
func (p *pgEventStore) Save(ctx context.Context, aggregate Aggregate) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pgEventStore.Save")
	defer span.Finish()
	span.LogFields(log.String("aggregate", aggregate.String()))

	if len(aggregate.GetChanges()) == 0 {
		p.log.Debug("(Save) aggregate.GetChanges()) == 0")
		span.LogFields(log.Int("events", len(aggregate.GetChanges())))
		return nil
	}

	tx, err := p.db.Begin(ctx)
	if err != nil {
		p.log.Errorf("(Save) db.Begin err: %v", err)
		return tracing.TraceWithErr(span, errors.Wrap(err, "db.Begin"))
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

	changes := aggregate.GetChanges()
	events := make([]Event, 0, len(changes))

	for i := range changes {
		event, err := p.serializer.SerializeEvent(aggregate, changes[i])
		if err != nil {
			p.log.Errorf("(Save) serializer.SerializeEvent err: %v", err)
			return tracing.TraceWithErr(span, errors.Wrap(err, "serializer.SerializeEvent"))
		}
		events = append(events, event)
	}

	if err := p.saveEventsTx(ctx, tx, events); err != nil {
		return tracing.TraceWithErr(span, errors.Wrap(err, "saveEventsTx"))
	}

	if aggregate.GetVersion()%p.cfg.SnapshotFrequency == 0 {
		aggregate.ToSnapshot()
		if err := p.saveSnapshotTx(ctx, tx, aggregate); err != nil {
			return tracing.TraceWithErr(span, errors.Wrap(err, "saveSnapshotTx"))
		}
	}

	if err := p.processEvents(ctx, events); err != nil {
		return tracing.TraceWithErr(span, errors.Wrap(err, "processEvents"))
	}

	p.log.Debugf("(Save Aggregate): aggregate: %s", aggregate.String())
	span.LogFields(log.String("aggregate with events", aggregate.String()))
	return tx.Commit(ctx)
}
