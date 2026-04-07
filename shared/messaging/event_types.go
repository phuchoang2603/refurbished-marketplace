package messaging

type EventType string

const (
	EventTypeUnspecified  EventType = "EVENT_TYPE_UNSPECIFIED"
	EventTypeOrderCreated EventType = "orders.created"
)
