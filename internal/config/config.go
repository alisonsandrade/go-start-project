// Package  config
package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type AdminSeedConfig struct {
	Name     string
	Email    string
	Password string
}

type Config struct {
	Port               string
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	DBSSLMode          string
	JWTSecret          string
	JWTExpirationHours string
	AdminSeed          AdminSeedConfig
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:               getEnv("PORT", "8000"),
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnv("DB_PORT", "5432"),
		DBUser:             getEnv("DB_USER", "postgres"),
		DBPassword:         getEnv("DB_PASSWORD", "postgres"),
		DBName:             getEnv("DB_NAME", "gostartdb"),
		DBSSLMode:          getEnv("DB_SSLMODE", "disable"),
		JWTSecret:          getEnv("JWT_SECRET", "default_secret_key"),
		JWTExpirationHours: getEnv("JWT_EXPIRATION_HOURS", "24"),
		AdminSeed: AdminSeedConfig{
			Name:     getEnv("ADMIN_DEFAULT_NAME", "Super Admin"),
			Email:    getEnv("ADMIN_DEFAULT_EMAIL", "admin@sandrade.com"),
			Password: getEnv("ADMIN_DEFAULT_PASSWORD", "Admin@123456"),
		},
	}

	return cfg, nil
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort, c.DBSSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return defaultValue
}
