package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"refurbished-marketplace/services/cart/internal/service"

	"github.com/google/uuid"
	"refurbished-marketplace/shared/testutil"
)

func TestCartLifecycle(t *testing.T) {
	svc := service.New(testutil.SetupRedisContainer(t), 24*time.Hour)
	ctx := context.Background()

	t.Run("add cart item", func(t *testing.T) {
		cartID := uuid.NewString()
		itemID := uuid.NewString()
		cart, err := svc.AddCartItem(ctx, cartID, itemID, 2)
		if err != nil {
			t.Fatalf("add item: %v", err)
		}
		if cart.CartID != cartID || len(cart.Items) != 1 {
			t.Fatalf("unexpected cart after add")
		}
	})

	t.Run("get cart", func(t *testing.T) {
		cartID := uuid.NewString()
		itemID := uuid.NewString()
		_, err := svc.AddCartItem(ctx, cartID, itemID, 2)
		if err != nil {
			t.Fatalf("add item: %v", err)
		}

		got, err := svc.GetCart(ctx, cartID)
		if err != nil {
			t.Fatalf("get cart: %v", err)
		}
		if got.CartID != cartID || len(got.Items) != 1 {
			t.Fatalf("unexpected cart after get")
		}
	})

	t.Run("set cart item quantity", func(t *testing.T) {
		cartID := uuid.NewString()
		itemID := uuid.NewString()
		_, err := svc.AddCartItem(ctx, cartID, itemID, 2)
		if err != nil {
			t.Fatalf("add item: %v", err)
		}

		updated, err := svc.SetCartItemQuantity(ctx, cartID, itemID, 5)
		if err != nil {
			t.Fatalf("set quantity: %v", err)
		}
		if updated.Items[0].Quantity != 7 {
			t.Fatalf("expected quantity 7, got %d", updated.Items[0].Quantity)
		}
	})

	t.Run("remove cart item", func(t *testing.T) {
		cartID := uuid.NewString()
		itemID := uuid.NewString()
		_, err := svc.AddCartItem(ctx, cartID, itemID, 2)
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

	t.Run("clear cart", func(t *testing.T) {
		cartID := uuid.NewString()
		itemID := uuid.NewString()
		_, err := svc.AddCartItem(ctx, cartID, itemID, 2)
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
		svc := service.New(testutil.SetupRedisContainer(t), 24*time.Hour)
		ctx := context.Background()

		if _, err := svc.AddCartItem(ctx, "", uuid.NewString(), 1); !errors.Is(err, service.ErrInvalidCartID) {
			t.Fatalf("expected ErrInvalidCartID, got %v", err)
		}
	})

	t.Run("invalid product id", func(t *testing.T) {
		svc := service.New(testutil.SetupRedisContainer(t), 24*time.Hour)
		ctx := context.Background()

		if _, err := svc.AddCartItem(ctx, uuid.NewString(), "", 1); !errors.Is(err, service.ErrInvalidProductID) {
			t.Fatalf("expected ErrInvalidProductID, got %v", err)
		}
	})

	t.Run("invalid quantity", func(t *testing.T) {
		svc := service.New(testutil.SetupRedisContainer(t), 24*time.Hour)
		ctx := context.Background()

		if _, err := svc.AddCartItem(ctx, uuid.NewString(), uuid.NewString(), 0); !errors.Is(err, service.ErrInvalidQuantity) {
			t.Fatalf("expected ErrInvalidQuantity, got %v", err)
		}
	})

	t.Run("missing item", func(t *testing.T) {
		svc := service.New(testutil.SetupRedisContainer(t), 24*time.Hour)
		ctx := context.Background()

		if _, err := svc.RemoveCartItem(ctx, uuid.NewString(), uuid.NewString()); !errors.Is(err, service.ErrItemNotFound) {
			t.Fatalf("expected ErrItemNotFound, got %v", err)
		}
	})
}
