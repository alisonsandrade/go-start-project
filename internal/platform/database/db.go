// Package database config
package database

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/alisonsandrade/go-start-project/internal/config"
	"github.com/alisonsandrade/go-start-project/pkg/token"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar no banco de dados: %w", err)
	}

	log.Println("✅ Conexão com PostgreSQL estabelecida com sucesso!")

	return db, nil
}

// TenantScope reads the HTTP request context and scopes the query to the current tenant
func TenantScope(ctx context.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		claims, ok := ctx.Value(token.ClaimsContextKey).(*token.CustomClaims)

		if ok && claims != nil {
			return db.Where("tenant_id = ?", claims.TenantID)
		}

		_ = db.AddError(errors.New("o banco de dados acessado não pertence ao seu tenant_id"))
		return db
	}
}
