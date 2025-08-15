package bootstrap

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luxixing/fx-gin-scaffold/internal/config"
	"github.com/luxixing/fx-gin-scaffold/internal/http/handler"
	"github.com/luxixing/fx-gin-scaffold/internal/http/middleware"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/fx"
)

// HTTPServerParams holds dependencies for HTTP server
type HTTPServerParams struct {
	fx.In
	Config        *config.Config
	AuthHandler   *handler.AuthHandler
	UserHandler   *handler.UserHandler
	JWTMiddleware *middleware.JWTMiddleware
}

// NewHTTPServer creates a new HTTP server with Gin
func NewHTTPServer(p HTTPServerParams) *http.Server {
	cfg := p.Config
	// Set Gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS
	if cfg.Server.EnableCORS {
		router.Use(corsMiddleware(cfg))
	}

	// Health check
	router.GET("/health", healthCheck)

	// Swagger documentation
	if cfg.Server.EnableSwagger {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/register", p.AuthHandler.Register)
			auth.POST("/login", p.AuthHandler.Login)
			auth.POST("/refresh", p.JWTMiddleware.RequireAuth(), p.AuthHandler.RefreshToken)
			auth.GET("/profile", p.JWTMiddleware.RequireAuth(), p.AuthHandler.GetProfile)
			auth.PUT("/profile", p.JWTMiddleware.RequireAuth(), p.AuthHandler.UpdateProfile)
		}

		// User management routes (admin only)
		users := v1.Group("/users", p.JWTMiddleware.RequireAdmin())
		{
			users.GET("", p.UserHandler.ListUsers)
			users.GET("/search", p.UserHandler.SearchUsers)
			users.GET("/:id", p.UserHandler.GetUser)
			users.PUT("/:id", p.UserHandler.UpdateUser)
			users.DELETE("/:id", p.UserHandler.DeleteUser)
		}
	}

	return &http.Server{
		Addr:         cfg.GetAddress(),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

// corsMiddleware configures CORS
func corsMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", cfg.Server.CORSOrigins)
		c.Header("Access-Control-Allow-Methods", cfg.Server.CORSMethods)
		c.Header("Access-Control-Allow-Headers", cfg.Server.CORSHeaders)
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

// healthCheck provides a simple health check endpoint
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().UTC(),
	})
}