package events

import "github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/es"

const (
	EmailChangedEventType es.EventType = "EMAIL_CHANGED_EVENT_V1"
)

type EmailChangedEventV1 struct {
	Email    string `json:"email"`
	Metadata []byte `json:"-"`
}

func NewEmailChangedEventV1(email string) *EmailChangedEventV1 {
	emailChangedEvent := EmailChangedEventV1{Email: email}
	return &emailChangedEvent
}
