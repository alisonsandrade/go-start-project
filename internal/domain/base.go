// Package domain provide a base model for all application
package domain

import (
	"errors"
	"time"

	"github.com/alisonsandrade/go-start-project/pkg/token"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel is the base for all entities
// It replaces then standard gorm.Model to use UUID and logical isolation.
type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BaseModelTenant is the base for all entities belonging to an institutions/category
type BaseModelTenant struct {
	BaseModel
	TenantID uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
}

func (base *BaseModelTenant) BeforeCreate(tx *gorm.DB) error {
	if tx == nil {
		return base.BaseModel.BeforeCreate(tx)
	}

	claims, ok := tx.Statement.Context.Value(token.ClaimsContextKey).(*token.CustomClaims)
	if !ok || claims == nil || claims.TenantID == uuid.Nil {
		return errors.New("tenant_id é obrigatório para criar este registro")
	}

	base.TenantID = claims.TenantID
	return base.BaseModel.BeforeCreate(tx)
}

// BeforeCreate é um Hook do GORM interceptando a criação do registro.
// Ele garante que toda entidade ganhe um UUID automaticamente sem depender de
// extensões nativas do Postgres (como uuid-ossp), facilitando testes no SQLite.
func (base *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if base.ID == uuid.Nil {
		base.ID = uuid.New()
	}
	return nil
}
