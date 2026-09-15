package clients

import (
	"fmt"

	sharedlog "github.com/phuchoang2603/refurbished-marketplace/shared/observe/log"
)

type Config struct {
	UsersAddr     string
	ProductsAddr  string
	InventoryAddr string
	SearchAddr    string
	OrdersAddr    string
	CartAddr      string
	PaymentAddr   string
}

type Clients struct {
	Users     *UsersClient
	Products  *ProductsClient
	Inventory *InventoryClient
	Search    *SearchClient
	Orders    *OrdersClient
	Cart      *CartClient
	Payment   *PaymentClient
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

	inventoryClient, err := newInventoryClient(cfg.InventoryAddr)
	if err != nil {
		closeClient(usersClient)
		closeClient(productsClient)
		return nil, fmt.Errorf("inventory grpc client: %w", err)
	}

	searchClient, err := newSearchClient(cfg.SearchAddr)
	if err != nil {
		closeClient(usersClient)
		closeClient(productsClient)
		closeClient(inventoryClient)
		return nil, fmt.Errorf("search grpc client: %w", err)
	}

	ordersClient, err := newOrdersClient(cfg.OrdersAddr)
	if err != nil {
		closeClient(usersClient)
		closeClient(productsClient)
		closeClient(inventoryClient)
		closeClient(searchClient)
		return nil, fmt.Errorf("orders grpc client: %w", err)
	}

	cartClient, err := newCartClient(cfg.CartAddr)
	if err != nil {
		closeClient(usersClient)
		closeClient(productsClient)
		closeClient(inventoryClient)
		closeClient(searchClient)
		closeClient(ordersClient)
		return nil, fmt.Errorf("cart grpc client: %w", err)
	}

	paymentClient, err := newPaymentClient(cfg.PaymentAddr)
	if err != nil {
		closeClient(usersClient)
		closeClient(productsClient)
		closeClient(inventoryClient)
		closeClient(searchClient)
		closeClient(ordersClient)
		closeClient(cartClient)
		return nil, fmt.Errorf("payment grpc client: %w", err)
	}

	return &Clients{
		Users:     usersClient,
		Products:  productsClient,
		Inventory: inventoryClient,
		Search:    searchClient,
		Orders:    ordersClient,
		Cart:      cartClient,
		Payment:   paymentClient,
	}, nil
}

func (c *Clients) Close() {
	if c == nil {
		return
	}
	closeClient(c.Users)
	closeClient(c.Products)
	closeClient(c.Inventory)
	closeClient(c.Search)
	closeClient(c.Orders)
	closeClient(c.Cart)
	closeClient(c.Payment)
}

func closeClient(client interface{ Close() error }) {
	if client == nil {
		return
	}
	if err := client.Close(); err != nil {
		sharedlog.Error("close grpc client", "err", err)
	}
}
