package views

import "fmt"

type NavLink struct {
	Label string
	Href  string
}

var DefaultNav = []NavLink{
	{Label: "Products", Href: "/products"},
	{Label: "Cart", Href: "/cart"},
	{Label: "Orders", Href: "/orders"},
	{Label: "Auth", Href: "/auth/login"},
}

type ProductView struct {
	ID          string
	Name        string
	Description string
	PriceCents  int64
	Stock       int32
	CreatedAt   string
	UpdatedAt   string
}

type UserView struct {
	ID        string
	Email     string
	CreatedAt string
	UpdatedAt string
}

type TokenView struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int64
}

type CartItemView struct {
	ProductID string
	Quantity  int32
}

type CartView struct {
	CartID    string
	Items     []CartItemView
	CreatedAt string
	UpdatedAt string
}

type OrderItemView struct {
	ID             string
	OrderID        string
	ProductID      string
	Quantity       int32
	UnitPriceCents int64
	LineTotalCents int64
	CreatedAt      string
}

type OrderView struct {
	ID          string
	BuyerUserID string
	Status      string
	TotalCents  int64
	Items       []OrderItemView
	CreatedAt   string
	UpdatedAt   string
}

func formatCents(v int64) string {
	return "$" + formatInt64(v)
}

func formatInt32(v int32) string {
	return formatInt64(int64(v))
}

func formatInt64(v int64) string {
	return fmt.Sprintf("%d", v)
}
