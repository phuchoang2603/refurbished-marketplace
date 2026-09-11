package auth

import "context"

type contextKey string

const (
	userIDKey    contextKey = "authUserID"
	userEmailKey contextKey = "authUserEmail"
)

func UserIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userIDKey).(string)
	return v, ok
}

func EmailFromContext(ctx context.Context) string {
	v, _ := ctx.Value(userEmailKey).(string)
	return v
}

func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func ContextWithEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, userEmailKey, email)
}
