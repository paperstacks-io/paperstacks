// Package config provides application-wide startup configuration loaded from
// environment variables.
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/paperstacks.io/paperstacks/internal/common/objectstorage"
)

// Config contains application startup configuration.
type Config struct {
	Host                 string
	Port                 string
	DatabaseURL          string
	DatabasePasswordFile string
	HankoAPIURL          string
	GrobidURL            string
	ObjectStorage        objectstorage.Config
}

// New loads configuration from environment variables and applies defaults.
func New() Config {
	return Config{
		Host:                 getEnvOrDefault("HOST", "127.0.0.1"),
		Port:                 getEnvOrDefault("PORT", "8080"),
		DatabaseURL:          getEnvOrDefault("DATABASE_URL", ""),
		DatabasePasswordFile: getEnvOrDefault("DATABASE_PASSWORD_FILE", "/secrets/db_app_rw_password.txt"),
		HankoAPIURL:          getEnvOrDefault("HANKO_API_URL", ""),
		GrobidURL:            getEnvOrDefault("GROBID_URL", "http://localhost:8070"),
		ObjectStorage: objectstorage.Config{
			Endpoint:        getEnvOrDefault("S3_ENDPOINT_URL", ""),
			AccessKeyID:     getEnvOrDefault("S3_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnvOrDefault("S3_SECRET_ACCESS_KEY", ""),
			Region:          getEnvOrDefault("S3_REGION", ""),
			Bucket:          getEnvOrDefault("S3_BUCKET", ""),
		},
	}
}

// DatabasePoolConfig loads PostgreSQL connection settings and the password file.
func (cfg Config) DatabasePoolConfig() (*pgxpool.Config, error) {
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}
	dbConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database configuration: %w", err)
	}

	password, err := os.ReadFile(cfg.DatabasePasswordFile)
	if err != nil {
		return nil, fmt.Errorf("read database password file: %w", err)
	}
	// Match the trailing whitespace removed by the database init script.
	dbConfig.ConnConfig.Password = strings.TrimRight(string(password), " \t\n\r\v\f")
	if dbConfig.ConnConfig.Password == "" {
		return nil, fmt.Errorf("database password file %q is empty", cfg.DatabasePasswordFile)
	}
	return dbConfig, nil
}

func getEnvOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}
