package migration

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/luxixing/fx-gin-scaffold/pkg/database"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// Migration represents a single database migration
type Migration interface {
	// Version returns the migration version (timestamp format: 20240101120000)
	Version() string
	
	// Description returns a human-readable description of the migration
	Description() string
	
	// Up executes the migration
	Up(ctx context.Context, db *database.Connection) error
	
	// Down rolls back the migration (optional, can return nil for irreversible migrations)
	Down(ctx context.Context, db *database.Connection) error
}

// Seeder represents a data seeder
type Seeder interface {
	// Name returns the seeder name
	Name() string
	
	// Run executes the seeder
	Run(ctx context.Context, db *database.Connection) error
	
	// ShouldRun determines if the seeder should run (e.g., only in development)
	ShouldRun(env string) bool
}

// Migrator handles migration execution
type Migrator struct {
	db         *database.Connection
	migrations []Migration
	seeders    []Seeder
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *database.Connection) *Migrator {
	return &Migrator{
		db:         db,
		migrations: make([]Migration, 0),
		seeders:    make([]Seeder, 0),
	}
}

// AddMigration adds a migration to the migrator
func (m *Migrator) AddMigration(migration Migration) {
	m.migrations = append(m.migrations, migration)
}

// AddSeeder adds a seeder to the migrator
func (m *Migrator) AddSeeder(seeder Seeder) {
	m.seeders = append(m.seeders, seeder)
}

// GetMigrations returns all registered migrations
func (m *Migrator) GetMigrations() []Migration {
	return m.migrations
}

// GetSeeders returns all registered seeders
func (m *Migrator) GetSeeders() []Seeder {
	return m.seeders
}

// EnsureMigrationTracking creates the migration tracking table/collection
func (m *Migrator) EnsureMigrationTracking(ctx context.Context) error {
	return m.ensureMigrationTracking(ctx)
}

// GetExecutedMigrations returns a map of executed migration versions
func (m *Migrator) GetExecutedMigrations(ctx context.Context) (map[string]bool, error) {
	return m.getExecutedMigrations(ctx)
}

// Migrate runs all pending migrations
func (m *Migrator) Migrate(ctx context.Context) error {
	// Sort migrations by version
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version() < m.migrations[j].Version()
	})

	// Create migration tracking table/collection if it doesn't exist
	if err := m.ensureMigrationTracking(ctx); err != nil {
		return fmt.Errorf("failed to create migration tracking: %w", err)
	}

	// Get already executed migrations
	executed, err := m.getExecutedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get executed migrations: %w", err)
	}

	// Run pending migrations
	for _, migration := range m.migrations {
		if _, exists := executed[migration.Version()]; exists {
			zap.L().Debug("migration already executed", 
				zap.String("version", migration.Version()),
				zap.String("description", migration.Description()))
			continue
		}

		zap.L().Info("running migration", 
			zap.String("version", migration.Version()),
			zap.String("description", migration.Description()))

		if err := migration.Up(ctx, m.db); err != nil {
			return fmt.Errorf("migration %s failed: %w", migration.Version(), err)
		}

		if err := m.recordMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migration.Version(), err)
		}

		zap.L().Info("migration completed", 
			zap.String("version", migration.Version()))
	}

	return nil
}

// Seed runs all applicable seeders
func (m *Migrator) Seed(ctx context.Context, env string) error {
	for _, seeder := range m.seeders {
		if !seeder.ShouldRun(env) {
			zap.L().Debug("skipping seeder", 
				zap.String("name", seeder.Name()),
				zap.String("env", env))
			continue
		}

		zap.L().Info("running seeder", zap.String("name", seeder.Name()))

		if err := seeder.Run(ctx, m.db); err != nil {
			return fmt.Errorf("seeder %s failed: %w", seeder.Name(), err)
		}

		zap.L().Info("seeder completed", zap.String("name", seeder.Name()))
	}

	return nil
}

// ensureMigrationTracking creates the migration tracking table/collection
func (m *Migrator) ensureMigrationTracking(ctx context.Context) error {
	if m.db.GORM != nil {
		// SQL databases - create migrations table
		return m.db.GORM.Exec(`
			CREATE TABLE IF NOT EXISTS migrations (
				version VARCHAR(255) PRIMARY KEY,
				description TEXT,
				executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`).Error
	}

	if m.db.Mongo != nil {
		// MongoDB - ensure migrations collection exists (it will be created automatically)
		// We can optionally create indexes here
		collection := m.db.Mongo.Database("fx_gin_scaffold").Collection("migrations")
		indexModel := mongo.IndexModel{
			Keys: map[string]interface{}{"version": 1},
			Options: options.Index().
				SetUnique(true).
				SetName("idx_migrations_version"),
		}
		_, err := collection.Indexes().CreateOne(ctx, indexModel)
		return err
	}

	return fmt.Errorf("no database connection available")
}

// getExecutedMigrations returns a map of executed migration versions
func (m *Migrator) getExecutedMigrations(ctx context.Context) (map[string]bool, error) {
	executed := make(map[string]bool)

	if m.db.GORM != nil {
		// SQL databases
		var versions []string
		if err := m.db.GORM.Raw("SELECT version FROM migrations").Scan(&versions).Error; err != nil {
			return nil, err
		}
		for _, version := range versions {
			executed[version] = true
		}
		return executed, nil
	}

	if m.db.Mongo != nil {
		// MongoDB
		collection := m.db.Mongo.Database("fx_gin_scaffold").Collection("migrations")
		cursor, err := collection.Find(ctx, map[string]interface{}{})
		if err != nil {
			return nil, err
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var doc map[string]interface{}
			if err := cursor.Decode(&doc); err != nil {
				return nil, err
			}
			if version, ok := doc["version"].(string); ok {
				executed[version] = true
			}
		}
		return executed, cursor.Err()
	}

	return nil, fmt.Errorf("no database connection available")
}

// recordMigration records a completed migration
func (m *Migrator) recordMigration(ctx context.Context, migration Migration) error {
	if m.db.GORM != nil {
		// SQL databases
		return m.db.GORM.Exec(
			"INSERT INTO migrations (version, description) VALUES (?, ?)",
			migration.Version(),
			migration.Description(),
		).Error
	}

	if m.db.Mongo != nil {
		// MongoDB
		collection := m.db.Mongo.Database("fx_gin_scaffold").Collection("migrations")
		_, err := collection.InsertOne(ctx, map[string]interface{}{
			"version":     migration.Version(),
			"description": migration.Description(),
			"executed_at": time.Now(),
		})
		return err
	}

	return fmt.Errorf("no database connection available")
}