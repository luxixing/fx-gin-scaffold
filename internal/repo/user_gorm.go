package repo

import (
	"context"
	"errors"

	"github.com/luxixing/fx-gin-scaffold/internal/domain"
	"gorm.io/gorm"
)

// userGormRepository implements UserRepository for GORM-based databases
type userGormRepository struct {
	db *gorm.DB
}

// NewUserGormRepository creates a new GORM-based user repository
func NewUserGormRepository(db *gorm.DB) domain.UserRepository {
	return &userGormRepository{
		db: db,
	}
}

// Create creates a new user
func (r *userGormRepository) Create(ctx context.Context, user *domain.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		if isUniqueConstraintError(err) {
			return domain.ErrUserExists
		}
		return domain.WrapError(err, domain.ErrCodeDatabase, "Failed to create user")
	}
	return nil
}

// GetByID retrieves a user by ID
func (r *userGormRepository) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, domain.WrapError(err, domain.ErrCodeDatabase, "Failed to get user by ID")
	}
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *userGormRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, domain.WrapError(err, domain.ErrCodeDatabase, "Failed to get user by email")
	}
	return &user, nil
}

// Update updates an existing user
func (r *userGormRepository) Update(ctx context.Context, user *domain.User) error {
	result := r.db.WithContext(ctx).Save(user)
	if result.Error != nil {
		if isUniqueConstraintError(result.Error) {
			return domain.ErrUserExists
		}
		return domain.WrapError(result.Error, domain.ErrCodeDatabase, "Failed to update user")
	}
	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

// Delete soft deletes a user
func (r *userGormRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&domain.User{}, id)
	if result.Error != nil {
		return domain.WrapError(result.Error, domain.ErrCodeDatabase, "Failed to delete user")
	}
	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

// List retrieves users with pagination
func (r *userGormRepository) List(ctx context.Context, offset, limit int) ([]*domain.User, int64, error) {
	var users []*domain.User
	var total int64

	// Count total records
	if err := r.db.WithContext(ctx).Model(&domain.User{}).Count(&total).Error; err != nil {
		return nil, 0, domain.WrapError(err, domain.ErrCodeDatabase, "Failed to count users")
	}

	// Get paginated records
	err := r.db.WithContext(ctx).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&users).Error
	if err != nil {
		return nil, 0, domain.WrapError(err, domain.ErrCodeDatabase, "Failed to list users")
	}

	return users, total, nil
}

// Search searches users by name or email
func (r *userGormRepository) Search(ctx context.Context, query string, offset, limit int) ([]*domain.User, int64, error) {
	var users []*domain.User
	var total int64

	searchPattern := "%" + query + "%"
	queryBuilder := r.db.WithContext(ctx).Model(&domain.User{}).
		Where("name ILIKE ? OR email ILIKE ?", searchPattern, searchPattern)

	// Count total records
	if err := queryBuilder.Count(&total).Error; err != nil {
		return nil, 0, domain.WrapError(err, domain.ErrCodeDatabase, "Failed to count search results")
	}

	// Get paginated records
	err := queryBuilder.
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&users).Error
	if err != nil {
		return nil, 0, domain.WrapError(err, domain.ErrCodeDatabase, "Failed to search users")
	}

	return users, total, nil
}