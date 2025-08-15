package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/luxixing/fx-gin-scaffold/internal/domain"
	"go.uber.org/fx"
)

// JWTMiddlewareParams holds dependencies for JWT middleware
type JWTMiddlewareParams struct {
	fx.In
	AuthService domain.AuthService
}

// JWTMiddleware handles JWT authentication
type JWTMiddleware struct {
	authService domain.AuthService
}

// NewJWTMiddleware creates a new JWT middleware
func NewJWTMiddleware(p JWTMiddlewareParams) *JWTMiddleware {
	return &JWTMiddleware{
		authService: p.AuthService,
	}
}

// RequireAuth middleware that requires valid JWT token
func (m *JWTMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, domain.NewErrorResponse(domain.ErrUnauthorized))
			c.Abort()
			return
		}

		claims, err := m.authService.ValidateToken(token)
		if err != nil {
			if domainErr, ok := err.(*domain.Error); ok {
				c.JSON(http.StatusUnauthorized, domain.NewErrorResponse(domainErr))
			} else {
				c.JSON(http.StatusUnauthorized, domain.NewErrorResponse(domain.ErrInvalidToken))
			}
			c.Abort()
			return
		}

		// Set user information in context
		c.Set(string(domain.UserIDContextKey), claims.UserID)
		c.Set(string(domain.UserContextKey), claims.Email)
		c.Set(string(domain.RoleContextKey), claims.Role)
		
		c.Next()
	}
}

// RequireAdmin middleware that requires admin role
func (m *JWTMiddleware) RequireAdmin() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// First check if user is authenticated
		m.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		// Check if user has admin role
		role, exists := c.Get(string(domain.RoleContextKey))
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, domain.NewErrorResponse(domain.ErrForbidden))
			c.Abort()
			return
		}

		c.Next()
	})
}

// OptionalAuth middleware that optionally validates JWT token
func (m *JWTMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.Next()
			return
		}

		claims, err := m.authService.ValidateToken(token)
		if err != nil {
			// Continue without setting user context for optional auth
			c.Next()
			return
		}

		// Set user information in context
		c.Set(string(domain.UserIDContextKey), claims.UserID)
		c.Set(string(domain.UserContextKey), claims.Email)
		c.Set(string(domain.RoleContextKey), claims.Role)
		
		c.Next()
	}
}

// extractToken extracts JWT token from Authorization header
func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check for Bearer token
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return ""
}

// GetUserID extracts user ID from gin context
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get(string(domain.UserIDContextKey))
	if !exists {
		return 0, false
	}
	
	id, ok := userID.(uint)
	return id, ok
}

// GetUserEmail extracts user email from gin context
func GetUserEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get(string(domain.UserContextKey))
	if !exists {
		return "", false
	}
	
	emailStr, ok := email.(string)
	return emailStr, ok
}

// GetUserRole extracts user role from gin context
func GetUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get(string(domain.RoleContextKey))
	if !exists {
		return "", false
	}
	
	roleStr, ok := role.(string)
	return roleStr, ok
}