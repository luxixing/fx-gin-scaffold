package migration

import (
	"context"

	"github.com/luxixing/fx-gin-scaffold/internal/migration/migrations"
	"github.com/luxixing/fx-gin-scaffold/internal/migration/seeders"
	"github.com/luxixing/fx-gin-scaffold/pkg/database"
)

// RegisterMigrations registers all migrations
func RegisterMigrations(migrator *Migrator) {
	// Add all migrations here in chronological order
	migrator.AddMigration(&migrations.CreateUsersTable{})
}

// RegisterSeeders registers all seeders
func RegisterSeeders(migrator *Migrator) {
	// Add all seeders here
	migrator.AddSeeder(&seeders.AdminUserSeeder{})
	migrator.AddSeeder(&seeders.TestUsersSeeder{})
}

// RunMigrations runs all migrations and seeders
func RunMigrations(ctx context.Context, db *database.Connection, env string) error {
	migrator := NewMigrator(db)
	
	// Register migrations and seeders
	RegisterMigrations(migrator)
	RegisterSeeders(migrator)
	
	// Run migrations first
	if err := migrator.Migrate(ctx); err != nil {
		return err
	}
	
	// Then run seeders
	return migrator.Seed(ctx, env)
}