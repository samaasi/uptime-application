package models

import (
	"encoding/json"
	"time"

	"github.com/samaasi/uptime-application/services/api-services/pkg/security"

	"gorm.io/gorm"
)

type User struct {
	Model
	FirstName             string          `json:"first_name" gorm:"type:varchar(50);not null"`
	LastName              string          `json:"last_name" gorm:"type:varchar(50);not null"`
	PhoneNumber           *string         `json:"phone_number" gorm:"type:varchar(20);uniqueIndex"`
	HashedPassword        string          `json:"-" gorm:"type:varchar(255)"`
	Email                 *string         `json:"email" gorm:"type:varchar(255);uniqueIndex"`
	PhoneNumberVerifiedAt *time.Time      `json:"phone_number_verified_at" gorm:"default:null"`
	EmailVerifiedAt       *time.Time      `json:"email_verified_at" gorm:"default:null"`
	DateOfBirth           *time.Time      `json:"date_of_birth" gorm:"default:null"`
	ProfilePictureUrl     *string         `json:"profile_picture_url" gorm:"default:null"`
	Preferences           json.RawMessage `json:"preferences" gorm:"type:jsonb"`
	DeletedAt             gorm.DeletedAt  `json:"-" gorm:"index"`

	// OwnedOrganizations lists organizations where this user is the owner
	OwnedOrganizations []Organization `json:"owned_organizations" gorm:"foreignKey:OwnerID"`
}

// BeforeCreate hook to hash password with Argon2id.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if len(u.HashedPassword) > 0 {
		hashedPassword, err := security.HashPassword(u.HashedPassword, nil)
		if err != nil {
			return err
		}
		u.HashedPassword = hashedPassword
	}
	return nil
}

// BeforeUpdate hook to hash password with Argon2id.
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	if len(u.HashedPassword) > 0 {
		hashedPassword, err := security.HashPassword(u.HashedPassword, nil)
		if err != nil {
			return err
		}
		u.HashedPassword = hashedPassword
	}
	return nil
}
