package main

import (
	"html/template"
	"net/http"
)

var payTemplate = template.Must(template.New("pay").Parse(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Hosted Payment Simulator</title>
    <style>
      body { font-family: sans-serif; background: #f8fafc; color: #0f172a; margin: 0; }
      main { max-width: 42rem; margin: 3rem auto; padding: 2rem; background: white; border: 1px solid #e2e8f0; border-radius: 1rem; }
      form { display: grid; gap: 1rem; }
      .grid { display: grid; gap: .5rem; }
      label { font-weight: 600; }
      input { padding: .75rem; border: 1px solid #cbd5e1; border-radius: .5rem; }
      .actions { display: flex; flex-wrap: wrap; gap: .75rem; }
      button { border: 0; border-radius: .5rem; padding: .75rem 1rem; cursor: pointer; }
      .primary { background: #2563eb; color: white; }
      .danger { background: #dc2626; color: white; }
      .muted { background: #e2e8f0; color: #0f172a; }
      .warning { background: #f59e0b; color: white; }
      .error { color: #b91c1c; }
      code { background: #f1f5f9; padding: .15rem .35rem; border-radius: .25rem; }
    </style>
  </head>
  <body>
    <main>
      <h1>Hosted Payment Simulator</h1>
      <p>This dev-only page simulates a hosted payment page and callback flow.</p>
      <div class="grid">
        <div><strong>Order:</strong> <code>{{ .OrderID }}</code></div>
        <div><strong>Session:</strong> <code>{{ .PaymentSessionID }}</code></div>
      </div>
      {{ if .Error }}<p class="error">{{ .Error }}</p>{{ end }}
      <form method="post" action="/pay">
        <input type="hidden" name="order_id" value="{{ .OrderID }}">
        <input type="hidden" name="payment_session_id" value="{{ .PaymentSessionID }}">
        <input type="hidden" name="return_url" value="{{ .ReturnURL }}">
        <input type="hidden" name="cancel_url" value="{{ .CancelURL }}">
        <input type="hidden" name="callback_url" value="{{ .CallbackURL }}">
        <div class="grid">
          <label for="card-number">Card number</label>
          <input id="card-number" name="card_number" value="4242 4242 4242 4242">
        </div>
        <div class="actions">
          <button class="primary" type="submit" name="action" value="succeeded">Pay successfully</button>
          <button class="danger" type="submit" name="action" value="failed">Fail payment</button>
          <button class="warning" type="submit" name="action" value="expired">Expire session</button>
          <button class="muted" type="submit" name="action" value="cancelled">Cancel and return</button>
        </div>
      </form>
    </main>
  </body>
</html>`))

func renderPayPage(w http.ResponseWriter, data pageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := payTemplate.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
