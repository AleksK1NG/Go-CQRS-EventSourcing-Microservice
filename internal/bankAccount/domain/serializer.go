package domain

import (
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/events"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es/serializer"
	"github.com/pkg/errors"
)

var (
	ErrInvalidEvent = errors.New("invalid event")
)

type eventSerializer struct {
}

func NewEventSerializer() *eventSerializer {
	return &eventSerializer{}
}

func (s *eventSerializer) SerializeEvent(aggregate es.Aggregate, event any) (es.Event, error) {
	eventBytes, err := serializer.Marshal(event)
	if err != nil {
		return es.Event{}, errors.Wrapf(err, "serializer.Marshal aggregateID: %s", aggregate.GetID())
	}

	switch evt := event.(type) {

	case *events.BankAccountCreatedEventV1:
		return es.NewEvent(aggregate, events.BankAccountCreatedEventType, eventBytes, evt.Metadata), nil

	case *events.BalanceDepositedEventV1:
		return es.NewEvent(aggregate, events.BalanceDepositedEventType, eventBytes, evt.Metadata), nil

	case *events.BalanceWithdrawnEventV1:
		return es.NewEvent(aggregate, events.BalanceWithdrawnEventType, eventBytes, evt.Metadata), nil

	case *events.EmailChangedEventV1:
		return es.NewEvent(aggregate, events.EmailChangedEventType, eventBytes, evt.Metadata), nil

	default:
		return es.Event{}, errors.Wrapf(ErrInvalidEvent, "aggregateID: %s, type: %T", aggregate.GetID(), event)
	}

}

func (s *eventSerializer) DeserializeEvent(event es.Event) (any, error) {
	switch event.GetEventType() {

	case events.BankAccountCreatedEventType:
		return deserializeEvent(event, new(events.BankAccountCreatedEventV1))

	case events.BalanceDepositedEventType:
		return deserializeEvent(event, new(events.BalanceDepositedEventV1))

	case events.BalanceWithdrawnEventType:
		return deserializeEvent(event, new(events.BalanceWithdrawnEventV1))

	case events.EmailChangedEventType:
		return deserializeEvent(event, new(events.EmailChangedEventV1))

	default:
		return nil, errors.Wrapf(ErrInvalidEvent, "type: %s", event.GetEventType())
	}
}

func deserializeEvent(event es.Event, targetEvent any) (any, error) {
	if err := event.GetJsonData(&targetEvent); err != nil {
		return nil, errors.Wrapf(err, "event.GetJsonData type: %s", event.GetEventType())
	}
	return targetEvent, nil
}
