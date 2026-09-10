// Package audit responsible per rules domains the audit logs
package audit

import (
	"time"

	"github.com/google/uuid"
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
