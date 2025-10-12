package repositories

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/samaasi/uptime-application/services/api-services/internal/api/models"
	"github.com/samaasi/uptime-application/services/api-services/pkg/logger"
	"gorm.io/gorm"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	EmailExists(ctx context.Context, email string) (bool, error)
	// AddToOrganization(ctx context.Context, userID, organizationID uuid.UUID) error
	// RemoveFromOrganization(ctx context.Context, userID, organizationID uuid.UUID) error
	IsInSameOrganization(ctx context.Context, userID1, userID2 uuid.UUID) (bool, error)
	// AssignPermission(ctx context.Context, userID, permissionID uuid.UUID) error
	// RemovePermission(ctx context.Context, userID, permissionID uuid.UUID) error
}

// userRepository implements UserRepository interface
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new instance of userRepository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user
func (ur *userRepository) Create(ctx context.Context, user *models.User) error {
	if err := ur.db.WithContext(ctx).Create(user).Error; err != nil {
		logger.Error("Failed to create user", logger.ErrorField(err))
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetByID retrieves a user by ID
func (ur *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := ur.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetByEmail retrieves a user by email
func (ur *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := ur.db.WithContext(ctx).
		Where("email = ? AND deleted_at IS NULL", email).
		First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// Update updates a user
func (ur *userRepository) Update(ctx context.Context, user *models.User) error {
	if err := ur.db.WithContext(ctx).Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// SoftDelete soft deletes a user
func (ur *userRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if err := ur.db.WithContext(ctx).Delete(&models.User{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// EmailExists checks if an email already exists
func (ur *userRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	err := ur.db.WithContext(ctx).
		Model(&models.User{}).
		Where("email = ? AND deleted_at IS NULL", email).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}
	return count > 0, nil
}

// IsInSameOrganization checks if two users are in the same organization
func (ur *userRepository) IsInSameOrganization(ctx context.Context, userID1, userID2 uuid.UUID) (bool, error) {
	var count int64
	err := ur.db.WithContext(ctx).
		Table("organization_users ou1").
		Joins("JOIN organization_users ou2 ON ou1.organization_id = ou2.organization_id").
		Where("ou1.user_id = ? AND ou2.user_id = ?", userID1, userID2).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to check organization membership: %w", err)
	}
	return count > 0, nil
}
