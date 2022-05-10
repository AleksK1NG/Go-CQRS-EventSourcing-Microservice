package es

import (
	"encoding/json"
	"fmt"
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
	return fmt.Sprintf("ID: {%s}, Type: {%s}, StateSize: {%d}, Version: {%d},",
		s.ID,
		string(s.Type),
		len(s.State),
		s.Version,
	)
}

// NewSnapshotFromAggregate create new Snapshot from the Aggregate state.
func NewSnapshotFromAggregate(aggregate Aggregate) (*Snapshot, error) {

	aggregateBytes, err := json.Marshal(aggregate)
	if err != nil {
		return nil, errors.Wrap(err, "json.Marshal")
	}

	return &Snapshot{
		ID:      aggregate.GetID(),
		Type:    aggregate.GetType(),
		State:   aggregateBytes,
		Version: aggregate.GetVersion(),
	}, nil
}
