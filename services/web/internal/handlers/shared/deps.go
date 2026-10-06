package shared

import (
	"context"

	cartv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/cart/v1"
	checkoutv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/checkout/v1"
	inventoryv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/inventory/v1"
	ordersv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/orders/v1"
	paymentv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/payment/v1"
	productsv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/products/v1"
	searchv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/search/v1"
	usersv1 "github.com/phuchoang2603/refurbished-marketplace/shared/proto/users/v1"
)

type UsersService interface {
	Login(ctx context.Context, email, password string) (*usersv1.TokenResponse, error)
	Logout(ctx context.Context, refreshToken string) (*usersv1.LogoutResponse, error)
	CreateUser(ctx context.Context, email, password string) (*usersv1.User, error)
}

type ProductsService interface {
	CreateProduct(ctx context.Context, name, description string, priceCents int64, merchantID string, initialStock int32) (*productsv1.Product, error)
	GetProductByID(ctx context.Context, id string) (*productsv1.Product, error)
	GetProductsByIDs(ctx context.Context, ids []string) (*productsv1.GetProductsByIDsResponse, error)
}

type InventoryService interface {
	GetStock(ctx context.Context, productID string) (*inventoryv1.Stock, error)
	GetStocksByIDs(ctx context.Context, productIDs []string) (*inventoryv1.GetStocksByIDsResponse, error)
}

type SearchService interface {
	SearchProducts(ctx context.Context, query, merchantID string, limit, offset int32) (*searchv1.SearchProductsResponse, error)
}

type OrdersService interface {
	GetOrderByID(ctx context.Context, id string) (*ordersv1.Order, error)
	ListOrdersByBuyer(ctx context.Context, buyerUserID string, limit, offset int32) (*ordersv1.ListOrdersByBuyerResponse, error)
}

type CartService interface {
	GetCart(ctx context.Context, cartID string) (*cartv1.Cart, error)
	AddCartItem(ctx context.Context, cartID, productID, merchantID, productName string, quantity int32, unitPriceCents int64) (*cartv1.Cart, error)
	SetCartItemQuantity(ctx context.Context, cartID, productID, merchantID, productName string, quantity int32, unitPriceCents int64) (*cartv1.Cart, error)
	RemoveCartItems(ctx context.Context, cartID string, productIDs []string) (*cartv1.Cart, error)
}

type PaymentService interface {
	GetHostedPaymentSessionByOrder(ctx context.Context, orderID string) (*paymentv1.HostedPaymentSession, error)
	HandleGatewayWebhook(ctx context.Context, req *paymentv1.HandleGatewayWebhookRequest) (*paymentv1.HandleGatewayWebhookResponse, error)
}

type CheckoutService interface {
	SubmitCheckout(ctx context.Context, request *checkoutv1.SubmitCheckoutRequest) (*checkoutv1.SubmitCheckoutResponse, error)
	GetCheckout(ctx context.Context, request *checkoutv1.GetCheckoutRequest) (*checkoutv1.CheckoutStatus, error)
}

type Dependencies struct {
	Users         UsersService
	Products      ProductsService
	Inventory     InventoryService
	Search        SearchService
	Orders        OrdersService
	Cart          CartService
	Payment       PaymentService
	Checkout      CheckoutService
	HostedPayment HostedPaymentConfig
}
