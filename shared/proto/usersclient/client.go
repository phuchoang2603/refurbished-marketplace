package usersclient

import (
	"context"

	usersv1 "refurbished-marketplace/shared/proto/users/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn   *grpc.ClientConn
	client usersv1.UsersServiceClient
}

func New(addr string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:   conn,
		client: usersv1.NewUsersServiceClient(conn),
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) Login(ctx context.Context, email, password string) (*usersv1.TokenResponse, error) {
	return c.client.Login(ctx, &usersv1.LoginRequest{Email: email, Password: password})
}

func (c *Client) Refresh(ctx context.Context, refreshToken string) (*usersv1.TokenResponse, error) {
	return c.client.Refresh(ctx, &usersv1.RefreshRequest{RefreshToken: refreshToken})
}

func (c *Client) Logout(ctx context.Context, refreshToken string) (*usersv1.LogoutResponse, error) {
	return c.client.Logout(ctx, &usersv1.LogoutRequest{RefreshToken: refreshToken})
}

func (c *Client) CreateUser(ctx context.Context, email, password string, xPos, yPos float64) (*usersv1.User, error) {
	return c.client.CreateUser(ctx, &usersv1.CreateUserRequest{Email: email, Password: password, XPos: xPos, YPos: yPos})
}

func (c *Client) GetUserByID(ctx context.Context, id string) (*usersv1.User, error) {
	return c.client.GetUserByID(ctx, &usersv1.GetUserByIDRequest{Id: id})
}
