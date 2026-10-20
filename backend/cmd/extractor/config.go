package main

import "os"

// Config contains extractor service startup configuration.
type Config struct {
	Host string
	Port string
}

// NewConfig loads extractor configuration from environment variables with sensible defaults.
func NewConfig() Config {
	return Config{
		Host: getEnvOrDefault("HOST", "0.0.0.0"),
		Port: getEnvOrDefault("PORT", "8081"),
	}
}

func getEnvOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

