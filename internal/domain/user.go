package domain

import (
	"context"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey" bson:"_id,omitempty"`
	Email     string    `json:"email" gorm:"uniqueIndex:idx_users_email;not null;size:255" bson:"email" validate:"required,email"`
	Password  string    `json:"-" gorm:"not null;size:255" bson:"password" validate:"required,min=8"`
	Name      string    `json:"name" gorm:"not null;size:100;index:idx_users_name" bson:"name" validate:"required,min=2"`
	Role      string    `json:"role" gorm:"default:user;size:50;index:idx_users_role,idx_users_role_active" bson:"role"`
	Active    bool      `json:"active" gorm:"default:true;index:idx_users_active,idx_users_role_active" bson:"active"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime;index:idx_users_created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime" bson:"updated_at"`
}

// TableName returns the table name for User model
func (User) TableName() string {
	return GetTableName("users")
}

// UserCreateRequest represents the request for creating a new user
type UserCreateRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required,min=2"`
	Role     string `json:"role,omitempty"`
}

// UserUpdateRequest represents the request for updating a user
type UserUpdateRequest struct {
	Name   *string `json:"name,omitempty" validate:"omitempty,min=2"`
	Role   *string `json:"role,omitempty"`
	Active *bool   `json:"active,omitempty"`
}

// UserLoginRequest represents the login request
type UserLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UserResponse represents the user data returned to clients (without sensitive data)
type UserResponse struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Role:      u.Role,
		Active:    u.Active,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// HashPassword hashes the user's password
func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword compares the provided password with the stored hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// IsAdmin returns true if the user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *User) error
	
	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uint) (*User, error)
	
	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*User, error)
	
	// Update updates an existing user
	Update(ctx context.Context, user *User) error
	
	// Delete soft deletes a user
	Delete(ctx context.Context, id uint) error
	
	// List retrieves users with pagination
	List(ctx context.Context, offset, limit int) ([]*User, int64, error)
	
	// Search searches users by name or email
	Search(ctx context.Context, query string, offset, limit int) ([]*User, int64, error)
}

// UserService defines the interface for user business logic
type UserService interface {
	// Register creates a new user account
	Register(ctx context.Context, req *UserCreateRequest) (*UserResponse, error)
	
	// Login authenticates a user and returns a token
	Login(ctx context.Context, req *UserLoginRequest) (string, *UserResponse, error)
	
	// GetProfile retrieves the user's profile
	GetProfile(ctx context.Context, userID uint) (*UserResponse, error)
	
	// UpdateProfile updates the user's profile
	UpdateProfile(ctx context.Context, userID uint, req *UserUpdateRequest) (*UserResponse, error)
	
	// GetUser retrieves a user by ID (admin only)
	GetUser(ctx context.Context, id uint) (*UserResponse, error)
	
	// ListUsers retrieves users with pagination (admin only)
	ListUsers(ctx context.Context, offset, limit int) ([]*UserResponse, int64, error)
	
	// SearchUsers searches users (admin only)
	SearchUsers(ctx context.Context, query string, offset, limit int) ([]*UserResponse, int64, error)
	
	// UpdateUser updates a user (admin only)
	UpdateUser(ctx context.Context, id uint, req *UserUpdateRequest) (*UserResponse, error)
	
	// DeleteUser deletes a user (admin only)
	DeleteUser(ctx context.Context, id uint) error
}