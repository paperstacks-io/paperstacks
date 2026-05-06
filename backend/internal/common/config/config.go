// Package config provides application-wide startup configuration loaded from
// environment variables.
package config

import (
	"fmt"
	"os"
)

// Config contains application startup configuration.
type Config struct {
	Host        string
	Port        string
	HankoAPIURL string
}

// New loads configuration from environment variables and applies defaults.
func New() (Config, error) {
	hankoAPIURL, err := getRequiredEnv("HANKO_API_URL")
	if err != nil {
		return Config{}, err
	}

	return Config{
		Host:        getEnvOrDefault("HOST", "127.0.0.1"),
		Port:        getEnvOrDefault("PORT", "8080"),
		HankoAPIURL: hankoAPIURL,
	}, nil
}

func getEnvOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func getRequiredEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("required environment variable %q is not set", key)
	}

	return value, nil
}
