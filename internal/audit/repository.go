// Package audit
package audit

import (
	"context"

	"gorm.io/gorm"
)

type AuditRepository interface {
	Create(ctx context.Context, log *Log) error
	List(ctx context.Context, limit, offset int) ([]Log, error)
}

type auditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Create(ctx context.Context, log *Log) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *auditRepository) List(ctx context.Context, limit, offset int) ([]Log, error) {
	var logs []Log
	err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}
