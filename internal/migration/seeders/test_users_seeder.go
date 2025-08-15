package seeders

import (
	"context"
	"fmt"
	"time"

	"github.com/luxixing/fx-gin-scaffold/internal/domain"
	"github.com/luxixing/fx-gin-scaffold/pkg/database"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

// TestUsersSeeder creates test users for development
type TestUsersSeeder struct{}

func (s *TestUsersSeeder) Name() string {
	return "TestUsersSeeder"
}

func (s *TestUsersSeeder) ShouldRun(env string) bool {
	// Only run in development environment
	return env == "development"
}

func (s *TestUsersSeeder) Run(ctx context.Context, db *database.Connection) error {
	testUsers := []*domain.User{
		{
			Email:     "user1@example.com",
			Password:  "password123",
			Name:      "Test User One",
			Role:      "user",
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Email:     "user2@example.com",
			Password:  "password123",
			Name:      "Test User Two",
			Role:      "user",
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Email:     "moderator@example.com",
			Password:  "password123",
			Name:      "Test Moderator",
			Role:      "moderator",
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			Email:     "inactive@example.com",
			Password:  "password123",
			Name:      "Inactive User",
			Role:      "user",
			Active:    false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// Hash passwords
	for _, user := range testUsers {
		if err := user.HashPassword(); err != nil {
			return fmt.Errorf("failed to hash password for user %s: %w", user.Email, err)
		}
	}

	if db.GORM != nil {
		return s.seedSQL(db.GORM, testUsers)
	}

	if db.Mongo != nil {
		return s.seedMongo(ctx, db.Mongo, testUsers)
	}

	return nil
}

func (s *TestUsersSeeder) seedSQL(gormDB *gorm.DB, users []*domain.User) error {
	for _, user := range users {
		// Check if user already exists
		var existingUser domain.User
		err := gormDB.Where("email = ?", user.Email).First(&existingUser).Error
		if err == nil {
			// User already exists, skip
			continue
		}
		if err != gorm.ErrRecordNotFound {
			// Some other error occurred
			return err
		}

		// Create the user
		if err := gormDB.Create(user).Error; err != nil {
			return fmt.Errorf("failed to create user %s: %w", user.Email, err)
		}
	}

	return nil
}

func (s *TestUsersSeeder) seedMongo(ctx context.Context, mongoDB *mongo.Client, users []*domain.User) error {
	dbName := "fx_gin_scaffold" // TODO: Get from config
	collection := mongoDB.Database(dbName).Collection("fx_users")

	for _, user := range users {
		// Check if user already exists
		count, err := collection.CountDocuments(ctx, map[string]interface{}{
			"email": user.Email,
		})
		if err != nil {
			return err
		}
		if count > 0 {
			// User already exists, skip
			continue
		}

		// Convert user to MongoDB document
		userDoc := map[string]interface{}{
			"email":      user.Email,
			"password":   user.Password,
			"name":       user.Name,
			"role":       user.Role,
			"active":     user.Active,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		}

		if _, err := collection.InsertOne(ctx, userDoc); err != nil {
			return fmt.Errorf("failed to create user %s: %w", user.Email, err)
		}
	}

	return nil
}