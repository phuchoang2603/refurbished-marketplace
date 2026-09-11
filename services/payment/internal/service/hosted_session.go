package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"

	"github.com/phuchoang2603/refurbished-marketplace/services/payment/internal/database"
	"github.com/phuchoang2603/refurbished-marketplace/shared/err/dberr"

	"github.com/google/uuid"
)

type PaymentTransactionView struct {
	ID                   string
	OrderID              string
	MerchantID           string
	AmountCents          int64
	Currency             string
	Status               string
	IdempotencyKey       string
	GatewayTransactionID string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type HostedPaymentSessionView struct {
	OrderID          string
	PaymentSessionID string
	Currency         string
	Status           string
	ReturnURL        string
	FailureReason    string
	ExpiresAt        time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type CreateHostedPaymentSessionParams struct {
	OrderID         uuid.UUID
	BuyerUserID     uuid.UUID
	MerchantID      uuid.UUID
	TotalCents      int64
	Currency        string
	ShippingAddress json.RawMessage
	LineItems       json.RawMessage
	Buyer           json.RawMessage
	Merchant        json.RawMessage
	ReturnURL       string
}

func (s *Service) CreateHostedPaymentSession(ctx context.Context, p CreateHostedPaymentSessionParams) (HostedPaymentSessionView, error) {
	intent, err := loadPaymentIntentByOrderID(ctx, s.queries, p.OrderID)
	if err == nil {
		if hostedPaymentSessionIsTerminal(intent.Status) {
			return HostedPaymentSessionView{}, ErrSessionTerminal
		}
		return mapDBHostedPaymentSessionView(intent), nil
	}
	if !errors.Is(err, ErrIntentNotFound) {
		return HostedPaymentSessionView{}, err
	}

	if p.BuyerUserID == uuid.Nil || p.MerchantID == uuid.Nil {
		return HostedPaymentSessionView{}, ErrInvalidSessionFacts
	}
	if p.TotalCents <= 0 {
		return HostedPaymentSessionView{}, ErrInvalidSessionFacts
	}
	p.Currency = defaultPaymentCurrency(p.Currency)
	if !usableShippingAddress(p.ShippingAddress) {
		return HostedPaymentSessionView{}, ErrInvalidSessionFacts
	}
	if len(p.LineItems) == 0 {
		p.LineItems = json.RawMessage(`[]`)
	}
	if len(p.Buyer) == 0 {
		buyer, err := partySnapshotJSON(p.BuyerUserID.String(), "")
		if err != nil {
			return HostedPaymentSessionView{}, err
		}
		p.Buyer = buyer
	}
	if len(p.Merchant) == 0 {
		merchant, err := partySnapshotJSON(p.MerchantID.String(), "")
		if err != nil {
			return HostedPaymentSessionView{}, err
		}
		p.Merchant = merchant
	}

	expiresAt := time.Now().UTC().Add(30 * time.Minute)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return HostedPaymentSessionView{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	q := s.queries.WithTx(tx)

	created, err := q.CreateHostedPaymentSession(ctx, database.CreateHostedPaymentSessionParams{
		OrderID:          p.OrderID,
		BuyerUserID:      p.BuyerUserID,
		Currency:         p.Currency,
		ShippingAddress:  p.ShippingAddress,
		Status:           HostedPaymentSessionStatusPending,
		PaymentSessionID: dberr.OptionalNullString(uuid.NewString()),
		ReturnUrl:        p.ReturnURL,
		ExpiresAt:        dberr.OptionalNullTime(expiresAt),
		LineItems:        p.LineItems,
		Buyer:            p.Buyer,
		Merchant:         p.Merchant,
	})
	if err != nil {
		return HostedPaymentSessionView{}, err
	}
	if _, err := q.CreatePaymentTransaction(ctx, database.CreatePaymentTransactionParams{
		ID:             uuid.New(),
		OrderID:        p.OrderID,
		MerchantID:     p.MerchantID,
		AmountCents:    p.TotalCents,
		Currency:       p.Currency,
		Status:         PaymentTxStatusInitialized,
		IdempotencyKey: "order:" + p.OrderID.String(),
	}); err != nil {
		return HostedPaymentSessionView{}, err
	}
	if err := tx.Commit(); err != nil {
		return HostedPaymentSessionView{}, err
	}

	view := mapDBHostedPaymentSessionView(created)
	sharedlog.InfoContext(
		ctx, "hosted payment session created",
		sharedlog.KeyOrderID, view.OrderID,
		sharedlog.KeyBuyerUserID, p.BuyerUserID.String(),
		sharedlog.KeyMerchantID, p.MerchantID.String(),
		"payment_session_id", view.PaymentSessionID,
		"currency", view.Currency,
		"total_cents", p.TotalCents,
	)
	return view, nil
}

func (s *Service) GetHostedPaymentSessionByOrder(ctx context.Context, orderID uuid.UUID) (HostedPaymentSessionView, error) {
	row, err := loadPaymentIntentByOrderID(ctx, s.queries, orderID)
	if err != nil {
		return HostedPaymentSessionView{}, err
	}
	return mapDBHostedPaymentSessionView(row), nil
}
