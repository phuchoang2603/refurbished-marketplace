package tests

import (
	"net"
	"testing"

	"refurbished-marketplace/services/products/internal/database"
	"refurbished-marketplace/services/products/internal/grpcserver"
	"refurbished-marketplace/services/products/internal/service"
	productsv1 "refurbished-marketplace/shared/proto/products/v1"
	"refurbished-marketplace/shared/testutil"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func TestProductsGRPCFlow(t *testing.T) {
	db := testutil.SetupPostgresWithMigrations(
		t,
		testutil.PostgresConfig{
			Database: "products_db",
			Username: "products_app",
			Password: "products_app_dev_password",
		},
		"../db/migrations",
	)

	svc := service.New(database.New(db))
	server := grpc.NewServer()
	productsv1.RegisterProductsServiceServer(server, grpcserver.New(svc))

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	go func() {
		_ = server.Serve(lis)
	}()

	t.Cleanup(func() {
		server.Stop()
		_ = lis.Close()
	})

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	client := productsv1.NewProductsServiceClient(conn)

	_, err = client.CreateProduct(t.Context(), &productsv1.CreateProductRequest{
		Name:        "MacBook Air M1",
		Description: "Refurbished laptop",
		PriceCents:  74900,
		Stock:       4,
	})
	if err != nil {
		t.Fatalf("create product: %v", err)
	}

	list, err := client.ListProducts(t.Context(), &productsv1.ListProductsRequest{Limit: 20, Offset: 0})
	if err != nil {
		t.Fatalf("list products: %v", err)
	}

	if len(list.Products) != 1 {
		t.Fatalf("expected 1 product, got %d", len(list.Products))
	}

	_, err = client.CreateProduct(t.Context(), &productsv1.CreateProductRequest{
		Name:       "",
		PriceCents: 100,
		Stock:      1,
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", status.Code(err))
	}

	_, err = client.GetProductByID(t.Context(), &productsv1.GetProductByIDRequest{Id: "not-a-uuid"})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for malformed id, got %v", status.Code(err))
	}
}
