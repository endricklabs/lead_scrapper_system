package model

import (
	"time"

	"github.com/google/uuid"
)

type UserSubscriptionStatus string

const (
	UserSubscriptionStatusActive   UserSubscriptionStatus = "active"
	UserSubscriptionStatusInactive UserSubscriptionStatus = "inactive"
	UserSubscriptionStatusExpired  UserSubscriptionStatus = "expired"
)

type UserSubscription struct {
	BaseModel
	UserID                uuid.UUID              `gorm:"type:uuid;not null"`
	User                  User                   `gorm:"foreignKey:UserID"`
	SubscriptionPackageID uuid.UUID              `gorm:"type:uuid;not null"`
	SubscriptionPackage   SubscriptionPackage    `gorm:"foreignKey:SubscriptionPackageID"`
	Status                UserSubscriptionStatus `gorm:"not null"`
	StartDate             time.Time              `gorm:"not null"`
	EndDate               time.Time              `gorm:"not null"`
}
