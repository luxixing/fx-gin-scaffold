package bootstrap

import (
	"context"
	"net/http"

	"github.com/luxixing/fx-gin-scaffold/internal/config"
	"github.com/luxixing/fx-gin-scaffold/internal/domain"
	"github.com/luxixing/fx-gin-scaffold/internal/http/handler"
	"github.com/luxixing/fx-gin-scaffold/internal/http/middleware"
	"github.com/luxixing/fx-gin-scaffold/internal/repo"
	"github.com/luxixing/fx-gin-scaffold/internal/service"
	"github.com/luxixing/fx-gin-scaffold/pkg/database"
	"github.com/luxixing/fx-gin-scaffold/pkg/logger"
	"go.uber.org/fx"
	"go.uber.org/zap"
)


// GetModule returns the complete fx.Option for the entire application
func GetModule() fx.Option {
	return fx.Options(
		// Configuration and Infrastructure
		fx.Provide(config.NewConfig),
		fx.Provide(initializeLogger),
		fx.Provide(initializeDatabase),

		// Repositories
		fx.Provide(
			fx.Annotate(
				repo.NewUserRepository,
				fx.As(new(domain.UserRepository)),
			),
		),

		// Services
		service.GetModule(),

		// Middleware
		fx.Provide(middleware.NewJWTMiddleware),

		// Handlers
		fx.Provide(handler.NewAuthHandler),
		fx.Provide(handler.NewUserHandler),

		// HTTP server
		fx.Provide(NewHTTPServer),
	)
}

// RegisterHooks registers application lifecycle hooks
func RegisterHooks(lc fx.Lifecycle, cfg *config.Config, db *database.Connection, server *http.Server) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return onStart(ctx, cfg, db, server)
		},
		OnStop: func(ctx context.Context) error {
			return onStop(ctx, db, server)
		},
	})
}


// initializeLogger initializes the logger based on configuration
func initializeLogger(cfg *config.Config) (bool, error) {
	err := logger.Initialize(logger.Config{
		Level:  cfg.Logger.Level,
		Format: cfg.Logger.Format,
		Output: cfg.Logger.Output,
	})
	return true, err // Return a dummy bool value for FX
}

// initializeDatabase creates database connection based on configuration
func initializeDatabase(cfg *config.Config) (*database.Connection, error) {
	// Set table prefix for all domain models
	domain.SetTablePrefix(cfg.Database.TablePrefix)

	dbConfig := database.Config{
		Driver: cfg.Database.Driver,
		SQLite: database.SQLiteConfig{
			Path: cfg.Database.SQLitePath,
		},
		// TODO: Add PostgreSQL configuration when needed
		// Postgres: database.PostgresConfig{
		//     Host: cfg.Database.PostgresHost,
		//     Port: fmt.Sprintf("%d", cfg.Database.PostgresPort),
		//     User: cfg.Database.PostgresUser,
		//     Pass: cfg.Database.PostgresPassword,
		//     DB:   cfg.Database.PostgresDatabase,
		//     SSL:  cfg.Database.PostgresSSLMode,
		// },
		// TODO: Add MongoDB configuration when needed
		// Mongo: database.MongoConfig{
		//     URI: cfg.Database.MongoURI,
		// },
	}
	return database.NewConnection(dbConfig)
}

// onStart handles application startup
func onStart(ctx context.Context, cfg *config.Config, db *database.Connection, server *http.Server) error {
	zap.L().Info("starting application",
		zap.String("env", cfg.App.Env),
		zap.String("address", cfg.GetAddress()),
	)


	// Start HTTP server in a goroutine
	go func() {
		zap.L().Info("http server starting", zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("http server failed to start", zap.Error(err))
		}
	}()

	return nil
}

// onStop handles application shutdown
func onStop(ctx context.Context, db *database.Connection, server *http.Server) error {
	zap.L().Info("stopping application")

	// Shutdown HTTP server gracefully
	if err := server.Shutdown(ctx); err != nil {
		zap.L().Error("error shutting down http server", zap.Error(err))
		return err
	}
	zap.L().Info("http server stopped")

	// Close database connections
	if err := db.Close(); err != nil {
		zap.L().Error("error closing database connections", zap.Error(err))
		return err
	}
	zap.L().Info("database connections closed")

	// Sync logger before exit
	logger.Sync()

	return nil
}
