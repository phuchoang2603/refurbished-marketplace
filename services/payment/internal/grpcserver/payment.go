package grpcserver

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/phuchoang2603/refurbished-marketplace/services/payment/internal/service"
	"github.com/phuchoang2603/refurbished-marketplace/shared/err/grpcerr"
	paymentv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/payment/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapHostedPaymentSession(session service.HostedPaymentSessionView) *paymentv1.HostedPaymentSession {
	return &paymentv1.HostedPaymentSession{
		Status:        hostedPaymentStatusStringToProto(session.Status),
		FailureReason: session.FailureReason,
	}
}

func hostedPaymentStatusStringToProto(dbStatus string) paymentv1.HostedPaymentSessionStatus {
	switch strings.ToUpper(strings.TrimSpace(dbStatus)) {
	case service.HostedPaymentSessionStatusPending:
		return paymentv1.HostedPaymentSessionStatus_HOSTED_PAYMENT_SESSION_STATUS_PENDING
	case service.HostedPaymentSessionStatusSucceeded:
		return paymentv1.HostedPaymentSessionStatus_HOSTED_PAYMENT_SESSION_STATUS_SUCCEEDED
	case service.HostedPaymentSessionStatusFailed:
		return paymentv1.HostedPaymentSessionStatus_HOSTED_PAYMENT_SESSION_STATUS_FAILED
	case service.HostedPaymentSessionStatusExpired:
		return paymentv1.HostedPaymentSessionStatus_HOSTED_PAYMENT_SESSION_STATUS_EXPIRED
	default:
		return paymentv1.HostedPaymentSessionStatus_HOSTED_PAYMENT_SESSION_STATUS_UNSPECIFIED
	}
}

func lineItemsToJSON(items []*paymentv1.HostedPaymentLineItem) (json.RawMessage, error) {
	if len(items) == 0 {
		return []byte("[]"), nil
	}
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		row := map[string]any{
			"product_id":       strings.TrimSpace(item.GetProductId()),
			"quantity":         item.GetQuantity(),
			"unit_price_cents": item.GetUnitPriceCents(),
		}
		if name := strings.TrimSpace(item.GetName()); name != "" {
			row["name"] = name
		}
		out = append(out, row)
	}
	return json.Marshal(out)
}

func partyToJSON(p *paymentv1.PartySnapshot) (json.RawMessage, error) {
	if p == nil {
		return []byte("{}"), nil
	}
	m := map[string]string{}
	if v := strings.TrimSpace(p.GetId()); v != "" {
		m["id"] = v
	}
	if v := strings.TrimSpace(p.GetEmail()); v != "" {
		m["email"] = v
	}
	if len(m) == 0 {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
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

func (s *Server) CreateHostedPaymentSession(ctx context.Context, req *paymentv1.CreateHostedPaymentSessionRequest) (*paymentv1.CreateHostedPaymentSessionResponse, error) {
	orderID, err := grpcerr.ParseUUID(req.GetOrderId(), "order id")
	if err != nil {
		return nil, err
	}
	buyerID, err := grpcerr.ParseUUID(req.GetBuyer().GetId(), "buyer id")
	if err != nil {
		return nil, err
	}
	returnURL := strings.TrimSpace(req.GetReturnUrl())
	if returnURL == "" {
		return nil, grpcerr.InvalidArgument("return_url is required")
	}
	merchantID, err := grpcerr.ParseUUID(req.GetMerchant().GetId(), "merchant id")
	if err != nil {
		return nil, err
	}
	if req.GetTotalCents() <= 0 {
		return nil, grpcerr.InvalidArgument("total_cents must be greater than zero")
	}
	shipping, err := addressToJSON(req.GetShippingAddress())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "shipping_address: %v", err)
	}
	lineItems, err := lineItemsToJSON(req.GetItems())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "items: %v", err)
	}
	buyer, err := partyToJSON(req.GetBuyer())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "buyer: %v", err)
	}
	merchant, err := partyToJSON(req.GetMerchant())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "merchant: %v", err)
	}

	session, err := s.svc.CreateHostedPaymentSession(ctx, service.CreateHostedPaymentSessionParams{
		OrderID:         orderID,
		BuyerUserID:     buyerID,
		MerchantID:      merchantID,
		TotalCents:      req.GetTotalCents(),
		Currency:        strings.TrimSpace(req.GetCurrency()),
		ShippingAddress: shipping,
		LineItems:       lineItems,
		Buyer:           buyer,
		Merchant:        merchant,
		ReturnURL:       returnURL,
	})
	if err != nil {
		return nil, grpcerr.Map(
			err,
			grpcerr.Mapping{Err: service.ErrInvalidSessionFacts, Code: codes.InvalidArgument, Message: "buyer, merchant, amount, and shipping are required"},
			grpcerr.Mapping{Err: service.ErrSessionTerminal, Code: codes.FailedPrecondition, Message: "hosted payment session is already terminal"},
		)
	}

	return &paymentv1.CreateHostedPaymentSessionResponse{
		OrderId:          session.OrderID,
		PaymentSessionId: session.PaymentSessionID,
		ReturnUrl:        session.ReturnURL,
	}, nil
}

func (s *Server) GetHostedPaymentSessionByOrder(ctx context.Context, req *paymentv1.GetHostedPaymentSessionByOrderRequest) (*paymentv1.HostedPaymentSession, error) {
	orderID, err := grpcerr.ParseUUID(req.GetOrderId(), "order id")
	if err != nil {
		return nil, err
	}
	session, err := s.svc.GetHostedPaymentSessionByOrder(ctx, orderID)
	if err != nil {
		return nil, grpcerr.Map(err, grpcerr.Mapping{Err: service.ErrIntentNotFound, Code: codes.NotFound, Message: "payment session not found"})
	}
	return mapHostedPaymentSession(session), nil
}

func hostedPaymentStatusProtoToString(v paymentv1.HostedPaymentSessionStatus) (string, error) {
	switch v {
	case paymentv1.HostedPaymentSessionStatus_HOSTED_PAYMENT_SESSION_STATUS_SUCCEEDED:
		return service.HostedPaymentSessionStatusSucceeded, nil
	case paymentv1.HostedPaymentSessionStatus_HOSTED_PAYMENT_SESSION_STATUS_FAILED:
		return service.HostedPaymentSessionStatusFailed, nil
	case paymentv1.HostedPaymentSessionStatus_HOSTED_PAYMENT_SESSION_STATUS_EXPIRED:
		return service.HostedPaymentSessionStatusExpired, nil
	case paymentv1.HostedPaymentSessionStatus_HOSTED_PAYMENT_SESSION_STATUS_UNSPECIFIED,
		paymentv1.HostedPaymentSessionStatus_HOSTED_PAYMENT_SESSION_STATUS_PENDING:
		return "", grpcerr.InvalidArgument("status must be a terminal hosted payment status")
	default:
		return "", grpcerr.InvalidArgument("unknown hosted payment status")
	}
}

func (s *Server) HandleGatewayWebhook(ctx context.Context, req *paymentv1.HandleGatewayWebhookRequest) (*paymentv1.HandleGatewayWebhookResponse, error) {
	orderID, err := grpcerr.ParseUUID(req.GetOrderId(), "order id")
	if err != nil {
		return nil, err
	}
	paymentSessionID := strings.TrimSpace(req.GetPaymentSessionId())
	if paymentSessionID == "" {
		return nil, grpcerr.InvalidArgument("payment_session_id is required")
	}
	statusValue, err := hostedPaymentStatusProtoToString(req.GetStatus())
	if err != nil {
		return nil, err
	}
	if err := s.svc.ApplyGatewayWebhook(ctx, orderID, paymentSessionID, statusValue, strings.TrimSpace(req.GetFailureReason())); err != nil {
		return nil, grpcerr.Map(
			err,
			grpcerr.Mapping{Err: service.ErrIntentNotFound, Code: codes.NotFound, Message: "payment session not found"},
			grpcerr.Mapping{Err: service.ErrSessionMismatch, Code: codes.NotFound, Message: "payment session not found"},
		)
	}
	return &paymentv1.HandleGatewayWebhookResponse{}, nil
}
