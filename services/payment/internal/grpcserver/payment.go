package grpcserver

import (
	"context"
	"encoding/json"
	"strings"

	"refurbished-marketplace/services/payment/internal/service"
	"refurbished-marketplace/shared/grpcerr"
	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"

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
	case service.HostedPaymentSessionStatusCancelled:
		return paymentv1.HostedPaymentSessionStatus_HOSTED_PAYMENT_SESSION_STATUS_CANCELLED
	case service.HostedPaymentSessionStatusExpired:
		return paymentv1.HostedPaymentSessionStatus_HOSTED_PAYMENT_SESSION_STATUS_EXPIRED
	default:
		return paymentv1.HostedPaymentSessionStatus_HOSTED_PAYMENT_SESSION_STATUS_UNSPECIFIED
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

func (s *Server) CreateHostedPaymentSession(ctx context.Context, req *paymentv1.CreateHostedPaymentSessionRequest) (*paymentv1.CreateHostedPaymentSessionResponse, error) {
	orderID, err := grpcerr.ParseUUID(req.GetOrderId(), "order id")
	if err != nil {
		return nil, err
	}
	buyerID, err := grpcerr.ParseUUID(req.GetBuyerUserId(), "buyer user id")
	if err != nil {
		return nil, err
	}
	returnURL := strings.TrimSpace(req.GetReturnUrl())
	if returnURL == "" {
		return nil, grpcerr.InvalidArgument("return_url is required")
	}
	cancelURL := strings.TrimSpace(req.GetCancelUrl())
	if cancelURL == "" {
		cancelURL = returnURL
	}
	shipping, err := addressToJSON(req.GetShippingAddress())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "shipping_address: %v", err)
	}

	session, err := s.svc.CreateHostedPaymentSession(ctx, service.CreateHostedPaymentSessionParams{
		OrderID:         orderID,
		BuyerUserID:     buyerID,
		Currency:        strings.TrimSpace(req.GetCurrency()),
		ShippingAddress: shipping,
		ReturnURL:       returnURL,
		CancelURL:       cancelURL,
	})
	if err != nil {
		return nil, grpcerr.Internal()
	}

	return &paymentv1.CreateHostedPaymentSessionResponse{
		OrderId:          session.OrderID,
		PaymentSessionId: session.PaymentSessionID,
		ReturnUrl:        session.ReturnURL,
		CancelUrl:        session.CancelURL,
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
	case paymentv1.HostedPaymentSessionStatus_HOSTED_PAYMENT_SESSION_STATUS_CANCELLED:
		return service.HostedPaymentSessionStatusCancelled, nil
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
