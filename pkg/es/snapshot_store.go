package es

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
)

// SaveSnapshot save es.Aggregate snapshot
func (p *pgEventStore) SaveSnapshot(ctx context.Context, aggregate Aggregate) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pgEventStore.SaveSnapshot")
	defer span.Finish()
	span.LogFields(log.String("aggregate", aggregate.String()))

	snapshot, err := NewSnapshotFromAggregate(aggregate)
	if err != nil {
		return errors.Wrap(err, "NewSnapshotFromAggregate")
	}

	_, err = p.db.Exec(ctx, saveSnapshotQuery, snapshot.ID, snapshot.Type, snapshot.State, snapshot.Version)
	if err != nil {
		return errors.Wrap(err, "db.Exec")
	}

	p.log.Debugf("(SaveSnapshot) snapshot: %s", snapshot.String())
	span.LogFields(log.String("snapshot", snapshot.String()))
	return nil
}

// GetSnapshot load es.Aggregate snapshot
func (p *pgEventStore) GetSnapshot(ctx context.Context, id string) (*Snapshot, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "pgEventStore.GetSnapshot")
	defer span.Finish()
	span.LogFields(log.String("AggregateID", id))

	var snapshot Snapshot
	if err := p.db.QueryRow(ctx, getSnapshotQuery, id).Scan(&snapshot.ID, &snapshot.Type, &snapshot.State, &snapshot.Version); err != nil {
		return nil, errors.Wrap(err, "db.QueryRow")
	}

	p.log.Debugf("(GetSnapshot) snapshot: %s", snapshot.String())
	span.LogFields(log.String("snapshot", snapshot.String()))
	return &snapshot, nil
}
