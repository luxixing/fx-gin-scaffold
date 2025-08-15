package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// ExtractToken extracts JWT token from Authorization header (utility function)
func ExtractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check for Bearer token
	if after, ok :=strings.CutPrefix(authHeader, "Bearer "); ok  {
		return after
	}

	return ""
}
