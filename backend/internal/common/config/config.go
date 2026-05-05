// Package config provides application-wide startup configuration loaded from
// environment variables.
package config

import "os"

// Config contains application startup configuration.
type Config struct {
	Host string
	Port string
}

// New loads configuration from environment variables and applies defaults.
func New() Config {
	return Config{
		Host: getEnv("HOST", "127.0.0.1"),
		Port: getEnv("PORT", "8080"),
	}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}
