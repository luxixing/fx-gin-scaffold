package migrations

import (
	"context"

	"github.com/luxixing/fx-gin-scaffold/internal/domain"
	"github.com/luxixing/fx-gin-scaffold/pkg/database"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateUsersTable creates the users table/collection
type CreateUsersTable struct{}

func (m *CreateUsersTable) Version() string {
	return "20240815120000"
}

func (m *CreateUsersTable) Description() string {
	return "Create users table/collection"
}

func (m *CreateUsersTable) Up(ctx context.Context, db *database.Connection) error {
	if db.GORM != nil {
		// SQL databases - use GORM AutoMigrate
		return db.GORM.AutoMigrate(&domain.User{})
	}

	if db.Mongo != nil {
		// MongoDB - create collection and indexes
		dbName := "fx_gin_scaffold" // TODO: Get from config
		collection := db.Mongo.Database(dbName).Collection("fx_users")

		// Create indexes for MongoDB
		indexes := []mongo.IndexModel{
			{
				Keys: map[string]interface{}{"email": 1},
				Options: options.Index().
					SetUnique(true).
					SetName("idx_users_email"),
			},
			{
				Keys: map[string]interface{}{"name": 1},
				Options: options.Index().
					SetName("idx_users_name"),
			},
			{
				Keys: map[string]interface{}{"role": 1},
				Options: options.Index().
					SetName("idx_users_role"),
			},
			{
				Keys: map[string]interface{}{"active": 1},
				Options: options.Index().
					SetName("idx_users_active"),
			},
			{
				Keys: map[string]interface{}{
					"role":   1,
					"active": 1,
				},
				Options: options.Index().
					SetName("idx_users_role_active"),
			},
			{
				Keys: map[string]interface{}{"created_at": 1},
				Options: options.Index().
					SetName("idx_users_created_at"),
			},
		}

		_, err := collection.Indexes().CreateMany(ctx, indexes)
		return err
	}

	return nil
}

func (m *CreateUsersTable) Down(ctx context.Context, db *database.Connection) error {
	if db.GORM != nil {
		// SQL databases - drop table
		return db.GORM.Migrator().DropTable(&domain.User{})
	}

	if db.Mongo != nil {
		// MongoDB - drop collection
		dbName := "fx_gin_scaffold" // TODO: Get from config
		collection := db.Mongo.Database(dbName).Collection("fx_users")
		return collection.Drop(ctx)
	}

	return nil
}