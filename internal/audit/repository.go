// Package audit
package audit

import (
	"context"

	"github.com/alisonsandrade/go-start-project/internal/platform/database"
	"github.com/alisonsandrade/go-start-project/pkg/token"
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
	// Se a requisição contiver contexto de tenant, aplicamos o TenantScope.
	// Caso contrário (como em rotas públicas de login/forgot-password), salvamos direto.
	db := r.db.WithContext(ctx)
	if claims, ok := ctx.Value(token.ClaimsContextKey).(*token.CustomClaims); ok && claims != nil {
		return db.Scopes(database.TenantScope(ctx)).Create(log).Error
	}

	return db.Create(log).Error
}

func (r *auditRepository) List(ctx context.Context, limit, offset int) ([]Log, error) {
	var logs []Log
	err := r.db.WithContext(ctx).Scopes(database.TenantScope(ctx)).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}
