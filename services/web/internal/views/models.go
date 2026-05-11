package views

import (
	"context"
	"fmt"
)

type NavLink struct {
	Label     string
	Href      string
	GuestOnly bool
}

var DefaultNav = []NavLink{
	{Label: "Products", Href: "/products"},
	{Label: "Cart", Href: "/cart"},
	{Label: "Orders", Href: "/orders"},
	{Label: "Account", Href: "/auth/login", GuestOnly: true},
	{Label: "Sign up", Href: "/auth/register", GuestOnly: true},
}

type authStateKey struct{}

type AuthState struct {
	Authenticated bool
}

func WithAuthState(ctx context.Context, state AuthState) context.Context {
	return context.WithValue(ctx, authStateKey{}, state)
}

func sessionSignals(ctx context.Context) string {
	state, _ := ctx.Value(authStateKey{}).(AuthState)
	if state.Authenticated {
		return `{"session":{"authenticated":true}}`
	}
	return `{"session":{"authenticated":false}}`
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
	TokenType string
	ExpiresIn int64
}

type CartItemView struct {
	ProductID      string
	ProductName    string
	ProductPrice   int64
	LineTotalCents int64
	Available      bool
	Quantity       int32
}

type CartView struct {
	CartID              string
	Items               []CartItemView
	EstimatedTotalCents int64
	CreatedAt           string
	UpdatedAt           string
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
