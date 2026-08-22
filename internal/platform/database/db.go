// Package database config
package database

import (
	"fmt"
	"log"

	"github.com/alisonsandrade/go-start-project/internal/config"
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
