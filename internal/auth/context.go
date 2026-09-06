package auth

import (
	"context"
)

type contextKey string

const userIDKey contextKey = "userID"

// returns new context with userID
func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// get userID from context request
func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDKey).(int64)
	if !ok {
		return 0, false
	}

	if userID <= 0 {
		return 0, false
	}

	return userID, true
}
