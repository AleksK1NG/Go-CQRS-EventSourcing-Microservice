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

type pgEventStore struct {
	log        logger.Logger
	cfg        Config
	db         *pgxpool.Pool
	eventBus   EventsBus
	serializer Serializer
}

func NewPgEventStore(log logger.Logger, cfg Config, db *pgxpool.Pool, eventBus EventsBus, serializer Serializer) *pgEventStore {
	return &pgEventStore{log: log, cfg: cfg, db: db, eventBus: eventBus, serializer: serializer}
}

// SaveEvents save aggregate uncommitted events as one batch and process with event bus using transaction
func (p *pgEventStore) SaveEvents(ctx context.Context, events []Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pgEventStore.SaveEvents")
	defer span.Finish()

	tx, err := p.db.Begin(ctx)
	if err != nil {
		p.log.Errorf("(SaveEvents) db.Begin err: %v", err)
		return tracing.TraceWithErr(span, errors.Wrap(err, "db.Begin"))
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
			p.log.Errorf("(SaveEvents) tx.Exec err: %v", tracing.TraceWithErr(span, err))
			return RollBackTx(ctx, tx, err)
		}

		if err := p.processEvents(ctx, events); err != nil {
			return RollBackTx(ctx, tx, tracing.TraceWithErr(span, err))
		}

		p.log.Debugf("(SaveEvents) result: %s, AggregateID: %s, AggregateVersion: %v", result.String(), events[0].GetAggregateID(), events[0].GetVersion())
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
		p.log.Errorf("(SaveEvents) tx.SendBatch err: %v", tracing.TraceWithErr(span, err))
		return RollBackTx(ctx, tx, err)
	}

	if err := p.processEvents(ctx, events); err != nil {
		return RollBackTx(ctx, tx, tracing.TraceWithErr(span, err))
	}

	return tx.Commit(ctx)
}

// LoadEvents load aggregate events by id
func (p *pgEventStore) LoadEvents(ctx context.Context, aggregateID string) ([]Event, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pgEventStore.LoadEvents")
	defer span.Finish()

	rows, err := p.db.Query(ctx, getEventsQuery, aggregateID)
	if err != nil {
		p.log.Errorf("(LoadEvents) db.Query err: %v", tracing.TraceWithErr(span, err))
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
			p.log.Errorf("(LoadEvents) rows.Next err: %v", tracing.TraceWithErr(span, err))
			return nil, errors.Wrap(err, "rows.Scan")
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		p.log.Errorf("(LoadEvents) rows.Err err: %v", err)
		return nil, tracing.TraceWithErr(span, errors.Wrap(err, "rows.Err"))
	}

	return events, nil
}

// LoadEvents load aggregate events by id
func (p *pgEventStore) loadEvents(ctx context.Context, aggregate Aggregate) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pgEventStore.loadEvents")
	defer span.Finish()
	span.LogFields(log.String("aggregate", aggregate.String()))

	rows, err := p.db.Query(ctx, getEventsQuery, aggregate.GetID())
	if err != nil {
		p.log.Errorf("(loadEvents) db.Query err: %v", err)
		return tracing.TraceWithErr(span, errors.Wrap(err, "db.Query"))
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
			p.log.Errorf("(loadEvents) rows.Next err: %v", err)
			return tracing.TraceWithErr(span, errors.Wrap(err, "rows.Scan"))
		}

		deserializedEvent, err := p.serializer.DeserializeEvent(event)
		if err != nil {
			p.log.Errorf("(loadEvents) serializer.DeserializeEvent err: %v", err)
			return tracing.TraceWithErr(span, errors.Wrap(err, "serializer.DeserializeEvent"))
		}

		if err := aggregate.RaiseEvent(deserializedEvent); err != nil {
			p.log.Errorf("(loadEvents) aggregate.RaiseEvent err: %v", err)
			return tracing.TraceWithErr(span, errors.Wrap(err, "RaiseEvent"))
		}
	}

	if err := rows.Err(); err != nil {
		p.log.Errorf("(loadEvents) rows.Err err: %v", err)
		return tracing.TraceWithErr(span, errors.Wrap(err, "rows.Err"))
	}

	span.LogFields(log.String("aggregate with events", aggregate.String()))
	return nil
}

// Exists check for exists aggregate by id
func (p *pgEventStore) Exists(ctx context.Context, aggregateID string) (bool, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pgEventStore.Exists")
	defer span.Finish()

	var id string
	if err := p.db.QueryRow(ctx, getEventQuery, aggregateID).Scan(&id); err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		p.log.Errorf("(Exists) db.QueryRow err: %v", err)
		return false, tracing.TraceWithErr(span, errors.Wrap(err, "db.QueryRow"))
	}

	p.log.Debugf("(Exists Aggregate): id: %s", id)
	return true, nil
}

func (p *pgEventStore) loadEventsByVersion(ctx context.Context, aggregateID string, versionFrom uint64) ([]Event, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pgEventStore.loadEventsByVersion")
	defer span.Finish()
	span.LogFields(log.String("aggregateID", aggregateID), log.Uint64("versionFrom", versionFrom))

	rows, err := p.db.Query(ctx, getEventsByVersionQuery, aggregateID, versionFrom)
	if err != nil {
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
			p.log.Errorf("(loadEventsByVersion) rows.Next err: %v", err)
			return nil, errors.Wrap(err, "rows.Scan")
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		p.log.Errorf("(loadEventsByVersion) rows.Err err: %v", err)
		return nil, tracing.TraceWithErr(span, errors.Wrap(err, "rows.Err"))
	}

	return events, nil
}

func (p *pgEventStore) loadAggregateEventsByVersion(ctx context.Context, aggregate Aggregate) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pgEventStore.loadAggregateEventsByVersion")
	defer span.Finish()
	span.LogFields(log.String("aggregate", aggregate.String()))

	rows, err := p.db.Query(ctx, getEventsByVersionQuery, aggregate.GetID(), aggregate.GetVersion())
	if err != nil {
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
			p.log.Errorf("(loadAggregateEventsByVersion) rows.Scan err: %v", err)
			return tracing.TraceWithErr(span, errors.Wrap(err, "rows.Scan"))
		}

		deserializedEvent, err := p.serializer.DeserializeEvent(event)
		if err != nil {
			p.log.Errorf("(loadAggregateEventsByVersion) serializer.DeserializeEvent err: %v", err)
			return tracing.TraceWithErr(span, errors.Wrap(err, "serializer.DeserializeEvent"))
		}

		if err := aggregate.RaiseEvent(deserializedEvent); err != nil {
			p.log.Errorf("(loadAggregateEventsByVersion) aggregate.RaiseEvent err: %v", err)
			return tracing.TraceWithErr(span, errors.Wrap(err, "RaiseEvent"))
		}

		p.log.Debugf("(loadAggregateEventsByVersion) event: %s", event.String())
	}

	if err := rows.Err(); err != nil {
		p.log.Errorf("(loadEventsByVersion) rows.Err err: %v", err)
		return tracing.TraceWithErr(span, errors.Wrap(err, "rows.Err"))
	}

	span.LogFields(log.String("aggregate with events", aggregate.String()))
	return nil
}

func (p *pgEventStore) loadEventsByVersionTx(ctx context.Context, tx pgx.Tx, aggregateID string, versionFrom int64) ([]Event, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pgEventStore.loadEventsByVersionTx")
	defer span.Finish()

	rows, err := tx.Query(ctx, getEventsByVersionQuery, aggregateID, versionFrom)
	if err != nil {
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
			p.log.Errorf("(loadEventsByVersionTx) rows.Next err: %v", err)
			return nil, tracing.TraceWithErr(span, errors.Wrap(err, "rows.Scan"))
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		p.log.Errorf("(loadEventsByVersionTx) rows.Err err: %v", err)
		return nil, tracing.TraceWithErr(span, errors.Wrap(err, "rows.Err"))
	}

	return events, nil
}

func (p *pgEventStore) handleConcurrency(ctx context.Context, tx pgx.Tx, events []Event) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pgEventStore.handleConcurrency")
	defer span.Finish()

	result, err := tx.Exec(ctx, handleConcurrentWriteQuery, events[0].GetAggregateID())
	if err != nil {
		p.log.Errorf("(handleConcurrency) tx.Exec err: %v", err)
		return errors.Wrap(err, "tx.Exec")
	}

	p.log.Debugf("(handleConcurrency) result: {%s}", result.String())
	return nil
}

func (p *pgEventStore) saveEventsTx(ctx context.Context, tx pgx.Tx, events []Event) error {
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
			p.log.Errorf("(saveEventsTx) tx.Exec err: %v", err)
			return tracing.TraceWithErr(span, errors.Wrap(err, "tx.Exec"))
		}

		p.log.Debugf("(saveEventsTx): %s, AggregateID: %s, AggregateVersion: %v", result.String(), events[0].GetAggregateID(), events[0].GetVersion())
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
		p.log.Errorf("(saveEventsTx) tx.SendBatch err: %v", err)
		return tracing.TraceWithErr(span, errors.Wrap(err, "tx.SendBatch"))
	}

	p.log.Debugf("(saveEventsTx) AggregateID: %s, AggregateVersion: %v, AggregateType: %s", events[0].GetAggregateID(), events[0].GetVersion(), events[0].GetAggregateType())
	return nil
}

func (p *pgEventStore) saveSnapshotTx(ctx context.Context, tx pgx.Tx, aggregate Aggregate) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pgEventStore.saveSnapshotTx")
	defer span.Finish()

	snapshot, err := NewSnapshotFromAggregate(aggregate)
	if err != nil {
		p.log.Errorf("(saveSnapshotTx) NewSnapshotFromAggregate err: %v", err)
		return err
	}

	_, err = tx.Exec(ctx, saveSnapshotQuery, snapshot.ID, snapshot.Type, snapshot.State, snapshot.Version)
	if err != nil {
		p.log.Errorf("(saveSnapshotTx) tx.Exec err: %v", err)
		return tracing.TraceWithErr(span, errors.Wrap(err, "tx.Exec"))
	}

	p.log.Debugf("(saveSnapshotTx) snapshot: %s", snapshot.String())
	return nil
}

func (p *pgEventStore) processEvents(ctx context.Context, events []Event) error {
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
