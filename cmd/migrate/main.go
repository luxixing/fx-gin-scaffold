package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/luxixing/fx-gin-scaffold/internal/config"
	"github.com/luxixing/fx-gin-scaffold/internal/domain"
	"github.com/luxixing/fx-gin-scaffold/internal/migration"
	"github.com/luxixing/fx-gin-scaffold/pkg/database"
	"github.com/luxixing/fx-gin-scaffold/pkg/logger"
)

func main() {
	// Parse command line flags
	var (
		checkOnly = flag.Bool("check", false, "Check pending migrations without running them")
		dryRun    = flag.Bool("dry-run", false, "Show what migrations would be executed")
	)
	flag.Parse()

	fmt.Println("🔄 Loading configuration...")
	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger (duplicated from bootstrap for independence)
	err = logger.Initialize(logger.Config{
		Level:  cfg.Logger.Level,
		Format: cfg.Logger.Format,
		Output: cfg.Logger.Output,
	})
	if err != nil {
		fmt.Printf("⚠️  Failed to initialize logger: %v\n", err)
		// Continue anyway
	}

	fmt.Println("🔗 Connecting to database...")
	
	// Set table prefix for domain models (duplicated from bootstrap)
	domain.SetTablePrefix(cfg.Database.TablePrefix)
	
	dbConfig := database.Config{
		Driver: cfg.Database.Driver,
		SQLite: database.SQLiteConfig{
			Path: cfg.Database.SQLitePath,
		},
		// Add other database configs when needed (commented out for now)
		// Postgres: database.PostgresConfig{
		//     Host: cfg.Database.PostgresHost,
		//     Port: fmt.Sprintf("%d", cfg.Database.PostgresPort),
		//     User: cfg.Database.PostgresUser,
		//     Pass: cfg.Database.PostgresPassword,
		//     DB:   cfg.Database.PostgresDatabase,
		//     SSL:  cfg.Database.PostgresSSLMode,
		// },
		// Mongo: database.MongoConfig{
		//     URI: cfg.Database.MongoURI,
		// },
	}
	
	db, err := database.NewConnection(dbConfig)
	if err != nil {
		fmt.Printf("❌ Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	ctx := context.Background()

	if *checkOnly {
		fmt.Println("🔍 Checking pending migrations...")
		if err := checkPendingMigrations(ctx, db); err != nil {
			fmt.Printf("❌ Check failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *dryRun {
		fmt.Println("🧪 Dry run - showing what would be executed...")
		if err := showPendingMigrations(ctx, db); err != nil {
			fmt.Printf("❌ Dry run failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	fmt.Println("🚀 Running migrations...")
	if err := migration.RunMigrations(ctx, db, cfg.App.Env); err != nil {
		fmt.Printf("❌ Migration failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Migrations completed successfully")
}

// checkPendingMigrations checks if there are pending migrations
func checkPendingMigrations(ctx context.Context, db *database.Connection) error {
	migrator := migration.NewMigrator(db)
	migration.RegisterMigrations(migrator)
	
	// Create migration tracking if it doesn't exist
	if err := migrator.EnsureMigrationTracking(ctx); err != nil {
		return err
	}

	executed, err := migrator.GetExecutedMigrations(ctx)
	if err != nil {
		return err
	}

	pending := 0
	for _, mig := range migrator.GetMigrations() {
		if _, exists := executed[mig.Version()]; !exists {
			pending++
			fmt.Printf("📋 Pending: %s - %s\n", mig.Version(), mig.Description())
		}
	}

	if pending == 0 {
		fmt.Println("✅ No pending migrations")
	} else {
		fmt.Printf("⚠️  Found %d pending migration(s)\n", pending)
	}

	return nil
}

// showPendingMigrations shows what migrations would be executed
func showPendingMigrations(ctx context.Context, db *database.Connection) error {
	migrator := migration.NewMigrator(db)
	migration.RegisterMigrations(migrator)
	migration.RegisterSeeders(migrator)

	fmt.Println("📋 Migrations that would be executed:")
	if err := checkPendingMigrations(ctx, db); err != nil {
		return err
	}

	fmt.Println("\n🌱 Seeders that would be executed:")
	for _, seeder := range migrator.GetSeeders() {
		env := os.Getenv("APP_ENV")
		if env == "" {
			env = "development"
		}
		if seeder.ShouldRun(env) {
			fmt.Printf("🌱 Would run: %s\n", seeder.Name())
		} else {
			fmt.Printf("⏭️  Would skip: %s (not for %s environment)\n", seeder.Name(), env)
		}
	}

	return nil
}