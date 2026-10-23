package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type gatewayParty struct {
	ExternalID       string    `json:"external_id"`
	AccountCreatedAt time.Time `json:"account_created_at"`
}
type gatewayMoney struct {
	MinorUnits int64  `json:"minor_units"`
	Currency   string `json:"currency"`
}
type gatewayItem struct {
	ProductID string `json:"product_id"`
	Category  string `json:"category"`
	Condition string `json:"condition"`
	Quantity  int32  `json:"quantity"`
	UnitPrice int64  `json:"unit_price"`
}
type gatewayAddress struct {
	Name       string `json:"name"`
	Line1      string `json:"line1"`
	Line2      string `json:"line2"`
	City       string `json:"city"`
	Region     string `json:"region"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}
type gatewayRequest struct {
	OrderReference  string         `json:"order_reference"`
	Seller          gatewayParty   `json:"seller"`
	Buyer           gatewayParty   `json:"buyer"`
	Amount          gatewayMoney   `json:"amount"`
	Items           []gatewayItem  `json:"items"`
	ShippingAddress gatewayAddress `json:"shipping_address"`
	ReturnURL       string         `json:"return_url"`
}
type gatewayPayment struct {
	PaymentID      string    `json:"payment_id"`
	CheckoutURL    string    `json:"checkout_url"`
	ExpiresAt      time.Time `json:"expires_at"`
	Status         string    `json:"status"`
	OrderReference string    `json:"order_reference"`
}

var (
	errGatewayDefect      = errors.New("gateway rejected the frozen payment request")
	errGatewayUnavailable = errors.New("gateway unavailable")
)

type gatewayClient struct {
	endpoint string
	apiKey   string
	http     *http.Client
}

func newGatewayClient(config Config) *gatewayClient {
	return &gatewayClient{
		endpoint: strings.TrimRight(config.PaymentGatewayURL, "/"), apiKey: config.PaymentGatewayAPIKey,
		http: &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }},
	}
}

func (client *gatewayClient) request(ctx context.Context, method, path, key string, body []byte) (gatewayPayment, error) {
	if client.endpoint == "" {
		return gatewayPayment{}, errGatewayUnavailable
	}
	request, err := http.NewRequestWithContext(ctx, method, client.endpoint+path, bytes.NewReader(body))
	if err != nil {
		return gatewayPayment{}, errGatewayDefect
	}
	request.Header.Set("Authorization", "Bearer "+client.apiKey)
	request.Header.Set("Content-Type", "application/json")
	if key != "" {
		request.Header.Set("Idempotency-Key", key)
	}
	response, err := client.http.Do(request)
	if err != nil {
		return gatewayPayment{}, errGatewayUnavailable
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode >= 500 {
		return gatewayPayment{}, errGatewayUnavailable
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return gatewayPayment{}, errGatewayDefect
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, (64<<10)+1))
	if err != nil {
		return gatewayPayment{}, errGatewayUnavailable
	}
	var payment gatewayPayment
	if len(raw) > 64<<10 || json.Unmarshal(raw, &payment) != nil || payment.PaymentID == "" {
		return gatewayPayment{}, errGatewayDefect
	}
	return payment, nil
}

func (client *gatewayClient) create(ctx context.Context, orderID string, body []byte) (gatewayPayment, error) {
	payment, err := client.request(ctx, http.MethodPost, "/v1/payments", orderID, body)
	if err != nil {
		return gatewayPayment{}, err
	}
	parsed, err := url.Parse(payment.CheckoutURL)
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" || parsed.User != nil || !payment.ExpiresAt.After(time.Now()) || payment.Status != "OPEN" {
		return gatewayPayment{}, errGatewayDefect
	}
	return payment, nil
}

func (client *gatewayClient) get(ctx context.Context, paymentID string) (gatewayPayment, error) {
	payment, err := client.request(ctx, http.MethodGet, "/v1/payments/"+url.PathEscape(paymentID), "", nil)
	if err == nil && payment.PaymentID != paymentID {
		return gatewayPayment{}, errGatewayDefect
	}
	return payment, err
}
