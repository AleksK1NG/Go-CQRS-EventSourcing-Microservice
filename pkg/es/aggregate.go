package es

import (
	"fmt"
	"github.com/pkg/errors"
)

const (
	changesEventsCap = 10
	startVersion     = 0
)

type When interface {
	When(event Event) error
}

type when func(event Event) error

// Apply process Aggregate Event
type Apply interface {
	Apply(event Event) error
}

// Load create Aggregate state from Event's.
type Load interface {
	Load(events []Event) error
}

// RaiseEvent process applied Aggregate Event from event store
type RaiseEvent interface {
	RaiseEvent(event Event) error
}

type Aggregate interface {
	When
	AggregateRoot
	RaiseEvent
}

// AggregateRoot contains all methods of AggregateBase
type AggregateRoot interface {
	GetID() string
	SetID(id string) *AggregateBase
	GetType() AggregateType
	SetType(aggregateType AggregateType)
	GetChanges() []Event
	ClearChanges()
	GetVersion() uint64
	ToSnapshot()
	String() string
	Load
	Apply
	RaiseEvent
}

// AggregateType type of the Aggregate
type AggregateType string

// AggregateBase base aggregate contains all main necessary fields
type AggregateBase struct {
	ID      string
	Version uint64
	Changes []Event
	Type    AggregateType
	when    when
}

// NewAggregateBase AggregateBase constructor, contains all main fields and methods,
// main aggregate must realize When interface and pass as argument to constructor
// Example of recommended aggregate constructor method:
//
// func NewOrderAggregate() *OrderAggregate {
//	orderAggregate := &OrderAggregate{
//		Order: models.NewOrder(),
//	}
//	base := es.NewAggregateBase(orderAggregate.When)
//	base.SetType(OrderAggregateType)
//	orderAggregate.AggregateBase = base
//	return orderAggregate
//}
func NewAggregateBase(when when) *AggregateBase {
	if when == nil {
		return nil
	}

	return &AggregateBase{
		Version: startVersion,
		Changes: make([]Event, 0, changesEventsCap),
		when:    when,
	}
}

// SetID set AggregateBase ID
func (a *AggregateBase) SetID(id string) *AggregateBase {
	a.ID = id
	return a
}

// GetID get AggregateBase ID
func (a *AggregateBase) GetID() string {
	return a.ID
}

// SetType set AggregateBase AggregateType
func (a *AggregateBase) SetType(aggregateType AggregateType) {
	a.Type = aggregateType
}

// GetType get AggregateBase AggregateType
func (a *AggregateBase) GetType() AggregateType {
	return a.Type
}

// GetVersion get AggregateBase version
func (a *AggregateBase) GetVersion() uint64 {
	return a.Version
}

// ClearChanges clear AggregateBase uncommitted Event's
func (a *AggregateBase) ClearChanges() {
	a.Changes = make([]Event, 0, changesEventsCap)
}

// GetChanges get AggregateBase uncommitted Event's
func (a *AggregateBase) GetChanges() []Event {
	return a.Changes
}

// Load add existing events from event store to aggregate using When interface method
func (a *AggregateBase) Load(events []Event) error {

	for _, evt := range events {
		if evt.GetAggregateID() != a.GetID() || evt.GetAggregateType() != a.GetType() {
			return ErrInvalidAggregate
		}

		if err := a.when(evt); err != nil {
			return err
		}

		a.Version++
	}

	return nil
}

// Apply push event to aggregate uncommitted events using When method
func (a *AggregateBase) Apply(event Event) error {
	if err := a.validateEvent(event); err != nil {
		return err
	}

	event.SetAggregateType(a.GetType())

	if err := a.when(event); err != nil {
		return err
	}

	a.Version++
	event.SetVersion(a.GetVersion())
	a.Changes = append(a.Changes, event)
	return nil
}

// RaiseEvent push event to aggregate applied events using When method, used for load directly from eventstore
func (a *AggregateBase) RaiseEvent(event Event) error {
	if err := a.validateEvent(event); err != nil {
		return err
	}

	event.SetAggregateType(a.GetType())

	if err := a.when(event); err != nil {
		return err
	}

	a.Version++
	return a.validateVersion(event)
}

// ToSnapshot prepare AggregateBase for saving Snapshot.
func (a *AggregateBase) ToSnapshot() {
	a.ClearChanges()
}

func (a *AggregateBase) String() string {
	return fmt.Sprintf("(Aggregate) ID: {%s}, Type: {%s}, Version: {%v}, Changes: {%d}",
		a.GetID(),
		string(a.GetType()),
		a.GetVersion(),
		len(a.GetChanges()),
	)
}

func (a *AggregateBase) validateVersion(event Event) error {
	if a.GetVersion() != event.GetVersion() {
		return errors.New("invalid event version")
	}
	return nil
}

func (a *AggregateBase) validateEvent(event Event) error {
	if event.GetAggregateID() != a.GetID() {
		return ErrInvalidAggregateID
	}
	if event.GetAggregateType() != a.GetType() {
		return ErrInvalidEventType
	}
	return nil
}
