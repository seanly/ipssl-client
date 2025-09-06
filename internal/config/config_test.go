package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Set test environment variables
	os.Setenv("CLIENT_IP", "192.168.1.1")
	os.Setenv("IPSSL_API_KEY", "test-api-key")
	os.Setenv("IPSSL_VALIDATION_DIR", "/test/validation")
	os.Setenv("IPSSL_SSL_DIR", "/test/ssl")
	os.Setenv("IPSSL_CONTAINER_NAME", "test-container")
	os.Setenv("RENEWAL_INTERVAL", "1h")
	os.Setenv("CERT_VALIDITY", "720h")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test values
	if cfg.ClientIP != "192.168.1.1" {
		t.Errorf("Expected ClientIP to be '192.168.1.1', got '%s'", cfg.ClientIP)
	}

	if cfg.APIKey != "test-api-key" {
		t.Errorf("Expected APIKey to be 'test-api-key', got '%s'", cfg.APIKey)
	}

	if cfg.ValidationDir != "/test/validation" {
		t.Errorf("Expected ValidationDir to be '/test/validation', got '%s'", cfg.ValidationDir)
	}

	if cfg.SSLDir != "/test/ssl" {
		t.Errorf("Expected SSLDir to be '/test/ssl', got '%s'", cfg.SSLDir)
	}

	if cfg.ContainerName != "test-container" {
		t.Errorf("Expected ContainerName to be 'test-container', got '%s'", cfg.ContainerName)
	}

	if cfg.RenewalInterval != time.Hour {
		t.Errorf("Expected RenewalInterval to be 1h, got %v", cfg.RenewalInterval)
	}

	if cfg.CertValidity != 30*time.Hour {
		t.Errorf("Expected CertValidity to be 30h, got %v", cfg.CertValidity)
	}

	// Clean up
	os.Unsetenv("CLIENT_IP")
	os.Unsetenv("IPSSL_API_KEY")
	os.Unsetenv("IPSSL_VALIDATION_DIR")
	os.Unsetenv("IPSSL_SSL_DIR")
	os.Unsetenv("IPSSL_CONTAINER_NAME")
	os.Unsetenv("RENEWAL_INTERVAL")
	os.Unsetenv("CERT_VALIDITY")
}

func TestLoadMissingAPIKey(t *testing.T) {
	// Ensure API key is not set
	os.Unsetenv("IPSSL_API_KEY")

	_, err := Load()
	if err == nil {
		t.Error("Expected error when API key is missing, got nil")
	}

	if err.Error() != "IPSSL_API_KEY environment variable is required" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestLoadDefaults(t *testing.T) {
	// Clear all environment variables
	os.Unsetenv("CLIENT_IP")
	os.Unsetenv("IPSSL_API_KEY")
	os.Unsetenv("IPSSL_VALIDATION_DIR")
	os.Unsetenv("IPSSL_SSL_DIR")
	os.Unsetenv("IPSSL_CONTAINER_NAME")
	os.Unsetenv("RENEWAL_INTERVAL")
	os.Unsetenv("CERT_VALIDITY")

	// Set only required API key
	os.Setenv("IPSSL_API_KEY", "test-api-key")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config with defaults: %v", err)
	}

	// Test default values
	if cfg.ClientIP != "127.0.0.1" {
		t.Errorf("Expected default ClientIP, got '%s'", cfg.ClientIP)
	}

	if cfg.ValidationDir != "/usr/share/caddy/" {
		t.Errorf("Expected default ValidationDir, got '%s'", cfg.ValidationDir)
	}

	if cfg.SSLDir != "/ipssl/" {
		t.Errorf("Expected default SSLDir, got '%s'", cfg.SSLDir)
	}

	if cfg.ContainerName != "caddy-1" {
		t.Errorf("Expected default ContainerName, got '%s'", cfg.ContainerName)
	}

	// Clean up
	os.Unsetenv("IPSSL_API_KEY")
}
