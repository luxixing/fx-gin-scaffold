package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

// Config holds all application configuration
type Config struct {
	App      AppConfig      `json:"app"`
	Database DatabaseConfig `json:"database"`
	JWT      JWTConfig      `json:"jwt"`
	Logger   LoggerConfig   `json:"logger"`
	Server   ServerConfig   `json:"server"`
}

// AppConfig contains general application settings
type AppConfig struct {
	Env   string `json:"env" env:"APP_ENV" envDefault:"development"`
	Debug bool   `json:"debug" env:"APP_DEBUG" envDefault:"false"`
}

// DatabaseConfig contains database connection settings
type DatabaseConfig struct {
	Driver     string `json:"driver" env:"DB_DRIVER" envDefault:"sqlite"`
	TablePrefix string `json:"table_prefix" env:"DB_TABLE_PREFIX" envDefault:"fx_"`

	// SQLite
	SQLitePath string `json:"sqlite_path" env:"SQLITE_PATH" envDefault:"./data/app.db"`

	// PostgreSQL
	PostgresHost     string `json:"postgres_host" env:"POSTGRES_HOST" envDefault:"localhost"`
	PostgresPort     int    `json:"postgres_port" env:"POSTGRES_PORT" envDefault:"5432"`
	PostgresUser     string `json:"postgres_user" env:"POSTGRES_USER" envDefault:"postgres"`
	PostgresPassword string `json:"postgres_password" env:"POSTGRES_PASSWORD" envDefault:""`
	PostgresDatabase string `json:"postgres_database" env:"POSTGRES_DATABASE" envDefault:"fx_gin_scaffold"`
	PostgresSSLMode  string `json:"postgres_sslmode" env:"POSTGRES_SSLMODE" envDefault:"disable"`
	PostgresTimezone string `json:"postgres_timezone" env:"POSTGRES_TIMEZONE" envDefault:"UTC"`

	// MongoDB
	MongoURI      string `json:"mongo_uri" env:"MONGO_URI" envDefault:"mongodb://localhost:27017"`
	MongoDatabase string `json:"mongo_database" env:"MONGO_DATABASE" envDefault:"fx_gin_scaffold"`
}

// JWTConfig contains JWT authentication settings
type JWTConfig struct {
	Secret     string        `json:"secret" env:"JWT_SECRET"`
	Expiration time.Duration `json:"expiration" env:"JWT_EXPIRATION" envDefault:"24h"`
}

// LoggerConfig contains logging configuration
type LoggerConfig struct {
	Level  string `json:"level" env:"LOG_LEVEL" envDefault:"info"`
	Format string `json:"format" env:"LOG_FORMAT" envDefault:"json"`
	Output string `json:"output" env:"LOG_OUTPUT" envDefault:"stdout"`
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Host string `json:"host" env:"APP_HOST" envDefault:"localhost"`
	Port int    `json:"port" env:"APP_PORT" envDefault:"8080"`

	// CORS
	EnableCORS  bool   `json:"enable_cors" env:"ENABLE_CORS" envDefault:"true"`
	CORSOrigins string `json:"cors_origins" env:"CORS_ORIGINS" envDefault:"*"`
	CORSMethods string `json:"cors_methods" env:"CORS_METHODS" envDefault:"GET,POST,PUT,DELETE,OPTIONS"`
	CORSHeaders string `json:"cors_headers" env:"CORS_HEADERS" envDefault:"Origin,Content-Type,Accept,Authorization,X-Requested-With"`

	// Documentation
	EnableSwagger bool `json:"enable_swagger" env:"ENABLE_SWAGGER" envDefault:"true"`
}

// NewConfig creates a new configuration instance
func NewConfig() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		zap.L().Debug("no .env file found, using environment variables only")
	}

	config := &Config{}

	// Parse environment variables using caarlos0/env
	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	// Validate required fields
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// validate checks if all required configuration fields are set
func (c *Config) validate() error {
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}

	if c.Database.Driver == "" {
		return fmt.Errorf("DB_DRIVER is required")
	}

	if strings.TrimSpace(c.Database.TablePrefix) == "" {
		return fmt.Errorf("DB_TABLE_PREFIX is required")
	}

	// Validate database driver
	switch c.Database.Driver {
	case "sqlite", "postgres", "mongo":
		// Valid drivers
	default:
		return fmt.Errorf("unsupported database driver: %s (supported: sqlite, postgres, mongo)", c.Database.Driver)
	}

	// Driver-specific validation
	switch c.Database.Driver {
	case "postgres":
		if c.Database.PostgresHost == "" {
			return fmt.Errorf("POSTGRES_HOST is required when using postgres driver")
		}
		if c.Database.PostgresUser == "" {
			return fmt.Errorf("POSTGRES_USER is required when using postgres driver")
		}
		if c.Database.PostgresDatabase == "" {
			return fmt.Errorf("POSTGRES_DATABASE is required when using postgres driver")
		}
	case "mongo":
		if c.Database.MongoURI == "" {
			return fmt.Errorf("MONGO_URI is required when using mongo driver")
		}
		if c.Database.MongoDatabase == "" {
			return fmt.Errorf("MONGO_DATABASE is required when using mongo driver")
		}
	}

	return nil
}

// IsDevelopment returns true if the app is running in development mode
func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

// IsProduction returns true if the app is running in production mode
func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}

// GetAddress returns the server address in host:port format
func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
