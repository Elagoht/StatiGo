package utils

import (
	"log/slog"
	"os"
	"strconv"
)

// GetEnvInt retrieves an integer environment variable with a default value.
func GetEnvInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// GetEnvString retrieves a string environment variable with a default value.
func GetEnvString(key string, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

// GetEnvBool retrieves a boolean environment variable with a default value.
func GetEnvBool(key string, defaultValue bool) bool {
	if val := os.Getenv(key); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

// RequiredEnvVars holds the configuration for required environment variables.
type RequiredEnvVars struct {
	Required []string
	Logger   *slog.Logger
}

// CheckRequired validates that all required environment variables are set.
func (r *RequiredEnvVars) CheckRequired() bool {
	allPresent := true
	for _, key := range r.Required {
		if os.Getenv(key) == "" {
			r.Logger.Error("Required environment variable not set", "key", key)
			allPresent = false
		}
	}
	return allPresent
}
