package shared

import (
	"context"
	"net/url"
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
	{Label: "Sign in", Href: "/auth/login", GuestOnly: true},
	{Label: "Sign up", Href: "/auth/register", GuestOnly: true},
}

type authStateKey struct{}

type AuthState struct {
	Authenticated bool
}

func WithAuthState(ctx context.Context, state AuthState) context.Context {
	return context.WithValue(ctx, authStateKey{}, state)
}

func SessionSignals(ctx context.Context) string {
	state, _ := ctx.Value(authStateKey{}).(AuthState)
	if state.Authenticated {
		return `{"session":{"authenticated":true}}`
	}
	return `{"session":{"authenticated":false}}`
}

func LoginURL(next string) string {
	if next == "" || next == "/products" {
		return "/auth/login"
	}
	return "/auth/login?next=" + urlQueryEscape(next)
}

func RegisterURL(next string) string {
	if next == "" || next == "/products" {
		return "/auth/register"
	}
	return "/auth/register?next=" + urlQueryEscape(next)
}

func urlQueryEscape(v string) string {
	return url.QueryEscape(v)
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
