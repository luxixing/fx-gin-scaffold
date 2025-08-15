package repo

import (
	"strings"

	"github.com/luxixing/fx-gin-scaffold/internal/config"
	"github.com/luxixing/fx-gin-scaffold/internal/domain"
	"github.com/luxixing/fx-gin-scaffold/pkg/database"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/fx"
)

// RepositoryParams holds dependencies for repository initialization
type RepositoryParams struct {
	fx.In
	Config *config.Config
	DB     *database.Connection
}

// NewUserRepository creates a user repository based on the configured database driver
func NewUserRepository(p RepositoryParams) domain.UserRepository {
	switch p.Config.Database.Driver {
	case "sqlite", "postgres":
		if p.DB.GORM == nil {
			panic("GORM connection is nil for " + p.Config.Database.Driver)
		}
		return NewUserGormRepository(p.DB.GORM)
	case "mongo":
		if p.DB.Mongo == nil {
			panic("MongoDB connection is nil")
		}
		database := p.DB.Mongo.Database(p.Config.Database.MongoDatabase)
		return NewUserMongoRepository(database)
	default:
		panic("unsupported database driver: " + p.Config.Database.Driver)
	}
}

// isUniqueConstraintError checks if the error is a unique constraint violation
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := strings.ToLower(err.Error())
	
	// PostgreSQL
	if strings.Contains(errStr, "duplicate key") || strings.Contains(errStr, "violates unique constraint") {
		return true
	}
	
	// SQLite
	if strings.Contains(errStr, "unique constraint failed") || strings.Contains(errStr, "constraint failed: unique") {
		return true
	}
	
	// GORM specific
	if strings.Contains(errStr, "duplicate entry") {
		return true
	}
	
	return false
}

// CreateIndexes creates necessary indexes for MongoDB
// Note: This is now handled by the migration system
func CreateIndexes(db *mongo.Database) error {
	// This is handled in individual repository constructors
	// But you could implement global index creation here
	return nil
}