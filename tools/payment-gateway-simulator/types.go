package main

type pageData struct {
	OrderID          string
	PaymentSessionID string
	ReturnURL        string
	CancelURL        string
	CallbackURL      string
	Error            string
}

type callbackRequest struct {
	OrderID          string `json:"order_id"`
	PaymentSessionID string `json:"payment_session_id"`
	Status           string `json:"status"`
	FailureReason    string `json:"failure_reason"`
}
