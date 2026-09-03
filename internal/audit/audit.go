// Package audit responsible per rules domains the audit logs
package audit

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Log representa o registro de auditoria persistido no banco de dados
type Log struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    *uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	Action    string     `gorm:"size:50;not null" json:"action"`
	Resource  string     `gorm:"size:255;not null" json:"resource"`
	IPAddress string     `gorm:"size:50" json:"ip_address"`
	UserAgent string     `gorm:"type:text" json:"user_agent"`
	CreatedAt time.Time  `json:"created_at"`
}

func (Log) TableName() string {
	return "audit_logs"
}

type Repository interface {
	Create(ctx context.Context, log *Log) error
	List(ctx context.Context, limit, offset int) ([]Log, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, log *Log) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *repository) List(ctx context.Context, limit, offset int) ([]Log, error) {
	var logs []Log
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}
