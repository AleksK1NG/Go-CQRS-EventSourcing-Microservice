package es

import "context"

// EventsBus ProcessEvents method publish events to the app specific message broker.
type EventsBus interface {
	ProcessEvents(ctx context.Context, events []Event) error
}
