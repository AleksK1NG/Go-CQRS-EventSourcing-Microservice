package events

import "github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es"

const (
	EmailChangedEventType es.EventType = "EMAIL_CHANGED_EVENT_V1"
)

type EmailChangedEventV1 struct {
	Email string `json:"email"`
}

func NewEmailChangedEventV1(aggregate es.Aggregate, email string) (es.Event, error) {
	emailChangedEvent := EmailChangedEventV1{Email: email}
	return es.NewEvent(aggregate, EmailChangedEventType, &emailChangedEvent)
}
