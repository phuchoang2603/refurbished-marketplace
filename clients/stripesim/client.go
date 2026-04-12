package stripesim

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

type ChargeRequest struct {
	PaymentTransactionID string            `json:"payment_transaction_id"`
	IdempotencyKey       string            `json:"idempotency_key"`
	AmountCents          int64             `json:"amount_cents"`
	Currency             string            `json:"currency"`
	MerchantID           string            `json:"merchant_id"`
	BuyerID              string            `json:"buyer_id"`
	PaymentToken         string            `json:"payment_token"`
	Metadata             map[string]string `json:"metadata,omitempty"`
}

type ChargeResponse struct {
	GatewayTransactionID string `json:"gateway_transaction_id"`
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) CreateCharge(ctx context.Context, req ChargeRequest) (ChargeResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return ChargeResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/charges", bytes.NewReader(body))
	if err != nil {
		return ChargeResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Idempotency-Key", req.IdempotencyKey)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return ChargeResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		return ChargeResponse{}, fmt.Errorf("stripe sim: unexpected status %d", resp.StatusCode)
	}

	var out ChargeResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return ChargeResponse{}, err
	}
	return out, nil
}

