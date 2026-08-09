package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"refurbished-marketplace/services/cart/internal/service"

	testredis "refurbished-marketplace/shared/testutil/redis"

	"github.com/google/uuid"
)

func testConfig() service.Config {
	return service.Config{CartTTL: 24 * time.Hour}
}

func newCartService(t *testing.T) *service.Service {
	t.Helper()
	return service.New(testredis.SetupRedisContainer(t), testConfig())
}

const (
	testProductName = "Test Phone"
	testUnitPrice   = int64(1200)
)

func TestCartLifecycle(t *testing.T) {
	svc := newCartService(t)
	ctx := context.Background()

	t.Run("add cart item", func(t *testing.T) {
		cartID := uuid.NewString()
		itemID := uuid.NewString()
		merchantID := uuid.NewString()
		cart, err := svc.AddCartItem(ctx, cartID, itemID, merchantID, testProductName, 2, testUnitPrice)
		if err != nil {
			t.Fatalf("add item: %v", err)
		}
		if cart.CartID != cartID || len(cart.Items) != 1 || cart.Items[0].MerchantID != merchantID {
			t.Fatalf("unexpected cart after add")
		}
		if cart.Items[0].ProductName != testProductName || cart.Items[0].UnitPriceCents != testUnitPrice {
			t.Fatalf("unexpected snapshot after add: %+v", cart.Items[0])
		}
	})

	t.Run("get cart", func(t *testing.T) {
		cartID := uuid.NewString()
		itemID := uuid.NewString()
		merchantID := uuid.NewString()
		_, err := svc.AddCartItem(ctx, cartID, itemID, merchantID, testProductName, 2, testUnitPrice)
		if err != nil {
			t.Fatalf("add item: %v", err)
		}

		got, err := svc.GetCart(ctx, cartID)
		if err != nil {
			t.Fatalf("get cart: %v", err)
		}
		if got.CartID != cartID || len(got.Items) != 1 || got.Items[0].MerchantID != merchantID {
			t.Fatalf("unexpected cart after get")
		}
		if got.Items[0].ProductName != testProductName || got.Items[0].UnitPriceCents != testUnitPrice {
			t.Fatalf("unexpected snapshot after get: %+v", got.Items[0])
		}
	})

	t.Run("set cart item quantity", func(t *testing.T) {
		cartID := uuid.NewString()
		itemID := uuid.NewString()
		merchantID := uuid.NewString()
		_, err := svc.AddCartItem(ctx, cartID, itemID, merchantID, testProductName, 2, testUnitPrice)
		if err != nil {
			t.Fatalf("add item: %v", err)
		}

		updated, err := svc.SetCartItemQuantity(ctx, cartID, itemID, merchantID, "Updated Name", 5, 1500)
		if err != nil {
			t.Fatalf("set quantity: %v", err)
		}
		if updated.Items[0].Quantity != 5 || updated.Items[0].MerchantID != merchantID {
			t.Fatalf("expected quantity 5, got %d", updated.Items[0].Quantity)
		}
		if updated.Items[0].ProductName != "Updated Name" || updated.Items[0].UnitPriceCents != 1500 {
			t.Fatalf("expected refreshed snapshot, got %+v", updated.Items[0])
		}
	})

	t.Run("add refreshes snapshot on existing line", func(t *testing.T) {
		cartID := uuid.NewString()
		itemID := uuid.NewString()
		merchantID := uuid.NewString()
		_, err := svc.AddCartItem(ctx, cartID, itemID, merchantID, testProductName, 1, testUnitPrice)
		if err != nil {
			t.Fatalf("add item: %v", err)
		}
		got, err := svc.AddCartItem(ctx, cartID, itemID, merchantID, "New Name", 1, 999)
		if err != nil {
			t.Fatalf("add again: %v", err)
		}
		if len(got.Items) != 1 || got.Items[0].Quantity != 2 {
			t.Fatalf("expected qty 2, got %+v", got.Items)
		}
		if got.Items[0].ProductName != "New Name" || got.Items[0].UnitPriceCents != 999 {
			t.Fatalf("expected refreshed snapshot, got %+v", got.Items[0])
		}
	})

	t.Run("remove cart item", func(t *testing.T) {
		cartID := uuid.NewString()
		itemID := uuid.NewString()
		_, err := svc.AddCartItem(ctx, cartID, itemID, uuid.NewString(), testProductName, 2, testUnitPrice)
		if err != nil {
			t.Fatalf("add item: %v", err)
		}

		removed, err := svc.RemoveCartItem(ctx, cartID, itemID)
		if err != nil {
			t.Fatalf("remove item: %v", err)
		}
		if len(removed.Items) != 0 {
			t.Fatalf("expected empty cart after remove")
		}
	})

	t.Run("remove cart items multi", func(t *testing.T) {
		cartID := uuid.NewString()
		id1 := uuid.NewString()
		id2 := uuid.NewString()
		id3 := uuid.NewString()
		merchantID := uuid.NewString()
		for _, id := range []string{id1, id2, id3} {
			if _, err := svc.AddCartItem(ctx, cartID, id, merchantID, testProductName, 1, testUnitPrice); err != nil {
				t.Fatalf("add item: %v", err)
			}
		}
		updated, err := svc.RemoveCartItems(ctx, cartID, []string{id1, id3, uuid.NewString()})
		if err != nil {
			t.Fatalf("remove items: %v", err)
		}
		if len(updated.Items) != 1 || updated.Items[0].ProductID != id2 {
			t.Fatalf("unexpected remaining items: %+v", updated.Items)
		}
	})

	t.Run("clear cart", func(t *testing.T) {
		cartID := uuid.NewString()
		itemID := uuid.NewString()
		_, err := svc.AddCartItem(ctx, cartID, itemID, uuid.NewString(), testProductName, 2, testUnitPrice)
		if err != nil {
			t.Fatalf("add item: %v", err)
		}

		if err := svc.ClearCart(ctx, cartID); err != nil {
			t.Fatalf("clear cart: %v", err)
		}

		_, err = svc.GetCart(ctx, cartID)
		if err != nil {
			t.Fatalf("expected no error getting cart, got %v", err)
		}
	})
}

func TestCartValidation(t *testing.T) {
	t.Run("invalid cart id", func(t *testing.T) {
		svc := newCartService(t)
		ctx := context.Background()

		if _, err := svc.AddCartItem(ctx, "", uuid.NewString(), uuid.NewString(), testProductName, 1, testUnitPrice); !errors.Is(err, service.ErrInvalidCartID) {
			t.Fatalf("expected ErrInvalidCartID, got %v", err)
		}
	})

	t.Run("invalid product id", func(t *testing.T) {
		svc := newCartService(t)
		ctx := context.Background()

		if _, err := svc.AddCartItem(ctx, uuid.NewString(), "", uuid.NewString(), testProductName, 1, testUnitPrice); !errors.Is(err, service.ErrInvalidProductID) {
			t.Fatalf("expected ErrInvalidProductID, got %v", err)
		}
	})

	t.Run("invalid merchant id", func(t *testing.T) {
		svc := newCartService(t)
		ctx := context.Background()

		if _, err := svc.AddCartItem(ctx, uuid.NewString(), uuid.NewString(), "", testProductName, 1, testUnitPrice); !errors.Is(err, service.ErrInvalidMerchantID) {
			t.Fatalf("expected ErrInvalidMerchantID, got %v", err)
		}
	})

	t.Run("invalid quantity", func(t *testing.T) {
		svc := newCartService(t)
		ctx := context.Background()

		if _, err := svc.AddCartItem(ctx, uuid.NewString(), uuid.NewString(), uuid.NewString(), testProductName, 0, testUnitPrice); !errors.Is(err, service.ErrInvalidQuantity) {
			t.Fatalf("expected ErrInvalidQuantity, got %v", err)
		}
	})

	t.Run("invalid snapshot name", func(t *testing.T) {
		svc := newCartService(t)
		ctx := context.Background()

		if _, err := svc.AddCartItem(ctx, uuid.NewString(), uuid.NewString(), uuid.NewString(), "  ", 1, testUnitPrice); !errors.Is(err, service.ErrInvalidSnapshot) {
			t.Fatalf("expected ErrInvalidSnapshot, got %v", err)
		}
	})

	t.Run("invalid snapshot price", func(t *testing.T) {
		svc := newCartService(t)
		ctx := context.Background()

		if _, err := svc.AddCartItem(ctx, uuid.NewString(), uuid.NewString(), uuid.NewString(), testProductName, 1, 0); !errors.Is(err, service.ErrInvalidSnapshot) {
			t.Fatalf("expected ErrInvalidSnapshot, got %v", err)
		}
	})

	t.Run("empty multi remove", func(t *testing.T) {
		svc := newCartService(t)
		ctx := context.Background()

		if _, err := svc.RemoveCartItems(ctx, uuid.NewString(), nil); !errors.Is(err, service.ErrEmptyProductIDs) {
			t.Fatalf("expected ErrEmptyProductIDs, got %v", err)
		}
	})

	t.Run("missing item", func(t *testing.T) {
		svc := newCartService(t)
		ctx := context.Background()

		if _, err := svc.RemoveCartItem(ctx, uuid.NewString(), uuid.NewString()); !errors.Is(err, service.ErrItemNotFound) {
			t.Fatalf("expected ErrItemNotFound, got %v", err)
		}
	})
}
