package ipssl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"ipssl-client/internal/config"
	"ipssl-client/internal/docker"
	"ipssl-client/internal/logger"
	"ipssl-client/internal/zerossl"
)

// Client represents the IPSSL client
type Client struct {
	config  *config.Config
	logger  *logger.Logger
	zerossl *zerossl.Client
	docker  *docker.Client
}

// NewClient creates a new IPSSL client
func NewClient(cfg *config.Config, logger *logger.Logger) (*Client, error) {
	// Initialize ZeroSSL client
	zerosslClient, err := zerossl.NewClient(cfg.APIKey, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create ZeroSSL client: %w", err)
	}

	// Initialize Docker client only if container name is specified
	var dockerClient *docker.Client
	if cfg.ContainerName != "" {
		dockerClient, err = docker.NewClient(logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create Docker client: %w", err)
		}
		logger.Info("Docker client initialized", "container_name", cfg.ContainerName)
	} else {
		logger.Info("Docker client not initialized - no container name specified")
	}

	return &Client{
		config:  cfg,
		logger:  logger,
		zerossl: zerosslClient,
		docker:  dockerClient,
	}, nil
}

// Start starts the IPSSL client with automatic renewal
func (c *Client) Start(ctx context.Context) error {
	c.logger.Info("Starting IPSSL client")

	// Ensure directories exist
	if err := c.ensureDirectories(); err != nil {
		return fmt.Errorf("failed to ensure directories: %w", err)
	}

	// Check if certificate already exists and is valid
	if c.isCertificateValid() {
		c.logger.Info("Valid certificate already exists, skipping initial download")
	} else {
		// Request new certificate (file missing or expired)
		c.logger.Info("Certificate needs to be downloaded (missing or invalid)")
		if err := c.requestCertificate(ctx); err != nil {
			return fmt.Errorf("failed to request certificate: %w", err)
		}
	}

	// Start renewal ticker
	ticker := time.NewTicker(c.config.RenewalInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("IPSSL client stopped")
			return ctx.Err()
		case <-ticker.C:
			if !c.isCertificateValid() {
				c.logger.Info("Certificate needs renewal (missing, expired, or expiring soon)")
				if err := c.requestCertificate(ctx); err != nil {
					c.logger.Error("Failed to renew certificate", "error", err)
					continue
				}
			} else {
				c.logger.Info("Certificate is still valid, skipping renewal")
			}
		}
	}
}

// ensureDirectories ensures that required directories exist
func (c *Client) ensureDirectories() error {
	dirs := []string{
		c.config.SSLDir,
		c.config.ValidationDir,
		filepath.Join(c.config.ValidationDir, ".well-known", "pki-validation"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// certificateFilesExist checks if both certificate and key files exist
func (c *Client) certificateFilesExist() (bool, string) {
	certPath := filepath.Join(c.config.SSLDir, "cert.pem")
	keyPath := filepath.Join(c.config.SSLDir, "key.pem")

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		return false, "certificate file missing"
	}

	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		return false, "private key file missing"
	}

	return true, "both files exist"
}

// isCertificateValid checks if the current certificate is valid
func (c *Client) isCertificateValid() bool {
	certPath := filepath.Join(c.config.SSLDir, "cert.pem")

	// First check if files exist
	filesExist, reason := c.certificateFilesExist()
	if !filesExist {
		c.logger.Info("Certificate files missing, will download new certificate", "reason", reason)
		return false
	}

	// Check certificate validity (expiration, etc.)
	valid, err := c.zerossl.IsCertificateValid(certPath, c.config.CertValidity)
	if err != nil {
		c.logger.Error("Failed to check certificate validity", "error", err, "cert_path", certPath)
		return false
	}

	if !valid {
		c.logger.Info("Certificate is expired or will expire soon, will download new certificate", "cert_path", certPath)
	}

	return valid
}

// requestCertificate requests a new certificate from ZeroSSL
func (c *Client) requestCertificate(ctx context.Context) error {
	c.logger.Info("Requesting new certificate", "ip", c.config.ClientIP)

	// Request certificate from ZeroSSL
	cert, key, err := c.zerossl.RequestCertificate(ctx, c.config.ClientIP)
	if err != nil {
		return fmt.Errorf("failed to request certificate from ZeroSSL: %w", err)
	}

	// Log certificate chain information
	certStr := string(cert)
	// Count certificate blocks (each certificate starts with -----BEGIN CERTIFICATE-----)
	certBlocks := 0
	beginMarker := "-----BEGIN CERTIFICATE-----"
	for i := 0; i < len(certStr)-len(beginMarker); i++ {
		if certStr[i:i+len(beginMarker)] == beginMarker {
			certBlocks++
		}
	}
	c.logger.Info("Certificate chain received", "total_certificates", certBlocks, "cert_size_bytes", len(cert))

	// Save certificate files
	certPath := filepath.Join(c.config.SSLDir, "cert.pem")
	keyPath := filepath.Join(c.config.SSLDir, "key.pem")

	if err := os.WriteFile(certPath, cert, 0644); err != nil {
		return fmt.Errorf("failed to save certificate: %w", err)
	}

	if err := os.WriteFile(keyPath, key, 0600); err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}

	c.logger.Info("Certificate saved successfully",
		"cert_path", certPath,
		"key_path", keyPath)

	// Reload Caddy container (only if Docker client is available)
	if c.docker != nil && c.config.ContainerName != "" {
		if err := c.docker.ReloadContainer(ctx, c.config.ContainerName); err != nil {
			c.logger.Error("Failed to reload Caddy container", "error", err)
			// Don't return error here as certificate was saved successfully
		}
	} else {
		c.logger.Info("Skipping container reload - Docker client not available or no container name specified")
	}

	return nil
}
