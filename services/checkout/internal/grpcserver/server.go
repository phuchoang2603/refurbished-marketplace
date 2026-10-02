package grpcserver

import (
	"context"

	"github.com/google/uuid"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	"google.golang.org/grpc/codes"
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
}

func New(reader Reader) *Server {
	return &Server{reader: reader}
}

func (server *Server) SubmitCheckout(ctx context.Context, request *checkoutv1.SubmitCheckoutRequest) (*checkoutv1.SubmitCheckoutResponse, error) {
	submitter, ok := server.reader.(Submitter)
	if !ok {
		return nil, status.Error(codes.Unimplemented, "checkout submission unavailable")
	}
	return submitter.SubmitCheckout(ctx, request)
}

func (server *Server) GetCheckout(ctx context.Context, request *checkoutv1.GetCheckoutRequest) (*checkoutv1.CheckoutStatus, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "checkout request is required")
	}
	checkoutID, err := uuid.Parse(request.GetCheckoutId())
	if err != nil || checkoutID == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "invalid checkout id")
	}
	buyerID, err := uuid.Parse(request.GetBuyerUserId())
	if err != nil || buyerID == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "invalid buyer id")
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
