package grpcserver

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"refurbished-marketplace/services/payment/internal/service"
	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"

	"github.com/google/uuid"
)

func mapPaymentTransaction(tx service.PaymentTransactionView) *paymentv1.PaymentTransaction {
	out := &paymentv1.PaymentTransaction{
		Id:             tx.ID,
		OrderId:        tx.OrderID,
		OrderItemId:    tx.OrderItemID,
		MerchantId:     tx.MerchantID,
		AmountCents:    tx.AmountCents,
		Currency:       tx.Currency,
		Status:         paymentStatusStringToProto(tx.Status),
		IdempotencyKey: tx.IdempotencyKey,
		CreatedAt:      timestamppb.New(tx.CreatedAt),
		UpdatedAt:      timestamppb.New(tx.UpdatedAt),
	}
	if tx.GatewayTransactionID != "" {
		out.GatewayTransactionId = tx.GatewayTransactionID
	}
	return out
}

func paymentStatusStringToProto(dbStatus string) paymentv1.PaymentTransactionStatus {
	switch strings.ToUpper(strings.TrimSpace(dbStatus)) {
	case service.PaymentTxStatusInitialized:
		return paymentv1.PaymentTransactionStatus_PAYMENT_TRANSACTION_STATUS_INITIALIZED
	case service.PaymentTxStatusSubmitted:
		return paymentv1.PaymentTransactionStatus_PAYMENT_TRANSACTION_STATUS_SUBMITTED
	case service.PaymentTxStatusSucceeded:
		return paymentv1.PaymentTransactionStatus_PAYMENT_TRANSACTION_STATUS_SUCCEEDED
	case service.PaymentTxStatusFailed:
		return paymentv1.PaymentTransactionStatus_PAYMENT_TRANSACTION_STATUS_FAILED
	default:
		return paymentv1.PaymentTransactionStatus_PAYMENT_TRANSACTION_STATUS_UNSPECIFIED
	}
}

func addressToJSON(a *paymentv1.Address) (json.RawMessage, error) {
	if a == nil {
		return []byte("{}"), nil
	}
	m := map[string]string{}
	if v := strings.TrimSpace(a.GetName()); v != "" {
		m["name"] = v
	}
	if v := strings.TrimSpace(a.GetLine1()); v != "" {
		m["line1"] = v
	}
	if v := strings.TrimSpace(a.GetLine2()); v != "" {
		m["line2"] = v
	}
	if v := strings.TrimSpace(a.GetCity()); v != "" {
		m["city"] = v
	}
	if v := strings.TrimSpace(a.GetRegion()); v != "" {
		m["region"] = v
	}
	if v := strings.TrimSpace(a.GetPostalCode()); v != "" {
		m["postal_code"] = v
	}
	if v := strings.TrimSpace(a.GetCountry()); v != "" {
		m["country"] = v
	}
	if len(m) == 0 {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}

func (s *Server) InitiatePayment(ctx context.Context, req *paymentv1.InitiatePaymentRequest) (*paymentv1.InitiatePaymentResponse, error) {
	orderID, err := uuid.Parse(req.GetOrderId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid order id")
	}
	buyerID, err := uuid.Parse(req.GetBuyerUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid buyer user id")
	}
	token := strings.TrimSpace(req.GetPaymentToken())
	if token == "" {
		return nil, status.Error(codes.InvalidArgument, "payment_token is required")
	}

	billing, err := addressToJSON(req.GetBillingAddress())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "billing_address: %v", err)
	}
	shipping, err := addressToJSON(req.GetShippingAddress())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "shipping_address: %v", err)
	}

	if err := s.svc.InitiatePayment(ctx, service.InitiatePaymentParams{
		OrderID:         orderID,
		BuyerUserID:     buyerID,
		PaymentToken:    token,
		Currency:        strings.TrimSpace(req.GetCurrency()),
		BillingAddress:  billing,
		ShippingAddress: shipping,
	}); err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &paymentv1.InitiatePaymentResponse{OrderId: orderID.String()}, nil
}

func (s *Server) HandleGatewayWebhook(ctx context.Context, req *paymentv1.HandleGatewayWebhookRequest) (*paymentv1.HandleGatewayWebhookResponse, error) {
	txID, err := uuid.Parse(req.GetPaymentTransactionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid payment transaction id")
	}

	var succeeded bool
	switch req.GetStatus() {
	case paymentv1.PaymentTransactionStatus_PAYMENT_TRANSACTION_STATUS_SUCCEEDED:
		succeeded = true
	case paymentv1.PaymentTransactionStatus_PAYMENT_TRANSACTION_STATUS_FAILED:
		succeeded = false
	case paymentv1.PaymentTransactionStatus_PAYMENT_TRANSACTION_STATUS_UNSPECIFIED,
		paymentv1.PaymentTransactionStatus_PAYMENT_TRANSACTION_STATUS_INITIALIZED,
		paymentv1.PaymentTransactionStatus_PAYMENT_TRANSACTION_STATUS_SUBMITTED:
		return nil, status.Error(codes.InvalidArgument, "status must be SUCCEEDED or FAILED")
	default:
		return nil, status.Error(codes.InvalidArgument, "unknown status")
	}

	if err := s.svc.ApplyGatewayWebhook(ctx, txID, req.GetGatewayTransactionId(), succeeded, req.GetFailureReason()); err != nil {
		if errors.Is(err, service.ErrTransactionNotFound) {
			return nil, status.Error(codes.NotFound, "payment transaction not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &paymentv1.HandleGatewayWebhookResponse{}, nil
}

func (s *Server) GetTransaction(ctx context.Context, req *paymentv1.GetTransactionRequest) (*paymentv1.PaymentTransaction, error) {
	id, err := uuid.Parse(req.GetPaymentTransactionId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid payment transaction id")
	}

	tx, err := s.svc.GetPaymentTransaction(ctx, id)
	if err != nil {
		if errors.Is(err, service.ErrTransactionNotFound) {
			return nil, status.Error(codes.NotFound, "payment transaction not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return mapPaymentTransaction(tx), nil
}
