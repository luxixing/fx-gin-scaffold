package repo

import (
	"context"
	"testing"
	"time"

	"github.com/luxixing/fx-gin-scaffold/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// UserGormRepositoryTestSuite defines the test suite for user GORM repository
type UserGormRepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo domain.UserRepository
}

// SetupSuite sets up the test suite
func (suite *UserGormRepositoryTestSuite) SetupSuite() {
	// Create in-memory SQLite database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(suite.T(), err)

	// Run migrations
	err = db.AutoMigrate(&domain.User{})
	require.NoError(suite.T(), err)

	suite.db = db
	suite.repo = NewUserGormRepository(db)
}

// TearDownSuite tears down the test suite
func (suite *UserGormRepositoryTestSuite) TearDownSuite() {
	sqlDB, err := suite.db.DB()
	require.NoError(suite.T(), err)
	sqlDB.Close()
}

// SetupTest sets up each test
func (suite *UserGormRepositoryTestSuite) SetupTest() {
	// Clean the database before each test
	suite.db.Exec("DELETE FROM users")
}

// TestCreateUser tests user creation
func (suite *UserGormRepositoryTestSuite) TestCreateUser() {
	ctx := context.Background()
	
	user := &domain.User{
		Email:     "test@example.com",
		Password:  "hashedpassword",
		Name:      "Test User",
		Role:      "user",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := suite.repo.Create(ctx, user)
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), user.ID)
}

// TestCreateUserDuplicateEmail tests creating a user with duplicate email
func (suite *UserGormRepositoryTestSuite) TestCreateUserDuplicateEmail() {
	ctx := context.Background()
	
	user1 := &domain.User{
		Email:    "test@example.com",
		Password: "hashedpassword",
		Name:     "Test User 1",
		Role:     "user",
		Active:   true,
	}

	err := suite.repo.Create(ctx, user1)
	assert.NoError(suite.T(), err)

	user2 := &domain.User{
		Email:    "test@example.com",
		Password: "hashedpassword",
		Name:     "Test User 2",
		Role:     "user",
		Active:   true,
	}

	err = suite.repo.Create(ctx, user2)
	assert.Equal(suite.T(), domain.ErrUserExists, err)
}

// TestGetByID tests getting a user by ID
func (suite *UserGormRepositoryTestSuite) TestGetByID() {
	ctx := context.Background()
	
	// Create a user first
	user := &domain.User{
		Email:    "test@example.com",
		Password: "hashedpassword",
		Name:     "Test User",
		Role:     "user",
		Active:   true,
	}

	err := suite.repo.Create(ctx, user)
	require.NoError(suite.T(), err)

	// Get the user by ID
	retrievedUser, err := suite.repo.GetByID(ctx, user.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), user.Email, retrievedUser.Email)
	assert.Equal(suite.T(), user.Name, retrievedUser.Name)
}

// TestGetByIDNotFound tests getting a non-existent user by ID
func (suite *UserGormRepositoryTestSuite) TestGetByIDNotFound() {
	ctx := context.Background()
	
	_, err := suite.repo.GetByID(ctx, 999)
	assert.Equal(suite.T(), domain.ErrUserNotFound, err)
}

// TestGetByEmail tests getting a user by email
func (suite *UserGormRepositoryTestSuite) TestGetByEmail() {
	ctx := context.Background()
	
	// Create a user first
	user := &domain.User{
		Email:    "test@example.com",
		Password: "hashedpassword",
		Name:     "Test User",
		Role:     "user",
		Active:   true,
	}

	err := suite.repo.Create(ctx, user)
	require.NoError(suite.T(), err)

	// Get the user by email
	retrievedUser, err := suite.repo.GetByEmail(ctx, "test@example.com")
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), user.Email, retrievedUser.Email)
	assert.Equal(suite.T(), user.Name, retrievedUser.Name)
}

// TestGetByEmailNotFound tests getting a non-existent user by email
func (suite *UserGormRepositoryTestSuite) TestGetByEmailNotFound() {
	ctx := context.Background()
	
	_, err := suite.repo.GetByEmail(ctx, "nonexistent@example.com")
	assert.Equal(suite.T(), domain.ErrUserNotFound, err)
}

// TestUpdateUser tests updating a user
func (suite *UserGormRepositoryTestSuite) TestUpdateUser() {
	ctx := context.Background()
	
	// Create a user first
	user := &domain.User{
		Email:    "test@example.com",
		Password: "hashedpassword",
		Name:     "Test User",
		Role:     "user",
		Active:   true,
	}

	err := suite.repo.Create(ctx, user)
	require.NoError(suite.T(), err)

	// Update the user
	user.Name = "Updated User"
	user.UpdatedAt = time.Now()

	err = suite.repo.Update(ctx, user)
	assert.NoError(suite.T(), err)

	// Verify the update
	retrievedUser, err := suite.repo.GetByID(ctx, user.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated User", retrievedUser.Name)
}

// TestDeleteUser tests deleting a user
func (suite *UserGormRepositoryTestSuite) TestDeleteUser() {
	ctx := context.Background()
	
	// Create a user first
	user := &domain.User{
		Email:    "test@example.com",
		Password: "hashedpassword",
		Name:     "Test User",
		Role:     "user",
		Active:   true,
	}

	err := suite.repo.Create(ctx, user)
	require.NoError(suite.T(), err)

	// Delete the user
	err = suite.repo.Delete(ctx, user.ID)
	assert.NoError(suite.T(), err)

	// Verify the user is deleted
	_, err = suite.repo.GetByID(ctx, user.ID)
	assert.Equal(suite.T(), domain.ErrUserNotFound, err)
}

// TestListUsers tests listing users with pagination
func (suite *UserGormRepositoryTestSuite) TestListUsers() {
	ctx := context.Background()
	
	// Create multiple users
	users := []*domain.User{
		{Email: "user1@example.com", Password: "pass", Name: "User 1", Role: "user", Active: true},
		{Email: "user2@example.com", Password: "pass", Name: "User 2", Role: "user", Active: true},
		{Email: "user3@example.com", Password: "pass", Name: "User 3", Role: "admin", Active: true},
	}

	for _, user := range users {
		err := suite.repo.Create(ctx, user)
		require.NoError(suite.T(), err)
	}

	// List users with pagination
	retrievedUsers, total, err := suite.repo.List(ctx, 0, 2)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(3), total)
	assert.Len(suite.T(), retrievedUsers, 2)
}

// TestSearchUsers tests searching users
func (suite *UserGormRepositoryTestSuite) TestSearchUsers() {
	ctx := context.Background()
	
	// Create multiple users
	users := []*domain.User{
		{Email: "john@example.com", Password: "pass", Name: "John Doe", Role: "user", Active: true},
		{Email: "jane@example.com", Password: "pass", Name: "Jane Smith", Role: "user", Active: true},
		{Email: "admin@example.com", Password: "pass", Name: "Admin User", Role: "admin", Active: true},
	}

	for _, user := range users {
		err := suite.repo.Create(ctx, user)
		require.NoError(suite.T(), err)
	}

	// Search by name
	searchResults, total, err := suite.repo.Search(ctx, "John", 0, 10)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), total)
	assert.Len(suite.T(), searchResults, 1)
	assert.Equal(suite.T(), "John Doe", searchResults[0].Name)

	// Search by email
	searchResults, total, err = suite.repo.Search(ctx, "admin", 0, 10)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), total)
	assert.Len(suite.T(), searchResults, 1)
	assert.Equal(suite.T(), "admin@example.com", searchResults[0].Email)
}

// TestUserGormRepository runs the test suite
func TestUserGormRepository(t *testing.T) {
	suite.Run(t, new(UserGormRepositoryTestSuite))
}