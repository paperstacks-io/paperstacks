// Package config provides application-wide startup configuration loaded from
// environment variables.
package config

import (
	"os"
)

// Config contains application startup configuration.
type Config struct {
	Host        string
	Port        string
	HankoAPIURL string
}

// New loads configuration from environment variables and applies defaults.
func New() Config {
	return Config{
		Host:        getEnvOrDefault("HOST", "127.0.0.1"),
		Port:        getEnvOrDefault("PORT", "8080"),
		HankoAPIURL: getEnvOrDefault("HANKO_API_URL", ""),
	}
}

func getEnvOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}
