package es

// Serializer events must be serializable, EventStore requires to provide a serializer instance, which implements the Serializer interface.
type Serializer interface {
	SerializeEvent(aggregate Aggregate, event any) (Event, error)
	DeserializeEvent(event Event) (any, error)
}
