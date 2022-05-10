package es

import (
	"context"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/logger"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/tracing"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
)

const (
	eventsCapacity = 10
)

type pgEventStore[T Aggregate] struct {
	log      logger.Logger
	cfg      Config
	db       *pgxpool.Pool
	eventBus EventsBus
}

func NewPgEventStore[T Aggregate](log logger.Logger, cfg Config, db *pgxpool.Pool, eventBus EventsBus) *pgEventStore[T] {
	return &pgEventStore[T]{log: log, cfg: cfg, db: db, eventBus: eventBus}
}

// SaveEvents save aggregate uncommitted events as one batch and process with event bus using transaction
func (p *pgEventStore[T]) SaveEvents(ctx context.Context, events []Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pgEventStore.SaveEvents")
	defer span.Finish()

	tx, err := p.db.Begin(ctx)
	if err != nil {
		tracing.TraceErr(span, err)
		p.log.Errorf("(SaveEvents) db.Begin err: %v", err)
		return errors.Wrap(err, "db.Begin")
	}

	if err := p.handleConcurrency(ctx, tx, events); err != nil {
		return RollBackTx(ctx, tx, err)
	}

	// If aggregate changes has single event save it
	if len(events) == 1 {
		result, err := tx.Exec(
			ctx,
			saveEventQuery,
			events[0].GetAggregateID(),
			events[0].GetAggregateType(),
			events[0].GetEventType(),
			events[0].GetData(),
			events[0].GetVersion(),
			events[0].GetMetadata(),
		)
		if err != nil {
			tracing.TraceErr(span, err)
			p.log.Errorf("(SaveEvents) tx.Exec err: %v", err)
			return RollBackTx(ctx, tx, err)
		}

		if err := p.processEvents(ctx, events); err != nil {
			tracing.TraceErr(span, err)
			return RollBackTx(ctx, tx, err)
		}

		p.log.Debugf("(SaveEvents) result: {%s}, AggregateID: {%s}, AggregateVersion: {%v}", result.String(), events[0].GetAggregateID(), events[0].GetVersion())
		return tx.Commit(ctx)
	}

	batch := &pgx.Batch{}
	for _, event := range events {
		batch.Queue(
			saveEventQuery,
			event.GetAggregateID(),
			event.GetAggregateType(),
			event.GetEventType(),
			event.GetData(),
			event.GetVersion(),
			event.GetMetadata(),
		)
	}

	if err := tx.SendBatch(ctx, batch).Close(); err != nil {
		tracing.TraceErr(span, err)
		p.log.Errorf("(SaveEvents) tx.SendBatch err: %v", err)
		return RollBackTx(ctx, tx, err)
	}

	if err := p.processEvents(ctx, events); err != nil {
		tracing.TraceErr(span, err)
		return RollBackTx(ctx, tx, err)
	}

	return tx.Commit(ctx)
}

// LoadEvents load aggregate events by id
func (p *pgEventStore[T]) LoadEvents(ctx context.Context, aggregateID string) ([]Event, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pgEventStore.LoadEvents")
	defer span.Finish()

	rows, err := p.db.Query(ctx, getEventsQuery, aggregateID)
	if err != nil {
		tracing.TraceErr(span, err)
		p.log.Errorf("(LoadEvents) db.Query err: %v", err)
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	events := make([]Event, 0, eventsCapacity)

	for rows.Next() {
		var event Event
		if err := rows.Scan(
			&event.EventID,
			&event.AggregateID,
			&event.AggregateType,
			&event.EventType,
			&event.Data,
			&event.Version,
			&event.Timestamp,
			&event.Metadata,
		); err != nil {
			tracing.TraceErr(span, err)
			p.log.Errorf("(LoadEvents) rows.Next err: %v", err)
			return nil, errors.Wrap(err, "rows.Scan")
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		tracing.TraceErr(span, err)
		p.log.Errorf("(LoadEvents) rows.Err err: %v", err)
		return nil, errors.Wrap(err, "rows.Err")
	}

	return events, nil
}

// LoadEvents load aggregate events by id
func (p *pgEventStore[T]) loadEvents(ctx context.Context, aggregate Aggregate) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pgEventStore.loadEvents")
	defer span.Finish()
	span.LogFields(log.String("aggregate", aggregate.String()))

	rows, err := p.db.Query(ctx, getEventsQuery, aggregate.GetID())
	if err != nil {
		tracing.TraceErr(span, err)
		p.log.Errorf("(loadEvents) db.Query err: %v", err)
		return errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	for rows.Next() {
		var event Event
		if err := rows.Scan(
			&event.EventID,
			&event.AggregateID,
			&event.AggregateType,
			&event.EventType,
			&event.Data,
			&event.Version,
			&event.Timestamp,
			&event.Metadata,
		); err != nil {
			tracing.TraceErr(span, err)
			p.log.Errorf("(loadEvents) rows.Next err: %v", err)
			return errors.Wrap(err, "rows.Scan")
		}

		if err := aggregate.RaiseEvent(event); err != nil {
			tracing.TraceErr(span, err)
			p.log.Errorf("(loadEvents) aggregate.RaiseEvent err: %v", err)
			return errors.Wrap(err, "RaiseEvent")
		}
	}

	if err := rows.Err(); err != nil {
		tracing.TraceErr(span, err)
		p.log.Errorf("(loadEvents) rows.Err err: %v", err)
		return errors.Wrap(err, "rows.Err")
	}

	span.LogFields(log.String("aggregate with events", aggregate.String()))
	return nil
}

// Exists check for exists aggregate by id
func (p *pgEventStore[T]) Exists(ctx context.Context, aggregateID string) (bool, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pgEventStore.Exists")
	defer span.Finish()

	var id string
	if err := p.db.QueryRow(ctx, getEventQuery, aggregateID).Scan(&id); err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		tracing.TraceErr(span, err)
		p.log.Errorf("(Exists) db.QueryRow err: %v", err)
		return false, errors.Wrap(err, "db.QueryRow")
	}

	p.log.Debugf("(Exists Aggregate): id: {%s}", id)
	return true, nil
}

func (p *pgEventStore[T]) loadEventsByVersion(ctx context.Context, aggregateID string, versionFrom uint64) ([]Event, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pgEventStore.loadEventsByVersion")
	defer span.Finish()
	span.LogFields(log.String("aggregateID", aggregateID), log.Uint64("versionFrom", versionFrom))

	rows, err := p.db.Query(ctx, getEventsByVersionQuery, aggregateID, versionFrom)
	if err != nil {
		tracing.TraceErr(span, err)
		p.log.Errorf("(loadEventsByVersion) db.Query err: %v", err)
		return nil, errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	events := make([]Event, 0, p.cfg.SnapshotFrequency)

	for rows.Next() {
		var event Event
		if err := rows.Scan(
			&event.AggregateID,
			&event.AggregateType,
			&event.EventType,
			&event.Data,
			&event.Version,
			&event.Timestamp,
			&event.Metadata,
		); err != nil {
			tracing.TraceErr(span, err)
			p.log.Errorf("(loadEventsByVersion) rows.Next err: %v", err)
			return nil, errors.Wrap(err, "rows.Scan")
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		tracing.TraceErr(span, err)
		p.log.Errorf("(loadEventsByVersion) rows.Err err: %v", err)
		return nil, errors.Wrap(err, "rows.Err")
	}

	return events, nil
}

func (p *pgEventStore[T]) loadAggregateEventsByVersion(ctx context.Context, aggregate Aggregate) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pgEventStore.loadAggregateEventsByVersion")
	defer span.Finish()
	span.LogFields(log.String("aggregate", aggregate.String()))

	rows, err := p.db.Query(ctx, getEventsByVersionQuery, aggregate.GetID(), aggregate.GetVersion())
	if err != nil {
		tracing.TraceErr(span, err)
		p.log.Errorf("(loadAggregateEventsByVersion) db.Query err: %v", err)
		return errors.Wrap(err, "db.Query")
	}
	defer rows.Close()

	for rows.Next() {
		var event Event
		if err := rows.Scan(
			&event.EventID,
			&event.AggregateID,
			&event.AggregateType,
			&event.EventType,
			&event.Data,
			&event.Version,
			&event.Timestamp,
			&event.Metadata,
		); err != nil {
			tracing.TraceErr(span, err)
			p.log.Errorf("(loadAggregateEventsByVersion) rows.Scan err: %v", err)
			return errors.Wrap(err, "rows.Scan")
		}

		if err := aggregate.RaiseEvent(event); err != nil {
			tracing.TraceErr(span, err)
			p.log.Errorf("(loadAggregateEventsByVersion) aggregate.RaiseEvent err: %v", err)
			return errors.Wrap(err, "RaiseEvent")
		}

		p.log.Debugf("(loadAggregateEventsByVersion) event: {%s}", event.String())
	}

	if err := rows.Err(); err != nil {
		tracing.TraceErr(span, err)
		p.log.Errorf("(loadEventsByVersion) rows.Err err: %v", err)
		return errors.Wrap(err, "rows.Err")
	}

	span.LogFields(log.String("aggregate with events", aggregate.String()))
	return nil
}

func (p *pgEventStore[T]) loadEventsByVersionTx(ctx context.Context, tx pgx.Tx, aggregateID string, versionFrom int64) ([]Event, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pgEventStore.loadEventsByVersionTx")
	defer span.Finish()

	rows, err := tx.Query(ctx, getEventsByVersionQuery, aggregateID, versionFrom)
	if err != nil {
		tracing.TraceErr(span, err)
		p.log.Errorf("(loadEventsByVersionTx) tx.Query err: %v", err)
		return nil, errors.Wrap(err, "tx.Query")
	}
	defer rows.Close()

	events := make([]Event, 0, p.cfg.SnapshotFrequency)

	for rows.Next() {
		var event Event
		if err := rows.Scan(
			&event.EventID,
			&event.AggregateID,
			&event.AggregateType,
			&event.EventType,
			&event.Data,
			&event.Version,
			&event.Timestamp,
			&event.Metadata,
		); err != nil {
			tracing.TraceErr(span, err)
			p.log.Errorf("(loadEventsByVersionTx) rows.Next err: %v", err)
			return nil, errors.Wrap(err, "rows.Scan")
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		tracing.TraceErr(span, err)
		p.log.Errorf("(loadEventsByVersionTx) rows.Err err: %v", err)
		return nil, errors.Wrap(err, "rows.Err")
	}

	return events, nil
}

func (p *pgEventStore[T]) handleConcurrency(ctx context.Context, tx pgx.Tx, events []Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pgEventStore.handleConcurrency")
	defer span.Finish()

	result, err := tx.Exec(ctx, handleConcurrentWriteQuery, events[0].GetAggregateID())
	if err != nil {
		tracing.TraceErr(span, err)
		p.log.Errorf("(handleConcurrency) tx.Exec err: %v", err)
		return errors.Wrap(err, "tx.Exec")
	}

	p.log.Debugf("(handleConcurrency) result: {%s}", result.String())
	return nil
}

func (p *pgEventStore[T]) saveEventsTx(ctx context.Context, tx pgx.Tx, events []Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pgEventStore.saveEventsTx")
	defer span.Finish()

	if err := p.handleConcurrency(ctx, tx, events); err != nil {
		return err
	}

	if len(events) == 1 {
		result, err := tx.Exec(
			ctx,
			saveEventQuery,
			events[0].GetAggregateID(),
			events[0].GetAggregateType(),
			events[0].GetEventType(),
			events[0].GetData(),
			events[0].GetVersion(),
			events[0].GetMetadata(),
		)
		if err != nil {
			tracing.TraceErr(span, err)
			p.log.Errorf("(saveEventsTx) tx.Exec err: %v", err)
			return errors.Wrap(err, "tx.Exec")
		}

		p.log.Debugf("(saveEventsTx): {%s}, AggregateID: {%s}, AggregateVersion: {%v}", result.String(), events[0].GetAggregateID(), events[0].GetVersion())
		return nil
	}

	batch := &pgx.Batch{}
	for _, event := range events {
		batch.Queue(
			saveEventQuery,
			event.GetAggregateID(),
			event.GetAggregateType(),
			event.GetEventType(),
			event.GetData(),
			event.GetVersion(),
			event.GetMetadata(),
		)
	}

	if err := tx.SendBatch(ctx, batch).Close(); err != nil {
		tracing.TraceErr(span, err)
		p.log.Errorf("(saveEventsTx) tx.SendBatch err: %v", err)
		return errors.Wrap(err, "tx.SendBatch")
	}

	p.log.Debugf("(saveEventsTx): AggregateID: {%s}, AggregateVersion: {%v}, AggregateType: {%s}", events[0].GetAggregateID(), events[0].GetVersion(), events[0].GetAggregateType())
	return nil
}

func (p *pgEventStore[T]) saveSnapshotTx(ctx context.Context, tx pgx.Tx, aggregate Aggregate) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pgEventStore.saveSnapshotTx")
	defer span.Finish()

	snapshot, err := NewSnapshotFromAggregate(aggregate)
	if err != nil {
		tracing.TraceErr(span, err)
		p.log.Errorf("(saveSnapshotTx) NewSnapshotFromAggregate err: %v", err)
		return err
	}

	_, err = tx.Exec(ctx, saveSnapshotQuery, snapshot.ID, snapshot.Type, snapshot.State, snapshot.Version)
	if err != nil {
		tracing.TraceErr(span, err)
		p.log.Errorf("(saveSnapshotTx) tx.Exec err: %v", err)
		return errors.Wrap(err, "tx.Exec")
	}

	p.log.Debugf("(saveSnapshotTx) snapshot: {%s}", snapshot.String())
	return nil
}

func (p *pgEventStore[T]) processEvents(ctx context.Context, events []Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pgEventStore.processEvents")
	defer span.Finish()

	return p.eventBus.ProcessEvents(ctx, events)
}

func RollBackTx(ctx context.Context, tx pgx.Tx, err error) error {
	if err := tx.Rollback(ctx); err != nil {
		return errors.Wrap(err, "tx.Rollback")
	}
	return err
}
