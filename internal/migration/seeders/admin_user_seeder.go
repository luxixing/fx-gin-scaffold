package seeders

import (
	"context"
	"time"

	"github.com/luxixing/fx-gin-scaffold/internal/domain"
	"github.com/luxixing/fx-gin-scaffold/pkg/database"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

// AdminUserSeeder creates the default admin user
type AdminUserSeeder struct{}

func (s *AdminUserSeeder) Name() string {
	return "AdminUserSeeder"
}

func (s *AdminUserSeeder) ShouldRun(env string) bool {
	// Run in development and staging environments
	return env == "development" || env == "staging"
}

func (s *AdminUserSeeder) Run(ctx context.Context, db *database.Connection) error {
	adminUser := &domain.User{
		Email:     "admin@example.com",
		Password:  "admin123456", // Will be hashed
		Name:      "System Administrator",
		Role:      "admin",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Hash the password
	if err := adminUser.HashPassword(); err != nil {
		return err
	}

	if db.GORM != nil {
		return s.seedSQL(db.GORM, adminUser)
	}

	if db.Mongo != nil {
		return s.seedMongo(ctx, db.Mongo, adminUser)
	}

	return nil
}

func (s *AdminUserSeeder) seedSQL(gormDB *gorm.DB, user *domain.User) error {
	// Check if admin user already exists
	var existingUser domain.User
	err := gormDB.Where("email = ?", user.Email).First(&existingUser).Error
	if err == nil {
		// User already exists, skip
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		// Some other error occurred
		return err
	}

	// Create the admin user
	return gormDB.Create(user).Error
}

func (s *AdminUserSeeder) seedMongo(ctx context.Context, mongoDB *mongo.Client, user *domain.User) error {
	dbName := "fx_gin_scaffold" // TODO: Get from config
	collection := mongoDB.Database(dbName).Collection("fx_users")

	// Check if admin user already exists
	count, err := collection.CountDocuments(ctx, map[string]interface{}{
		"email": user.Email,
	})
	if err != nil {
		return err
	}
	if count > 0 {
		// User already exists, skip
		return nil
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

	_, err = collection.InsertOne(ctx, userDoc)
	return err
}