package seeder

import (
	"context"
	"fmt"

	"errors"

	"github.com/samaasi/uptime-application/services/api-services/internal/api/models"
	"github.com/samaasi/uptime-application/services/api-services/internal/utils"
	"github.com/samaasi/uptime-application/services/api-services/pkg/logger"

	"gorm.io/gorm"
)

// PermissionSeeder handles seeding of permissions into the database
type PermissionSeeder struct {
	db *gorm.DB
}

// NewPermissionSeeder creates a new instance of PermissionSeeder
func NewPermissionSeeder(db *gorm.DB) *PermissionSeeder {
	return &PermissionSeeder{
		db: db,
	}
}

// SeedPermissions seeds the predefined permissions into the database
func (ps *PermissionSeeder) SeedPermissions(ctx context.Context) error {

	permissions := []models.Permission{
		// Organization permissions
		{Name: "organization:create", Description: utils.StringPtr("Create new organizations.")},
		{Name: "organization:read", Description: utils.StringPtr("Read organization details or lists.")},
		{Name: "organization:update", Description: utils.StringPtr("Update organization details.")},
		{Name: "organization:delete", Description: utils.StringPtr("Delete organizations.")}, // (Deletion: Soft)

		// User permissions
		{Name: "user:read", Description: utils.StringPtr("Read user details or lists (within organization or self).")},
		{Name: "user:update", Description: utils.StringPtr("Update other users' details (admin).")},
		{Name: "user:update:self", Description: utils.StringPtr("Update own user details.")},
		{Name: "user:delete", Description: utils.StringPtr("Delete other users (admin).")},
		{Name: "user:delete:self", Description: utils.StringPtr("Delete own user.")},
		{Name: "user:assign:organization", Description: utils.StringPtr("Add/remove users from organizations.")},
		{Name: "user:assign:role", Description: utils.StringPtr("Add/remove roles from users.")},
		{Name: "user:assign:policy", Description: utils.StringPtr("Add/remove policies from users.")},

		// Role permissions
		{Name: "role:create", Description: utils.StringPtr("Create roles in an organization.")},
		{Name: "role:read", Description: utils.StringPtr("Read role details or lists.")},
		{Name: "role:update", Description: utils.StringPtr("Update roles.")},
		{Name: "role:delete", Description: utils.StringPtr("Delete roles.")}, // (Deletion: Soft)
		{Name: "role:assign", Description: utils.StringPtr("Assign/remove roles to/from users.")},

		// Permission permissions
		{Name: "permission:create", Description: utils.StringPtr("Create new permissions.")},
		{Name: "permission:read", Description: utils.StringPtr("Read permission details or lists.")},
		{Name: "permission:update", Description: utils.StringPtr("Update permissions.")},
		{Name: "permission:delete", Description: utils.StringPtr("Delete permissions.")}, // (Deletion: Soft)
		{Name: "permission:assign", Description: utils.StringPtr("Assign/remove permissions to roles/users.")},

		// Policy permissions
		{Name: "policy:create", Description: utils.StringPtr("Create ABAC policies.")},
		{Name: "policy:read", Description: utils.StringPtr("Read policy details or lists.")},
		{Name: "policy:update", Description: utils.StringPtr("Update policies.")},
		{Name: "policy:delete", Description: utils.StringPtr("Delete policies.")}, // (Deletion: Soft)
	}

	logger.Info("Starting permission seeding process")

	return ps.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, permission := range permissions {
			var existingPermission models.Permission
			err := tx.Where("name = ?", permission.Name).First(&existingPermission).Error

			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.Create(&permission).Error; err != nil {
					logger.Error("Failed to create permission", logger.String("name", permission.Name), logger.ErrorField(err))
					return fmt.Errorf("failed to create permission '%s': %w", permission.Name, err)
				}
				logger.Info("Created permission", logger.String("name", permission.Name))
			} else if err != nil {
				logger.Error("Failed to check permission", logger.String("name", permission.Name), logger.ErrorField(err))
				return fmt.Errorf("failed to check permission '%s': %w", permission.Name, err)
			} else {
				logger.Info("Permission already exists, skipping", logger.String("name", permission.Name))
			}
		}
		return nil
	})
}

// SeedDefaultData seeds all default data including permissions
func SeedDefaultData(ctx context.Context, db *gorm.DB) error {
	logger.Info("Starting default data seeding")

	permissionSeeder := NewPermissionSeeder(db)
	if err := permissionSeeder.SeedPermissions(ctx); err != nil {
		logger.Error("Failed to seed permissions", logger.ErrorField(err))
		return fmt.Errorf("failed to seed permissions: %w", err)
	}

	logger.Info("Default data seeding completed successfully")
	return nil
}
