package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Application represents an application in the system.
// An application can have multiple environments.
type Application struct {
	Model
	Name              string          `json:"name" gorm:"type:varchar(100);not null"`
	Icon              *string         `json:"icon" gorm:"type:varchar(100);not null"`
	Region            string          `json:"region" gorm:"type:varchar(100);not null"`
	Environments      []Environment   `json:"environments" gorm:"foreignKey:ApplicationID"`
	ApplicationTypeID uuid.UUID       `json:"application_type_id" gorm:"type:uuid;not null;index"`
	ApplicationType   ApplicationType `json:"application_type" gorm:"foreignKey:ApplicationTypeID"`
	OrganizationID    uuid.UUID       `json:"organization_id" gorm:"type:uuid;not null;index"`
	Organization      Organization    `json:"organization" gorm:"foreignKey:OrganizationID"`
	DeletedAt         gorm.DeletedAt  `json:"deleted_at" gorm:"index"`
}

type ApplicationType struct {
	Model
	Name        string  `json:"name" gorm:"type:varchar(100);not null"`
	Description *string `json:"description" gorm:"type:varchar(100);not null"`
}
