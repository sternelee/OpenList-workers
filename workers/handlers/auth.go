package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/OpenListTeam/OpenList-workers/workers/auth"
	"github.com/OpenListTeam/OpenList-workers/workers/db"
	"github.com/OpenListTeam/OpenList-workers/workers/models"
)

type AuthHandler struct {
	db   *db.D1Client
	jwt  *auth.JWTAuth
}

func NewAuthHandler(db *db.D1Client, jwt *auth.JWTAuth) *AuthHandler {
	return &AuthHandler{
		db:  db,
		jwt: jwt,
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	OtpCode  string `json:"otp_code,omitempty"`
}

type LoginResponse struct {
	Token string           `json:"token"`
	User  *UserResponse    `json:"user"`
}

type UserResponse struct {
	ID         int       `json:"id"`
	Username   string    `json:"username"`
	Role       int       `json:"role"`
	Permission int32     `json:"permission"`
	BasePath   string    `json:"base_path"`
	Disabled   bool      `json:"disabled"`
	SsoID      string    `json:"sso_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	HasOtp     bool      `json:"has_otp"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		h.errorResponse(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	// Get user from database
	user, err := h.db.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		h.errorResponse(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Check if user is disabled
	if user.Disabled {
		h.errorResponse(w, "User account is disabled", http.StatusUnauthorized)
		return
	}

	// Validate password
	if err := user.ValidateRawPassword(req.Password); err != nil {
		h.errorResponse(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Check 2FA if enabled
	if user.OtpSecret != "" {
		if req.OtpCode == "" {
			h.errorResponse(w, "2FA code is required", http.StatusUnauthorized)
			return
		}
		// TODO: Implement TOTP validation
		// For now, we'll skip 2FA validation
	}

	// Generate JWT token
	token, err := h.jwt.GenerateToken(user)
	if err != nil {
		h.errorResponse(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Return success response
	response := LoginResponse{
		Token: token,
		User:  h.userToResponse(user),
	}

	h.jsonResponse(w, response, http.StatusOK)
}

func (h *AuthHandler) CurrentUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		h.errorResponse(w, "User not found in context", http.StatusUnauthorized)
		return
	}

	h.jsonResponse(w, h.userToResponse(user), http.StatusOK)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// For stateless JWT, we just return success
	// In a real implementation, you might want to blacklist the token
	h.jsonResponse(w, map[string]string{"message": "Logged out successfully"}, http.StatusOK)
}

func (h *AuthHandler) userToResponse(user *models.User) *UserResponse {
	return &UserResponse{
		ID:         user.ID,
		Username:   user.Username,
		Role:       user.Role,
		Permission: user.Permission,
		BasePath:   user.BasePath,
		Disabled:   user.Disabled,
		SsoID:      user.SsoID,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
		HasOtp:     user.OtpSecret != "",
	}
}

func (h *AuthHandler) jsonResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func (h *AuthHandler) errorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// Middleware for JWT authentication
func (h *AuthHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		token := auth.ExtractTokenFromHeader(authHeader)

		if token == "" {
			h.errorResponse(w, "Authorization token required", http.StatusUnauthorized)
			return
		}

		claims, err := h.jwt.ValidateToken(token)
		if err != nil {
			h.errorResponse(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Get user from database to ensure it still exists and is not disabled
		user, err := h.db.GetUserByID(r.Context(), claims.UserID)
		if err != nil {
			h.errorResponse(w, "User not found", http.StatusUnauthorized)
			return
		}

		if user.Disabled {
			h.errorResponse(w, "User account is disabled", http.StatusUnauthorized)
			return
		}

		// Add user to context
		ctx := auth.SetUserInContext(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Middleware for admin-only access
func (h *AuthHandler) AdminOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := auth.GetUserFromContext(r.Context())
		if !ok {
			h.errorResponse(w, "User not found in context", http.StatusUnauthorized)
			return
		}

		if !user.IsAdmin() {
			h.errorResponse(w, "Admin access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
} 