// Package messaging defines shared Kafka topic / event names and the Kafka consumer helpers.
package messaging

const (
	EventTypePaymentItemSucceeded = "payment.item.succeeded"
	EventTypePaymentItemFailed    = "payment.item.failed"
	EventTypeOrderItemCreated     = "orders.item.created"
)
