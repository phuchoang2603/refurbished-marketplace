package service

import (
	"context"
	"encoding/json"
	"net/url"
	"time"

	"github.com/google/uuid"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	usersv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/users/v1"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func validGrading(category, condition string) bool {
	switch category {
	case "phone", "laptop", "tablet", "accessory", "headphones", "smartwatch", "game_console", "camera":
	default:
		return false
	}
	return condition == "excellent" || condition == "good" || condition == "fair"
}

func (service *Service) freezeGatewaySnapshot(ctx context.Context, request *checkoutv1.SubmitCheckoutRequest, orderID uuid.UUID) ([]byte, error) {
	if service.users == nil {
		return nil, status.Error(codes.Unavailable, "account service unavailable")
	}
	group, groupCtx := errgroup.WithContext(ctx)
	accountTime := func(id string) (time.Time, error) {
		accountCtx, cancel := context.WithTimeout(groupCtx, 5*time.Second)
		defer cancel()
		user, err := service.users.GetUserByID(accountCtx, &usersv1.GetUserByIDRequest{Id: id})
		if err != nil || user.GetId() != id || user.GetCreatedAt() == nil || user.GetCreatedAt().CheckValid() != nil || user.GetCreatedAt().AsTime().IsZero() {
			return time.Time{}, status.Error(codes.Unavailable, "account facts unavailable")
		}
		return user.GetCreatedAt().AsTime().UTC(), nil
	}
	var sellerTime, buyerTime time.Time
	group.Go(func() error {
		var err error
		sellerTime, err = accountTime(request.GetMerchantId())
		return err
	})
	group.Go(func() error {
		var err error
		buyerTime, err = accountTime(request.GetBuyerUserId())
		return err
	})
	if err := group.Wait(); err != nil {
		return nil, err
	}
	base, err := url.Parse(service.config.CheckoutReturnBaseURL)
	returned, returnErr := url.Parse(request.GetReturnUrl())
	if err != nil || returnErr != nil || base == nil || returned == nil || returned.Scheme != "https" || returned.Host != base.Host || returned.User != nil || len(returned.String()) > 2048 {
		return nil, status.Error(codes.InvalidArgument, "invalid checkout return URL")
	}
	address := request.GetShippingAddress()
	snapshot := gatewayRequest{OrderReference: orderID.String(), Seller: gatewayParty{ExternalID: request.GetMerchantId(), AccountCreatedAt: sellerTime}, Buyer: gatewayParty{ExternalID: request.GetBuyerUserId(), AccountCreatedAt: buyerTime}, Amount: gatewayMoney{MinorUnits: request.GetTotalCents(), Currency: request.GetCurrency()}, ReturnURL: returned.String(), ShippingAddress: gatewayAddress{Name: address.GetName(), Line1: address.GetLine1(), Line2: address.GetLine2(), City: address.GetCity(), Region: address.GetRegion(), PostalCode: address.GetPostalCode(), Country: address.GetCountry()}}
	for _, item := range request.GetItems() {
		snapshot.Items = append(snapshot.Items, gatewayItem{ProductID: item.GetProductId(), Category: item.GetCategory(), Condition: item.GetCondition(), Quantity: item.GetQuantity(), UnitPrice: item.GetUnitPriceCents()})
	}
	raw, err := json.Marshal(snapshot)
	if err != nil || len(raw) > 64<<10 {
		return nil, status.Error(codes.InvalidArgument, "gateway snapshot is too large")
	}
	return raw, nil
}
