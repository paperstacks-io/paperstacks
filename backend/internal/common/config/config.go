// Package config provides application-wide startup configuration loaded from
// environment variables.
package config

import (
	"os"

	"github.com/paperstacks.io/paperstacks/internal/common/objectstorage"
)

// Config contains application startup configuration.
type Config struct {
	Host          string
	Port          string
	HankoAPIURL   string
	ObjectStorage objectstorage.Config
}

// New loads configuration from environment variables and applies defaults.
func New() Config {
	return Config{
		Host:        getEnvOrDefault("HOST", "127.0.0.1"),
		Port:        getEnvOrDefault("PORT", "8080"),
		HankoAPIURL: getEnvOrDefault("HANKO_API_URL", ""),
		ObjectStorage: objectstorage.Config{
			Endpoint:        getEnvOrDefault("S3_ENDPOINT_URL", ""),
			AccessKeyID:     getEnvOrDefault("S3_ACCESS_KEY_ID", ""),
			SecretAccessKey: getEnvOrDefault("S3_SECRET_ACCESS_KEY", ""),
			Region:          getEnvOrDefault("S3_REGION", objectstorage.DefaultRegion),
		},
	}
}

func getEnvOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}
