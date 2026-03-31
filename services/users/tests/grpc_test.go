package tests

import (
	"net"
	"testing"

	"refurbished-marketplace/services/users/internal/database"
	"refurbished-marketplace/services/users/internal/grpcserver"
	"refurbished-marketplace/services/users/internal/service"
	usersv1 "refurbished-marketplace/shared/proto/users/v1"
	"refurbished-marketplace/shared/testutil"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func TestUsersGRPCFlow(t *testing.T) {
	db := testutil.SetupPostgresWithMigrations(
		t,
		testutil.PostgresConfig{
			Database: "users_db",
			Username: "users_app",
			Password: "users_app_dev_password",
		},
		"../db/migrations",
	)

	queries := database.New(db)
	cfg := service.DefaultConfig("test-secret")
	svc := service.New(queries, cfg)

	server := grpc.NewServer()
	usersv1.RegisterUsersServiceServer(server, grpcserver.New(svc))

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

	client := usersv1.NewUsersServiceClient(conn)

	_, err = client.CreateUser(t.Context(), &usersv1.CreateUserRequest{
		Email:    "grpc-user@test.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	login, err := client.Login(t.Context(), &usersv1.LoginRequest{
		Email:    "grpc-user@test.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	_, err = client.Login(t.Context(), &usersv1.LoginRequest{
		Email:    "grpc-user@test.com",
		Password: "wrong",
	})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected Unauthenticated for invalid credentials, got %v", status.Code(err))
	}

	_, err = client.GetUserByID(t.Context(), &usersv1.GetUserByIDRequest{Id: "not-a-uuid"})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument for malformed id, got %v", status.Code(err))
	}

	if login.AccessToken == "" || login.RefreshToken == "" {
		t.Fatalf("expected non-empty tokens")
	}
}
