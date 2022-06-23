package domain

import (
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/events"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es/serializer"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"time"
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
		return es.Event{
			EventID:       uuid.NewV4().String(),
			AggregateID:   aggregate.GetID(),
			EventType:     events.BankAccountCreatedEventType,
			AggregateType: aggregate.GetType(),
			Version:       aggregate.GetVersion(),
			Data:          eventBytes,
			Metadata:      evt.Metadata,
			Timestamp:     time.Now().UTC(),
		}, nil
	case *events.BalanceDepositedEventV1:
		return es.Event{
			EventID:       uuid.NewV4().String(),
			AggregateID:   aggregate.GetID(),
			EventType:     events.BalanceDepositedEventType,
			AggregateType: aggregate.GetType(),
			Version:       aggregate.GetVersion(),
			Data:          eventBytes,
			Metadata:      evt.Metadata,
			Timestamp:     time.Now().UTC(),
		}, nil
	case *events.EmailChangedEventV1:
		return es.Event{
			EventID:       uuid.NewV4().String(),
			AggregateID:   aggregate.GetID(),
			EventType:     events.EmailChangedEventType,
			AggregateType: aggregate.GetType(),
			Version:       aggregate.GetVersion(),
			Data:          eventBytes,
			Metadata:      evt.Metadata,
			Timestamp:     time.Now().UTC(),
		}, nil
	default:
		return es.Event{}, errors.Wrapf(ErrInvalidEvent, "aggregateID: %s", aggregate.GetID())
	}

}

func (s *eventSerializer) DeserializeEvent(event es.Event) (any, error) {
	switch event.GetEventType() {
	case events.BankAccountCreatedEventType:
		var evt events.BankAccountCreatedEventV1
		if err := event.GetJsonData(&evt); err != nil {
			return nil, errors.Wrapf(err, "event.GetJsonData type: %s", event.GetEventType())
		}
		return &evt, nil
	case events.BalanceDepositedEventType:
		var evt events.BalanceDepositedEventV1
		if err := event.GetJsonData(&evt); err != nil {
			return nil, errors.Wrapf(err, "event.GetJsonData type: %s", event.GetEventType())
		}
		return &evt, nil
	case events.EmailChangedEventType:
		var evt events.EmailChangedEventV1
		if err := event.GetJsonData(&evt); err != nil {
			return nil, errors.Wrapf(err, "event.GetJsonData type: %s", event.GetEventType())
		}
		return &evt, nil
	default:
		return nil, errors.Wrapf(ErrInvalidEvent, "type: %s", event.GetEventType())
	}
}
