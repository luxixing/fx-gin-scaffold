package service

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/luxixing/fx-gin-scaffold/internal/config"
	"github.com/luxixing/fx-gin-scaffold/internal/domain"
	"go.uber.org/fx"
)

// AuthServiceParams holds dependencies for AuthService
type AuthServiceParams struct {
	fx.In
	Config *config.Config
}

// authService implements domain.AuthService
type authService struct {
	config *config.Config
}

// NewAuthService creates a new auth service
func NewAuthService(p AuthServiceParams) domain.AuthService {
	return &authService{
		config: p.Config,
	}
}

// GenerateToken generates a JWT token for the user
func (s *authService) GenerateToken(user *domain.User) (string, error) {
	claims := &domain.JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.JWT.Expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "fx-gin-scaffold",
			Subject:   user.Email,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.JWT.Secret))
	if err != nil {
		return "", domain.WrapError(err, domain.ErrCodeInternal, "Failed to generate token")
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns claims
func (s *authService) ValidateToken(tokenString string) (*domain.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &domain.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, domain.NewError(domain.ErrCodeInvalidToken, "Invalid signing method")
		}
		return []byte(s.config.JWT.Secret), nil
	})

	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	if !token.Valid {
		return nil, domain.ErrInvalidToken
	}

	claims, ok := token.Claims.(*domain.JWTClaims)
	if !ok {
		return nil, domain.ErrInvalidToken
	}

	return claims, nil
}

// RefreshToken refreshes an existing token
func (s *authService) RefreshToken(ctx context.Context, tokenString string) (string, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Check if token is close to expiration (within 1 hour)
	if time.Until(claims.ExpiresAt.Time) > time.Hour {
		return "", domain.NewError(domain.ErrCodeInvalid, "Token is not close to expiration")
	}

	// Create new token with updated expiration
	newClaims := &domain.JWTClaims{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.JWT.Expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "fx-gin-scaffold",
			Subject:   claims.Email,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	newTokenString, err := token.SignedString([]byte(s.config.JWT.Secret))
	if err != nil {
		return "", domain.WrapError(err, domain.ErrCodeInternal, "Failed to refresh token")
	}

	return newTokenString, nil
}