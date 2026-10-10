// Package audit responsible per rules domains the audit logs
package audit

import (
	baseDomain "github.com/alisonsandrade/go-start-project/internal/domain"
	"github.com/alisonsandrade/go-start-project/pkg/token"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Log representa o registro de auditoria persistido no banco de dados
type Log struct {
	baseDomain.BaseModelTenant
	UserID    *uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	Action    string     `gorm:"size:50;not null" json:"action"`
	Resource  string     `gorm:"size:255;not null" json:"resource"`
	IPAddress string     `gorm:"size:50" json:"ip_address"`
	UserAgent string     `gorm:"type:text" json:"user_agent"`
}

func (Log) TableName() string {
	return "audit_logs"
}

// BeforeCreate sobrescreve o hook BeforeCreate do BaseModelTenant.
// Isso evita a falha quando o TenantID já foi populado diretamente na struct
// ou quando o log de auditoria é persistido de forma assíncrona.
func (l *Log) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}

	if l.TenantID == uuid.Nil && tx != nil && tx.Statement != nil && tx.Statement.Context != nil {
		if claims, ok := tx.Statement.Context.Value(token.ClaimsContextKey).(*token.CustomClaims); ok && claims != nil {
			l.TenantID = claims.TenantID
		}
	}

	return nil
}

// CreateLogRequest representa a estrutura para criação de um registro de auditoria com validações.
type CreateLogRequest struct {
	UserID    *uuid.UUID `json:"user_id" validate:"omitempty,uuid"`
	Action    string     `json:"action" validate:"required,min=1,max=50"`
	Resource  string     `json:"resource" validate:"required,min=1,max=255"`
	IPAddress string     `json:"ip_address" validate:"omitempty,max=50"`
	UserAgent string     `json:"user_agent" validate:"omitempty"`
}

// ListAuditLogsQuery representa os parâmetros para consulta/listagem com validação.
type ListAuditLogsQuery struct {
	Limit  int `json:"limit" validate:"omitempty,min=1,max=100"`
	Offset int `json:"offset" validate:"omitempty,min=0"`
}
