package shared

import (
	"context"

	cartv1 "refurbished-marketplace/shared/proto/cart/v1"
	ordersv1 "refurbished-marketplace/shared/proto/orders/v1"
	paymentv1 "refurbished-marketplace/shared/proto/payment/v1"
	productsv1 "refurbished-marketplace/shared/proto/products/v1"
	usersv1 "refurbished-marketplace/shared/proto/users/v1"
)

type UsersService interface {
	Login(ctx context.Context, email, password string) (*usersv1.TokenResponse, error)
	Logout(ctx context.Context, refreshToken string) (*usersv1.LogoutResponse, error)
	CreateUser(ctx context.Context, email, password string) (*usersv1.User, error)
}

type ProductsService interface {
	GetProductByID(ctx context.Context, id string) (*productsv1.Product, error)
	ListProducts(ctx context.Context, limit, offset int32) (*productsv1.ListProductsResponse, error)
}

type OrdersService interface {
	CreateOrder(ctx context.Context, buyerUserID, merchantID string, items []*ordersv1.CreateOrderItem, totalCents int64) (*ordersv1.Order, error)
	GetOrderByID(ctx context.Context, id string) (*ordersv1.Order, error)
	ListOrdersByBuyer(ctx context.Context, buyerUserID string, limit, offset int32) (*ordersv1.ListOrdersByBuyerResponse, error)
}

type CartService interface {
	GetCart(ctx context.Context, cartID string) (*cartv1.Cart, error)
	AddCartItem(ctx context.Context, cartID, productID string, quantity int32) (*cartv1.Cart, error)
	SetCartItemQuantity(ctx context.Context, cartID, productID string, quantity int32) (*cartv1.Cart, error)
	RemoveCartItem(ctx context.Context, cartID, productID string) (*cartv1.Cart, error)
	ClearCart(ctx context.Context, cartID string) error
}

type PaymentService interface {
	HandleGatewayWebhook(ctx context.Context, req *paymentv1.HandleGatewayWebhookRequest) (*paymentv1.HandleGatewayWebhookResponse, error)
}

type Dependencies struct {
	Users    UsersService
	Products ProductsService
	Orders   OrdersService
	Cart     CartService
	Payment  PaymentService
}
