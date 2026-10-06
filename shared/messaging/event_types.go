// Package messaging defines shared Kafka topic / event names and the Kafka consumer helpers.
package messaging

const (
	EventTypeProductCreated             = "products.created"
	EventTypeOrderCreated               = "orders.created"
	EventTypeInventoryReserved          = "inventory.reserved"
	EventTypeInventoryReservationFailed = "inventory.reservation-failed"
	EventTypeCheckoutCreateOrder        = "checkout.order-create.requested.v1"
	EventTypeCheckoutOrderCreated       = "checkout.order-created.v1"
	EventTypeCheckoutOrderRejected      = "checkout.order-rejected.v1"
	EventTypeCheckoutReserveStock       = "checkout.stock-reserve.requested.v1"
	EventTypeCheckoutStockReserved      = "checkout.stock-reserved.v1"
	EventTypeCheckoutStockRejected      = "checkout.stock-rejected.v1"
	EventTypeCheckoutCancelStock        = "checkout.stock-cancel.requested.v1"
	EventTypeCheckoutStockReleased      = "checkout.stock-released.v1"
	EventTypeCheckoutCreateSession      = "checkout.payment-session-create.requested.v1"
	EventTypeCheckoutSessionReady       = "checkout.payment-session-ready.v1"
	EventTypeCheckoutSessionRejected    = "checkout.payment-session-rejected.v1"
	EventTypeCheckoutCancelSession      = "checkout.payment-session-cancel.requested.v1"
	EventTypeCheckoutPaymentCancelled   = "checkout.payment-cancelled.v1"
	EventTypeCheckoutPaymentSucceeded   = "checkout.payment-succeeded.v1"
	EventTypeCheckoutPaymentFailed      = "checkout.payment-failed.v1"
	EventTypeCheckoutPaymentUncertain   = "checkout.payment-uncertain.v1"
	EventTypeCheckoutCommitStock        = "checkout.stock-commit.requested.v1"
	EventTypeCheckoutStockCommitted     = "checkout.stock-committed.v1"
	EventTypeCheckoutFinalizeOrder      = "checkout.order-finalize.requested.v1"
	EventTypeCheckoutOrderFinalized     = "checkout.order-finalized.v1"
)
