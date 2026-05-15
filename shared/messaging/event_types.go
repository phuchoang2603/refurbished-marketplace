// Package messaging defines shared Kafka topic / event names and the Kafka consumer helpers.
package messaging

const (
	EventTypePaymentSucceeded = "payment.succeeded"
	EventTypePaymentFailed    = "payment.failed"
	EventTypeOrderCreated     = "orders.created"
)
