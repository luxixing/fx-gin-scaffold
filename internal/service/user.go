package service

import (
	"context"
	"strings"
	"time"

	"github.com/luxixing/fx-gin-scaffold/internal/domain"
	"go.uber.org/fx"
)

// UserServiceParams holds dependencies for UserService
type UserServiceParams struct {
	fx.In
	UserRepo    domain.UserRepository
	AuthService domain.AuthService
}

// userService implements domain.UserService
type userService struct {
	userRepo    domain.UserRepository
	authService domain.AuthService
}

// NewUserService creates a new user service
func NewUserService(p UserServiceParams) domain.UserService {
	return &userService{
		userRepo:    p.UserRepo,
		authService: p.AuthService,
	}
}

// Register creates a new user account
func (s *userService) Register(ctx context.Context, req *domain.UserCreateRequest) (*domain.UserResponse, error) {
	// Validate input
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Check if user already exists
	if _, err := s.userRepo.GetByEmail(ctx, req.Email); err == nil {
		return nil, domain.ErrUserExists
	} else if err != domain.ErrUserNotFound {
		return nil, err
	}

	// Create user
	user := &domain.User{
		Email:     strings.ToLower(strings.TrimSpace(req.Email)),
		Password:  req.Password,
		Name:      strings.TrimSpace(req.Name),
		Role:      s.getDefaultRole(req.Role),
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Hash password
	if err := user.HashPassword(); err != nil {
		return nil, domain.WrapError(err, domain.ErrCodeInternal, "Failed to hash password")
	}

	// Save user
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user.ToResponse(), nil
}

// Login authenticates a user and returns a token
func (s *userService) Login(ctx context.Context, req *domain.UserLoginRequest) (string, *domain.UserResponse, error) {
	// Validate input
	if err := s.validateLoginRequest(req); err != nil {
		return "", nil, err
	}

	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(req.Email)))
	if err != nil {
		if err == domain.ErrUserNotFound {
			return "", nil, domain.ErrInvalidPassword
		}
		return "", nil, err
	}

	// Check if user is active
	if !user.Active {
		return "", nil, domain.NewError(domain.ErrCodeForbidden, "Account is deactivated")
	}

	// Verify password
	if !user.CheckPassword(req.Password) {
		return "", nil, domain.ErrInvalidPassword
	}

	// Generate token
	token, err := s.authService.GenerateToken(user)
	if err != nil {
		return "", nil, err
	}

	return token, user.ToResponse(), nil
}

// GetProfile retrieves the user's profile
func (s *userService) GetProfile(ctx context.Context, userID uint) (*domain.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return user.ToResponse(), nil
}

// UpdateProfile updates the user's profile
func (s *userService) UpdateProfile(ctx context.Context, userID uint, req *domain.UserUpdateRequest) (*domain.UserResponse, error) {
	// Get current user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Name != nil {
		user.Name = strings.TrimSpace(*req.Name)
		if user.Name == "" {
			return nil, domain.ValidationError("name", "cannot be empty")
		}
	}

	user.UpdatedAt = time.Now()

	// Save changes
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user.ToResponse(), nil
}

// GetUser retrieves a user by ID (admin only)
func (s *userService) GetUser(ctx context.Context, id uint) (*domain.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return user.ToResponse(), nil
}

// ListUsers retrieves users with pagination (admin only)
func (s *userService) ListUsers(ctx context.Context, offset, limit int) ([]*domain.UserResponse, int64, error) {
	users, total, err := s.userRepo.List(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*domain.UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
	}

	return responses, total, nil
}

// SearchUsers searches users (admin only)
func (s *userService) SearchUsers(ctx context.Context, query string, offset, limit int) ([]*domain.UserResponse, int64, error) {
	if strings.TrimSpace(query) == "" {
		return s.ListUsers(ctx, offset, limit)
	}

	users, total, err := s.userRepo.Search(ctx, query, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*domain.UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
	}

	return responses, total, nil
}

// UpdateUser updates a user (admin only)
func (s *userService) UpdateUser(ctx context.Context, id uint, req *domain.UserUpdateRequest) (*domain.UserResponse, error) {
	// Get current user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Name != nil {
		user.Name = strings.TrimSpace(*req.Name)
		if user.Name == "" {
			return nil, domain.ValidationError("name", "cannot be empty")
		}
	}

	if req.Role != nil {
		if *req.Role != "user" && *req.Role != "admin" {
			return nil, domain.ValidationError("role", "must be 'user' or 'admin'")
		}
		user.Role = *req.Role
	}

	if req.Active != nil {
		user.Active = *req.Active
	}

	user.UpdatedAt = time.Now()

	// Save changes
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user.ToResponse(), nil
}

// DeleteUser deletes a user (admin only)
func (s *userService) DeleteUser(ctx context.Context, id uint) error {
	// Check if user exists
	if _, err := s.userRepo.GetByID(ctx, id); err != nil {
		return err
	}

	return s.userRepo.Delete(ctx, id)
}

// validateCreateRequest validates user creation request
func (s *userService) validateCreateRequest(req *domain.UserCreateRequest) error {
	if strings.TrimSpace(req.Email) == "" {
		return domain.ValidationError("email", "is required")
	}

	if strings.TrimSpace(req.Name) == "" {
		return domain.ValidationError("name", "is required")
	}

	if len(strings.TrimSpace(req.Name)) < 2 {
		return domain.ValidationError("name", "must be at least 2 characters")
	}

	if len(req.Password) < 8 {
		return domain.ValidationError("password", "must be at least 8 characters")
	}

	return nil
}

// validateLoginRequest validates login request
func (s *userService) validateLoginRequest(req *domain.UserLoginRequest) error {
	if strings.TrimSpace(req.Email) == "" {
		return domain.ValidationError("email", "is required")
	}

	if req.Password == "" {
		return domain.ValidationError("password", "is required")
	}

	return nil
}

// getDefaultRole returns the default role for a user
func (s *userService) getDefaultRole(requestedRole string) string {
	if requestedRole == "admin" || requestedRole == "user" {
		return requestedRole
	}
	return "user"
}