package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Organization represents an organization in the system.
// An organization can have multiple users, and policies.
type Organization struct {
	Model
	OwnerID   uuid.UUID      `json:"owner_id" gorm:"type:uuid;index"`
	Owner     *User          `json:"owner" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Name      string         `json:"name" gorm:"type:varchar(100);not null"`
	Users     []User         `json:"users" gorm:"many2many:organization_users;"`
	Policies  []Policy       `json:"policies" gorm:"foreignKey:OrganizationID"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// OrganizationUser represents the pivot table for Organization-User relationship.
type OrganizationUser struct {
	OrganizationID uuid.UUID `json:"organization_id" gorm:"type:uuid;not null;index:idx_organization_users,priority:1"`
	UserID         uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index:idx_organization_users,priority:2"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
}
