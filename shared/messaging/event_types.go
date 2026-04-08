package messaging

type EventType string

const (
	EventTypeUnspecified      EventType = "EVENT_TYPE_UNSPECIFIED"
	EventTypeOrderItemCreated EventType = "orders.item.created"
)
