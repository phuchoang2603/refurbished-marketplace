package grpcserver

import (
	"context"
	"strings"

	"github.com/google/uuid"
	authconfig "github.com/phuchoang2603/refurbished-marketplace/shared/auth/config"
	sharedjwt "github.com/phuchoang2603/refurbished-marketplace/shared/auth/jwt"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type CheckoutView struct {
	CheckoutID       uuid.UUID
	BuyerUserID      uuid.UUID
	OrderID          string
	State            checkoutv1.CheckoutState
	PaymentSessionID string
	ReturnURL        string
	FailureReason    string
}

type Reader interface {
	GetCheckout(context.Context, uuid.UUID) (CheckoutView, error)
}

type Submitter interface {
	SubmitCheckout(context.Context, *checkoutv1.SubmitCheckoutRequest) (*checkoutv1.SubmitCheckoutResponse, error)
}

type Server struct {
	checkoutv1.UnimplementedCheckoutServiceServer
	reader Reader
	auth   authconfig.Config
}

func New(reader Reader, auth authconfig.Config) *Server {
	return &Server{reader: reader, auth: auth}
}

func (server *Server) authenticatedBuyer(ctx context.Context) (string, error) {
	values, ok := metadata.FromIncomingContext(ctx)
	if !ok || len(values.Get("authorization")) != 1 {
		return "", status.Error(codes.Unauthenticated, "access token required")
	}
	raw, ok := strings.CutPrefix(values.Get("authorization")[0], "Bearer ")
	if !ok || raw == "" {
		return "", status.Error(codes.Unauthenticated, "access token required")
	}
	claims, err := sharedjwt.ParseAndValidate(raw, server.auth.JWTSecret, "access", server.auth.JWTIssuer, server.auth.JWTAudience)
	if err != nil {
		return "", status.Error(codes.Unauthenticated, "invalid access token")
	}
	return claims.Subject, nil
}

func (server *Server) SubmitCheckout(ctx context.Context, request *checkoutv1.SubmitCheckoutRequest) (*checkoutv1.SubmitCheckoutResponse, error) {
	buyerID, err := server.authenticatedBuyer(ctx)
	if err != nil {
		return nil, err
	}
	if request == nil || request.GetBuyerUserId() != buyerID {
		return nil, status.Error(codes.PermissionDenied, "checkout buyer does not match access token")
	}
	submitter, ok := server.reader.(Submitter)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "checkout submission unavailable")
	}
	return submitter.SubmitCheckout(ctx, request)
}

func (server *Server) GetCheckout(ctx context.Context, request *checkoutv1.GetCheckoutRequest) (*checkoutv1.CheckoutStatus, error) {
	authenticatedBuyer, err := server.authenticatedBuyer(ctx)
	if err != nil {
		return nil, err
	}
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "checkout request is required")
	}
	checkoutID, err := uuid.Parse(request.GetCheckoutId())
	if err != nil || checkoutID == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "invalid checkout id")
	}
	buyerID, err := uuid.Parse(authenticatedBuyer)
	if err != nil || buyerID == uuid.Nil {
		return nil, status.Error(codes.Unauthenticated, "invalid buyer identity")
	}
	if request.GetBuyerUserId() != authenticatedBuyer {
		return nil, status.Error(codes.NotFound, "checkout not found")
	}
	view, err := server.reader.GetCheckout(ctx, checkoutID)
	if err != nil {
		return nil, err
	}
	if view.BuyerUserID != buyerID {
		return nil, status.Error(codes.NotFound, "checkout not found")
	}
	return &checkoutv1.CheckoutStatus{
		CheckoutId:       view.CheckoutID.String(),
		OrderId:          view.OrderID,
		State:            view.State,
		PaymentSessionId: view.PaymentSessionID,
		ReturnUrl:        view.ReturnURL,
		FailureReason:    view.FailureReason,
		BuyerUserId:      view.BuyerUserID.String(),
	}, nil
}
