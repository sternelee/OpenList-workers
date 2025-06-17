package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sternelee/OpenList-workers/workers/models"
	"github.com/golang-jwt/jwt/v4"
)

const (
	DefaultJWTSecret     = "openlist-default-secret-please-change-in-production"
	DefaultTokenDuration = 24 * time.Hour
)

type JWTClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     int    `json:"role"`
	jwt.RegisteredClaims
}

type JWTAuth struct {
	secret []byte
}

func NewJWTAuth(secret string) *JWTAuth {
	if secret == "" {
		secret = DefaultJWTSecret
	}
	return &JWTAuth{
		secret: []byte(secret),
	}
}

func (j *JWTAuth) GenerateToken(user *models.User) (string, error) {
	claims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(DefaultTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "openlist-workers",
			Subject:   user.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWTAuth) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func ExtractTokenFromHeader(authHeader string) string {
	if authHeader == "" {
		return ""
	}

	// Check for "Bearer " prefix
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Also support direct token without Bearer prefix
	return authHeader
}

type contextKey string

const UserContextKey contextKey = "user"

func SetUserInContext(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}

func GetUserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	return user, ok
} 