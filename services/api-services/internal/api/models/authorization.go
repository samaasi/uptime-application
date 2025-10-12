package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role represents a role within an organization for RBAC.
type Role struct {
	Model
	OrganizationID uuid.UUID      `json:"organization_id" gorm:"type:uuid;not null;index"`
	Name           string         `json:"name" gorm:"type:varchar(50);not null"`
	Permissions    []Permission   `json:"permissions" gorm:"many2many:role_permissions;"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// Permission represents an action a role or user can perform.
type Permission struct {
	Model
	Name        string         `json:"name" gorm:"type:varchar(50);unique;not null"` // e.g., "user:assign"
	Description *string        `json:"description,omitempty" gorm:"type:text"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// RolePermission represents the pivot table for Role-Permission relationship.
type RolePermission struct {
	RoleID       uuid.UUID      `json:"role_id" gorm:"type:uuid;not null;index:idx_role_permissions,priority:1"`
	PermissionID uuid.UUID      `json:"permission_id" gorm:"type:uuid;not null;index:idx_role_permissions,priority:2"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// UserPermission represents the pivot table for direct User-Permission relationship.
type UserPermission struct {
	UserID       uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index:idx_user_permissions,priority:1"`
	PermissionID uuid.UUID      `json:"permission_id" gorm:"type:uuid;not null;index:idx_user_permissions,priority:2"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// UserRole represents the pivot table for User-Role relationship.
type UserRole struct {
	UserID    uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index:idx_user_roles,priority:1"`
	RoleID    uuid.UUID      `json:"role_id" gorm:"type:uuid;not null;index:idx_user_roles,priority:2"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// Policy represents an ABAC policy with conditions.
type Policy struct {
	Model
	OrganizationID uuid.UUID              `json:"organization_id" gorm:"type:uuid;not null;index"`
	Name           string                 `json:"name" gorm:"type:varchar(100);not null"`
	Conditions     map[string]interface{} `json:"conditions" gorm:"type:jsonb;not null"`
	Effect         string                 `json:"effect" gorm:"type:varchar(10);not null"`
	DeletedAt      gorm.DeletedAt         `json:"deleted_at" gorm:"index"`
}
