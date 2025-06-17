package handlers

import (
	"context"

	"github.com/sternelee/OpenList-workers/workers/models"
)

type contextKey string

const userContextKey contextKey = "user"

// SetUserInContext 将用户信息添加到上下文
func SetUserInContext(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// GetUserFromContext 从上下文获取用户信息
func GetUserFromContext(ctx context.Context) *models.User {
	if user, ok := ctx.Value(userContextKey).(*models.User); ok {
		return user
	}
	return nil
}

