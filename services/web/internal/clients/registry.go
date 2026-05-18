package clients

import "fmt"

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
		usersClient.Close()
		return nil, fmt.Errorf("products grpc client: %w", err)
	}

	ordersClient, err := newOrdersClient(cfg.OrdersAddr)
	if err != nil {
		usersClient.Close()
		productsClient.Close()
		return nil, fmt.Errorf("orders grpc client: %w", err)
	}

	cartClient, err := newCartClient(cfg.CartAddr)
	if err != nil {
		usersClient.Close()
		productsClient.Close()
		ordersClient.Close()
		return nil, fmt.Errorf("cart grpc client: %w", err)
	}

	paymentClient, err := newPaymentClient(cfg.PaymentAddr)
	if err != nil {
		usersClient.Close()
		productsClient.Close()
		ordersClient.Close()
		cartClient.Close()
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
	if c.Users != nil {
		c.Users.Close()
	}
	if c.Products != nil {
		c.Products.Close()
	}
	if c.Orders != nil {
		c.Orders.Close()
	}
	if c.Cart != nil {
		c.Cart.Close()
	}
	if c.Payment != nil {
		c.Payment.Close()
	}
}
