package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration
type Config struct {
	ClientIP        string        `json:"client_ip"`
	APIKey          string        `json:"api_key"`
	ValidationDir   string        `json:"validation_dir"`
	SSLDir          string        `json:"ssl_dir"`
	ContainerName   string        `json:"container_name"`
	RenewalInterval time.Duration `json:"renewal_interval"`
	CertValidity    time.Duration `json:"cert_validity"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		ClientIP:        getEnv("CLIENT_IP", "127.0.0.1"),
		APIKey:          getEnv("IPSSL_API_KEY", ""),
		ValidationDir:   getEnv("IPSSL_VALIDATION_DIR", "/usr/share/caddy/"),
		SSLDir:          getEnv("IPSSL_SSL_DIR", "/ipssl/"),
		ContainerName:   getEnv("IPSSL_CONTAINER_NAME", "caddy-1"),
		RenewalInterval: getDurationEnv("RENEWAL_INTERVAL", 24*time.Hour),
		CertValidity:    getDurationEnv("CERT_VALIDITY", 30*24*time.Hour),
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("IPSSL_API_KEY environment variable is required")
	}

	return cfg, nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getDurationEnv gets a duration environment variable with a default value
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// getIntEnv gets an integer environment variable with a default value
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
