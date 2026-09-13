// Package domain define as entidades de negócio e DTOs.
package domain

import (
	"time"

	baseDomain "github.com/alisonsandrade/go-start-project/internal/domain"
	"github.com/google/uuid"
)

type RefreshToken struct {
	baseDomain.BaseModelTenant
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Token     string    `gorm:"size:512;uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
}
