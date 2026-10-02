package grpcserver

import (
	"context"
	"strings"

	"github.com/phuchoang2603/refurbished-marketplace/services/payment/internal/service"
	"github.com/phuchoang2603/refurbished-marketplace/shared/err/grpcerr"
	paymentv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/payment/v1"

	"google.golang.org/grpc/codes"
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
