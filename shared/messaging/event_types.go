// Package messaging defines shared Kafka topic / event names and the Kafka consumer helpers.
package messaging

const (
	EventTypeProductCreated             = "products.created"
	EventTypePaymentSucceeded           = "payment.succeeded"
	EventTypePaymentFailed              = "payment.failed"
	EventTypeOrderCreated               = "orders.created"
	EventTypeInventoryReserved          = "inventory.reserved"
	EventTypeInventoryReservationFailed = "inventory.reservation-failed"
)
