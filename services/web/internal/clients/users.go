package clients

import (
	"context"

	usersv1 "refurbished-marketplace/shared/proto/users/v1"

	"google.golang.org/grpc"
)

type UsersClient struct {
	conn   *grpc.ClientConn
	client usersv1.UsersServiceClient
}

func newUsersClient(addr string) (*UsersClient, error) {
	conn, err := newConn(addr)
	if err != nil {
		return nil, err
	}
	return &UsersClient{conn: conn, client: usersv1.NewUsersServiceClient(conn)}, nil
}

func (c *UsersClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *UsersClient) Login(ctx context.Context, email, password string) (*usersv1.TokenResponse, error) {
	return c.client.Login(ctx, &usersv1.LoginRequest{Email: email, Password: password})
}

func (c *UsersClient) Refresh(ctx context.Context, refreshToken string) (*usersv1.TokenResponse, error) {
	return c.client.Refresh(ctx, &usersv1.RefreshRequest{RefreshToken: refreshToken})
}

func (c *UsersClient) Logout(ctx context.Context, refreshToken string) (*usersv1.LogoutResponse, error) {
	return c.client.Logout(ctx, &usersv1.LogoutRequest{RefreshToken: refreshToken})
}

func (c *UsersClient) CreateUser(ctx context.Context, email, password string) (*usersv1.User, error) {
	return c.client.CreateUser(ctx, &usersv1.CreateUserRequest{Email: email, Password: password})
}
