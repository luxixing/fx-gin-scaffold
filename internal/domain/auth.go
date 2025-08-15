package domain

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents JWT claims
type JWTClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Token string        `json:"token"`
	User  *UserResponse `json:"user"`
}

// AuthService defines the interface for authentication operations
type AuthService interface {
	// GenerateToken generates a JWT token for the user
	GenerateToken(user *User) (string, error)
	
	// ValidateToken validates a JWT token and returns claims
	ValidateToken(tokenString string) (*JWTClaims, error)
	
	// RefreshToken refreshes an existing token
	RefreshToken(ctx context.Context, tokenString string) (string, error)
}

// ContextKey represents context keys
type ContextKey string

const (
	// UserContextKey is the key for user in context
	UserContextKey ContextKey = "user"
	
	// UserIDContextKey is the key for user ID in context
	UserIDContextKey ContextKey = "user_id"
	
	// RoleContextKey is the key for user role in context
	RoleContextKey ContextKey = "role"
)