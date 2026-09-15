package tests

import (
	"errors"
	"testing"

	"github.com/phuchoang2603/refurbished-marketplace/services/products/internal/catalog"
	"github.com/phuchoang2603/refurbished-marketplace/services/products/internal/service"
	testmongo "github.com/phuchoang2603/refurbished-marketplace/shared/testutil/mongo"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

func newProductsService(t *testing.T) *service.Service {
	t.Helper()
	return service.New(newProductsStore(t))
}

func newProductsStore(t *testing.T) *catalog.Store {
	t.Helper()
	return catalog.New(testmongo.SetupReplicaSet(t))
}

func TestCreateAndReadProducts(t *testing.T) {
	svc := newProductsService(t)
	ctx := t.Context()

	t.Run("create product", func(t *testing.T) {
		merchantID := uuid.New()
		created, err := svc.CreateProduct(ctx, "iPhone 13", "Refurbished - Grade A", 49900, merchantID, proto.Int32(5))
		if err != nil {
			t.Fatalf("create product: %v", err)
		}
		if created.Name == "" {
			t.Fatalf("expected created product")
		}
	})

	t.Run("get product by id", func(t *testing.T) {
		merchantID := uuid.New()
		created, err := svc.CreateProduct(ctx, "iPhone 13", "Refurbished - Grade A", 49900, merchantID, proto.Int32(5))
		if err != nil {
			t.Fatalf("create product: %v", err)
		}

		got, err := svc.GetProductByID(ctx, created.ID)
		if err != nil {
			t.Fatalf("get product: %v", err)
		}

		if got.Name != created.Name {
			t.Fatalf("expected name %q, got %q", created.Name, got.Name)
		}
		if got.MerchantID != merchantID {
			t.Fatalf("expected merchant id %s, got %s", merchantID, got.MerchantID)
		}
	})
}

func TestProductValidation(t *testing.T) {
	t.Run("missing product", func(t *testing.T) {
		svc := newProductsService(t)
		ctx := t.Context()

		_, err := svc.GetProductByID(ctx, uuid.New())
		if !errors.Is(err, service.ErrProductNotFound) {
			t.Fatalf("expected ErrProductNotFound, got %v", err)
		}
	})

	t.Run("invalid product name", func(t *testing.T) {
		svc := newProductsService(t)
		ctx := t.Context()

		_, err := svc.CreateProduct(ctx, "", "x", 100, uuid.New(), proto.Int32(5))
		if !errors.Is(err, service.ErrInvalidProductName) {
			t.Fatalf("expected ErrInvalidProductName, got %v", err)
		}
	})

	t.Run("invalid merchant id", func(t *testing.T) {
		svc := newProductsService(t)
		ctx := t.Context()

		_, err := svc.CreateProduct(ctx, "Laptop", "x", 100, uuid.Nil, proto.Int32(5))
		if !errors.Is(err, service.ErrInvalidMerchantID) {
			t.Fatalf("expected ErrInvalidMerchantID, got %v", err)
		}
	})

	t.Run("invalid price", func(t *testing.T) {
		svc := newProductsService(t)
		ctx := t.Context()

		_, err := svc.CreateProduct(ctx, "Laptop", "x", 0, uuid.New(), proto.Int32(5))
		if !errors.Is(err, service.ErrInvalidPrice) {
			t.Fatalf("expected ErrInvalidPrice, got %v", err)
		}
	})
}
