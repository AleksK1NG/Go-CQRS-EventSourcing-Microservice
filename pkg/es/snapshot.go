package es

import (
	"fmt"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es/serializer"
	"github.com/pkg/errors"
)

// Snapshot Event Sourcing Snapshotting is an optimisation that reduces time spent on reading event from an event store.
type Snapshot struct {
	ID      string        `json:"id"`
	Type    AggregateType `json:"type"`
	State   []byte        `json:"state"`
	Version uint64        `json:"version"`
}

func (s *Snapshot) String() string {
	return fmt.Sprintf("AggregateID: %s, Type: %s, StateSize: %d, Version: %d",
		s.ID,
		string(s.Type),
		len(s.State),
		s.Version,
	)
}

// NewSnapshotFromAggregate create new Snapshot from the Aggregate state.
func NewSnapshotFromAggregate(aggregate Aggregate) (*Snapshot, error) {

	aggregateBytes, err := serializer.Marshal(aggregate)
	if err != nil {
		return nil, errors.Wrapf(err, "serializer.Marshal aggregateID: %s", aggregate.GetID())
	}

	return &Snapshot{
		ID:      aggregate.GetID(),
		Type:    aggregate.GetType(),
		State:   aggregateBytes,
		Version: aggregate.GetVersion(),
	}, nil
}
