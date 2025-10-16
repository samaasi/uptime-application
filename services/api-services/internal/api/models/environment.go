package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Environment represents an environment in the system.
// An environment belongs to an application.
type Environment struct {
	Model
	Name          string         `json:"name"`
	Color         string         `json:"color" gorm:"type:varchar(100);not null"`
	Url           *string        `json:"url" gorm:"type:varchar(100);not null"`
	ApplicationID uuid.UUID      `json:"application_id" gorm:"type:uuid;not null;index"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	Application Application `json:"application" gorm:"foreignKey:ApplicationID"`
}
