package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SQLiteConfig holds SQLite specific configuration
type SQLiteConfig struct {
	Path string `json:"path" yaml:"path"`
}

// PostgresConfig holds PostgreSQL specific configuration
type PostgresConfig struct {
	DSN  string `json:"dsn" yaml:"dsn"`
	Host string `json:"host" yaml:"host"`
	Port string `json:"port" yaml:"port"`
	User string `json:"user" yaml:"user"`
	Pass string `json:"pass" yaml:"pass"`
	DB   string `json:"db" yaml:"db"`
	SSL  string `json:"ssl" yaml:"ssl"`
}

// GetDSN returns PostgreSQL DSN, prioritizing explicit DSN over individual fields
func (c PostgresConfig) GetDSN() string {
	if c.DSN != "" {
		return c.DSN
	}
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Pass, c.DB, c.SSL)
}

// MongoConfig holds MongoDB specific configuration
type MongoConfig struct {
	URI string `json:"uri" yaml:"uri"`
}

// Config holds database configuration
type Config struct {
	Driver   string         `json:"driver" yaml:"driver"`
	SQLite   SQLiteConfig   `json:"sqlite" yaml:"sqlite"`
	Postgres PostgresConfig `json:"postgres" yaml:"postgres"`
	Mongo    MongoConfig    `json:"mongo" yaml:"mongo"`
}

// Connection holds database connections
type Connection struct {
	GORM  *gorm.DB
	Mongo *mongo.Client
}

// NewConnection creates database connections based on configuration
func NewConnection(cfg Config) (*Connection, error) {
	conn := &Connection{}

	switch cfg.Driver {
	case "sqlite":
		gormDB, err := connectSQLite(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to SQLite: %w", err)
		}
		conn.GORM = gormDB

	case "postgres":
		gormDB, err := connectPostgres(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
		}
		conn.GORM = gormDB

	case "mongo":
		mongoDB, err := connectMongo(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
		}
		conn.Mongo = mongoDB

	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}

	return conn, nil
}

// connectSQLite establishes SQLite connection
func connectSQLite(cfg Config) (*gorm.DB, error) {
	// Ensure directory exists
	dir := filepath.Dir(cfg.SQLite.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	db, err := gorm.Open(sqlite.Open(cfg.SQLite.Path), &gorm.Config{
		Logger: newGormLogger(),
	})
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// SQLite specific settings
	sqlDB.SetMaxOpenConns(1) // SQLite doesn't support concurrent writes
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// connectPostgres establishes PostgreSQL connection
func connectPostgres(cfg Config) (*gorm.DB, error) {
	dsn := cfg.Postgres.GetDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newGormLogger(),
	})
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

// connectMongo establishes MongoDB connection
func connectMongo(cfg Config) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.Mongo.URI)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Test the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// newGormLogger creates a custom GORM logger that integrates with zap
func newGormLogger() logger.Interface {
	return logger.New(
		&gormLogWriter{},
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
}

// gormLogWriter implements GORM's logger.Writer interface using zap
type gormLogWriter struct{}

func (w *gormLogWriter) Printf(format string, args ...any) {
	zap.L().Info(fmt.Sprintf(format, args...))
}

// Close gracefully closes database connections
func (c *Connection) Close() error {
	var errors []error

	if c.GORM != nil {
		sqlDB, err := c.GORM.DB()
		if err == nil {
			if err := sqlDB.Close(); err != nil {
				errors = append(errors, fmt.Errorf("failed to close GORM connection: %w", err))
			}
		}
	}

	if c.Mongo != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := c.Mongo.Disconnect(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to close MongoDB connection: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("database close errors: %v", errors)
	}

	return nil
}

// GetDatabase returns the appropriate database for the driver
func (c *Connection) GetDatabase(driver string) any {
	switch driver {
	case "sqlite", "postgres":
		return c.GORM
	case "mongo":
		return c.Mongo
	default:
		return nil
	}
}

// Health checks database connectivity
func (c *Connection) Health(ctx context.Context) error {
	if c.GORM != nil {
		sqlDB, err := c.GORM.DB()
		if err != nil {
			return err
		}
		return sqlDB.PingContext(ctx)
	}

	if c.Mongo != nil {
		return c.Mongo.Ping(ctx, nil)
	}

	return fmt.Errorf("no database connection available")
}
