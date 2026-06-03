package main

import (
	"net/http"
	"strings"
)

func newServer() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/pay", handlePay)
	return mux
}

func handlePay(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		renderPayPage(w, pageData{
			OrderID:          r.URL.Query().Get("order_id"),
			PaymentSessionID: r.URL.Query().Get("payment_session_id"),
			ReturnURL:        r.URL.Query().Get("return_url"),
			CancelURL:        r.URL.Query().Get("cancel_url"),
			CallbackURL:      r.URL.Query().Get("callback_url"),
		})
	case http.MethodPost:
		handlePaySubmit(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handlePaySubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		renderPayPage(w, pageData{Error: "invalid form submission"})
		return
	}

	action := strings.TrimSpace(r.FormValue("action"))
	data := pageData{
		OrderID:          r.FormValue("order_id"),
		PaymentSessionID: r.FormValue("payment_session_id"),
		ReturnURL:        r.FormValue("return_url"),
		CancelURL:        r.FormValue("cancel_url"),
		CallbackURL:      r.FormValue("callback_url"),
	}

	if err := postCallback(r.Context(), data.CallbackURL, callbackRequest{
		OrderID:          data.OrderID,
		PaymentSessionID: data.PaymentSessionID,
		Status:           strings.ToUpper(action),
		FailureReason:    failureReasonForAction(action),
	}); err != nil {
		data.Error = err.Error()
		renderPayPage(w, data)
		return
	}

	if action == "cancelled" {
		http.Redirect(w, r, data.CancelURL, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, data.ReturnURL, http.StatusSeeOther)
}

func failureReasonForAction(action string) string {
	switch action {
	case "failed":
		return "Card declined"
	case "expired":
		return "Hosted payment session expired"
	case "cancelled":
		return "Buyer cancelled hosted payment"
	default:
		return ""
	}
}
