package clients

import (
	"fmt"
	"log"
)

type Config struct {
	UsersAddr    string
	ProductsAddr string
	OrdersAddr   string
	CartAddr     string
	PaymentAddr  string
}

type Clients struct {
	Users    *UsersClient
	Products *ProductsClient
	Orders   *OrdersClient
	Cart     *CartClient
	Payment  *PaymentClient
}

func New(cfg Config) (*Clients, error) {
	usersClient, err := newUsersClient(cfg.UsersAddr)
	if err != nil {
		return nil, fmt.Errorf("users grpc client: %w", err)
	}

	productsClient, err := newProductsClient(cfg.ProductsAddr)
	if err != nil {
		closeClient(usersClient)
		return nil, fmt.Errorf("products grpc client: %w", err)
	}

	ordersClient, err := newOrdersClient(cfg.OrdersAddr)
	if err != nil {
		closeClient(usersClient)
		closeClient(productsClient)
		return nil, fmt.Errorf("orders grpc client: %w", err)
	}

	cartClient, err := newCartClient(cfg.CartAddr)
	if err != nil {
		closeClient(usersClient)
		closeClient(productsClient)
		closeClient(ordersClient)
		return nil, fmt.Errorf("cart grpc client: %w", err)
	}

	paymentClient, err := newPaymentClient(cfg.PaymentAddr)
	if err != nil {
		closeClient(usersClient)
		closeClient(productsClient)
		closeClient(ordersClient)
		closeClient(cartClient)
		return nil, fmt.Errorf("payment grpc client: %w", err)
	}

	return &Clients{
		Users:    usersClient,
		Products: productsClient,
		Orders:   ordersClient,
		Cart:     cartClient,
		Payment:  paymentClient,
	}, nil
}

func (c *Clients) Close() {
	if c == nil {
		return
	}
	closeClient(c.Users)
	closeClient(c.Products)
	closeClient(c.Orders)
	closeClient(c.Cart)
	closeClient(c.Payment)
}

func closeClient(client interface{ Close() error }) {
	if client == nil {
		return
	}
	if err := client.Close(); err != nil {
		log.Printf("close grpc client: %v", err)
	}
}
