package requestctx

import "context"

type Key string

const (
	UserKey   Key = "user"
	UserIDKey Key = "user_id"
)

// WithUserID stores the authenticated user ID on a request context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// UserID extracts the authenticated user ID from a context.
func UserID(ctx context.Context) string {
	id, _ := ctx.Value(UserIDKey).(string)
	return id
}
