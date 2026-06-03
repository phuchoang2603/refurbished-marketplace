package shared

import (
	"context"
	"net/url"
)

type NavLink struct {
	Label      string
	Href       string
	Visibility NavVisibility
}

type NavVisibility int

const (
	NavVisible NavVisibility = iota
	NavVisibleWhenAuthenticated
	NavVisibleWhenGuest
)

var DefaultNav = []NavLink{
	{Label: "Products", Href: "/products"},
	{Label: "Cart", Href: "/cart"},
	{Label: "Sell", Href: "/seller/products", Visibility: NavVisibleWhenAuthenticated},
	{Label: "Orders", Href: "/orders", Visibility: NavVisibleWhenAuthenticated},
	{Label: "Sign in", Href: "/auth/login", Visibility: NavVisibleWhenGuest},
	{Label: "Sign up", Href: "/auth/register", Visibility: NavVisibleWhenGuest},
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
	MerchantID  string
	IsOwner     bool
	Name        string
	Description string
	PriceCents  int64
	Stock       int32
	CreatedAt   string
	UpdatedAt   string
}

type CartItemView struct {
	ProductID      string
	MerchantID     string
	ProductName    string
	ProductPrice   int64
	LineTotalCents int64
	Available      bool
	Quantity       int32
}

type CartMerchantGroupView struct {
	MerchantID    string
	Items         []CartItemView
	SubtotalCents int64
}

type CartView struct {
	CartID              string
	Items               []CartItemView
	MerchantGroups      []CartMerchantGroupView
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
	ID                   string
	BuyerUserID          string
	Status               string
	PaymentStatus        string
	PaymentFailureReason string
	TotalCents           int64
	Items                []OrderItemView
	CreatedAt            string
	UpdatedAt            string
}

type UnavailableView struct {
	PageTitle string
	SectionID string
	Title     string
	Subtitle  string
}
