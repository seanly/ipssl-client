package zerossl

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"ipssl-client/internal/logger"

	"github.com/caddyserver/zerossl"
)

// Client represents a ZeroSSL API client
type Client struct {
	apiKey      string
	logger      *logger.Logger
	client      *zerossl.Client
	privateKeys map[string]*rsa.PrivateKey
}

// NewClient creates a new ZeroSSL client
func NewClient(apiKey string, logger *logger.Logger) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	client := zerossl.Client{
		AccessKey: apiKey,
	}

	return &Client{
		apiKey:      apiKey,
		logger:      logger,
		client:      &client,
		privateKeys: make(map[string]*rsa.PrivateKey),
	}, nil
}

// RequestCertificate requests a new certificate for the given IP address
func (c *Client) RequestCertificate(ctx context.Context, ip string) ([]byte, []byte, error) {
	c.logger.Info("Requesting certificate from ZeroSSL", "ip", ip)
	c.logger.Info("=== ENTERING RequestCertificate METHOD ===")

	// First, check if there's already an existing certificate request
	c.logger.Info("Checking for existing certificate", "ip", ip)
	existingCertID, err := c.findExistingCertificate(ctx, ip)
	if err != nil {
		c.logger.Warn("Failed to check for existing certificate", "error", err)
	} else if existingCertID != "" {
		c.logger.Info("Found existing certificate", "cert_id", existingCertID)
	} else {
		c.logger.Info("No existing certificate found, will create new one")
	}

	var certObj *zerossl.CertificateObject
	if existingCertID != "" {
		c.logger.Info("Found existing certificate request", "cert_id", existingCertID)
		// Get existing certificate details
		certDetails, err := c.client.GetCertificate(ctx, existingCertID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get existing certificate details: %w", err)
		}
		certObj = &certDetails
	} else {
		// Create new certificate request
		newCertObj, err := c.createIPCertificate(ctx, ip)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create IP certificate: %w", err)
		}
		certObj = newCertObj
		c.logger.Info("Certificate request created", "cert_id", certObj.ID)
	}

	// First, we need to validate the certificate
	validationDir := os.Getenv("IPSSL_VALIDATION_DIR")
	if validationDir == "" {
		validationDir = "/usr/share/caddy/"
	}

	err = c.ValidateCertificate(ctx, certObj.ID, validationDir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to validate certificate: %w", err)
	}

	// Wait for certificate to be issued
	certDetails, err := c.waitForCertificateIssuance(ctx, certObj.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to wait for certificate issuance: %w", err)
	}
	// Download certificate with cross-signed certificates (intermediate certificates)
	certBundle, err := c.client.DownloadCertificate(ctx, certDetails.ID, true)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to download certificate: %w", err)
	}

	// For auto-generated certificates, we need to get the private key from ZeroSSL
	// This might require a different API call or the private key might be included in the certificate bundle
	keyPEM, err := c.getPrivateKey(ctx, certDetails.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get private key: %w", err)
	}

	// Combine the main certificate with the intermediate certificate chain
	var fullCertChain string
	if certBundle.CertificateCrt != "" {
		fullCertChain = certBundle.CertificateCrt
	}
	if certBundle.CABundleCrt != "" {
		if fullCertChain != "" {
			fullCertChain += "\n"
		}
		fullCertChain += certBundle.CABundleCrt
	}

	c.logger.Info("Certificate downloaded successfully", "cert_id", certDetails.ID, "has_intermediate", certBundle.CABundleCrt != "")
	return []byte(fullCertChain), keyPEM, nil
}

// IsCertificateValid checks if a certificate is valid and not expired
func (c *Client) IsCertificateValid(certPath string, validityDuration time.Duration) (bool, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return false, fmt.Errorf("failed to read certificate file: %w", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return false, fmt.Errorf("failed to decode PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Check if certificate is expired
	if time.Now().After(cert.NotAfter) {
		return false, nil
	}

	// Check if certificate expires within the validity duration
	expiryThreshold := time.Now().Add(validityDuration)
	if cert.NotAfter.Before(expiryThreshold) {
		return false, nil
	}

	return true, nil
}

// createIPCertificate creates a certificate request for IP address using ZeroSSL library
func (c *Client) createIPCertificate(ctx context.Context, ip string) (*zerossl.CertificateObject, error) {
	c.logger.Info("Creating IP certificate using ZeroSSL library", "ip", ip)

	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Parse IP address
	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ip)
	}

	// Create CSR with minimal fields to avoid duplication
	// Use IP address as CommonName
	csrTemplate := &x509.CertificateRequest{
		Subject: pkix.Name{
			Country:      []string{"US"},
			Organization: []string{"IPSSL Client"},
			CommonName:   ip, // Use IP address as CommonName
		},
		// Don't include IPAddresses to avoid duplication
		// ZeroSSL will handle IP validation separately
	}

	// Create CSR
	csrDER, err := x509.CreateCertificateRequest(rand.Reader, csrTemplate, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create CSR: %w", err)
	}

	// Parse CSR back to verify
	csr, err := x509.ParseCertificateRequest(csrDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSR: %w", err)
	}

	// Verify CSR was created successfully
	c.logger.Info("CSR created successfully", "ip", ip, "common_name", csr.Subject.CommonName)

	// Store the private key for later retrieval
	c.privateKeys[ip] = privateKey

	// Create certificate request with ZeroSSL library
	// The library should handle the API call properly
	certObj, err := c.client.CreateCertificate(ctx, csr, 90)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate request: %w", err)
	}

	return &certObj, nil
}

// findExistingCertificate looks for an existing certificate request for the given IP
func (c *Client) findExistingCertificate(ctx context.Context, ip string) (string, error) {
	// List all certificates to find one for this IP
	params := zerossl.ListAllCertificates()
	certificateList, err := c.client.ListCertificates(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to list certificates: %w", err)
	}

	// Look for a certificate with matching CommonName (IP address)
	for _, cert := range certificateList.Results {
		if cert.CommonName == ip {
			c.logger.Info("Found existing certificate", "cert_id", cert.ID, "status", cert.Status)

			// Only return valid certificates (issued or pending validation)
			// Skip cancelled, expired, or failed certificates
			if cert.Status == "issued" || cert.Status == "pending_validation" || cert.Status == "draft" {
				return cert.ID, nil
			} else {
				c.logger.Info("Skipping certificate with invalid status", "cert_id", cert.ID, "status", cert.Status)
			}
		}
	}

	return "", nil // No existing certificate found
}

// getPrivateKey retrieves the private key for the given certificate
func (c *Client) getPrivateKey(ctx context.Context, certID string) ([]byte, error) {
	// Get certificate details to find the IP address
	certDetails, err := c.client.GetCertificate(ctx, certID)
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate details: %w", err)
	}

	// Find the private key for this IP address
	// We'll use the CommonName to match the IP
	ip := certDetails.CommonName

	// First, try to get from in-memory storage
	if privateKey, exists := c.privateKeys[ip]; exists {
		// Convert private key to PEM
		keyPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		})
		return keyPEM, nil
	}

	// If not in memory, try to load from file
	keyPath := filepath.Join("/ipssl", "key.pem")
	if keyPEM, err := os.ReadFile(keyPath); err == nil {
		c.logger.Info("Loaded private key from file", "path", keyPath)
		return keyPEM, nil
	}

	// If still not found, generate a new private key and store it
	c.logger.Info("Generating new private key for IP", "ip", ip)
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new private key: %w", err)
	}

	// Store the private key in memory for future use
	c.privateKeys[ip] = privateKey

	// Convert private key to PEM
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Save the private key to file for persistence
	if err := os.WriteFile(keyPath, keyPEM, 0600); err != nil {
		c.logger.Warn("Failed to save private key to file", "error", err)
	}

	c.logger.Info("Generated and stored new private key", "ip", ip)
	return keyPEM, nil
}

// waitForCertificateIssuance waits for the certificate to be issued
func (c *Client) waitForCertificateIssuance(ctx context.Context, certID string) (*zerossl.CertificateObject, error) {
	c.logger.Info("Waiting for certificate issuance", "cert_id", certID)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			certDetails, err := c.client.GetCertificate(ctx, certID)
			if err != nil {
				c.logger.Error("Failed to get certificate details", "error", err)
				continue
			}

			c.logger.Info("Certificate status", "status", certDetails.Status, "cert_id", certID)

			switch certDetails.Status {
			case "issued":
				return &certDetails, nil
			case "cancelled", "expired":
				return nil, fmt.Errorf("certificate %s failed with status: %s", certID, certDetails.Status)
			case "draft", "pending_validation":
				// Continue waiting
				continue
			default:
				c.logger.Warn("Unknown certificate status", "status", certDetails.Status)
				continue
			}
		}
	}
}

// ValidateCertificate performs domain validation for IP addresses
func (c *Client) ValidateCertificate(ctx context.Context, certID string, validationDir string) error {
	c.logger.Info("Starting certificate validation", "cert_id", certID)
	c.logger.Info("=== ENTERING ValidateCertificate METHOD ===")

	// Get certificate details to check validation method
	c.logger.Info("Getting certificate details", "cert_id", certID)
	certDetails, err := c.client.GetCertificate(ctx, certID)
	if err != nil {
		c.logger.Error("Failed to get certificate details", "error", err)
		return fmt.Errorf("failed to get certificate details: %w", err)
	}
	c.logger.Info("Successfully got certificate details", "cert_id", certID)

	// For IP addresses, we typically need HTTP validation
	// Place validation files in the webroot directory
	c.logger.Info("Certificate validation details", "validation", certDetails.Validation)

	if certDetails.Validation != nil && certDetails.Validation.OtherMethods != nil {
		c.logger.Info("Found validation methods", "methods", certDetails.Validation.OtherMethods)

		for method, validation := range certDetails.Validation.OtherMethods {
			c.logger.Info("Processing validation method", "method", method, "validation", validation)

			if len(validation.FileValidationContent) > 0 {
				// Combine all validation content parts (token, comodoca.com, hash)
				var validationContent string
				for i, content := range validation.FileValidationContent {
					if i > 0 {
						validationContent += "\n"
					}
					validationContent += content
				}

				// Extract filename from the validation URL
				filename := filepath.Base(validation.FileValidationURLHTTP)
				validationPath := filepath.Join(validationDir, ".well-known", "pki-validation", filename)

				// Ensure directory exists
				if err := os.MkdirAll(filepath.Dir(validationPath), 0755); err != nil {
					return fmt.Errorf("failed to create validation directory: %w", err)
				}

				// Write validation file
				if err := os.WriteFile(validationPath, []byte(validationContent), 0644); err != nil {
					return fmt.Errorf("failed to write validation file: %w", err)
				}

				c.logger.Info("Validation file created", "path", validationPath, "content", validationContent)
			} else {
				c.logger.Warn("Skipping validation method", "method", method, "has_content", len(validation.FileValidationContent) > 0)
			}
		}
	} else {
		c.logger.Warn("No validation methods found", "validation_nil", certDetails.Validation == nil, "other_methods_nil", certDetails.Validation != nil && certDetails.Validation.OtherMethods == nil)
	}

	// First, let's try to trigger validation to get the validation details
	c.logger.Info("Attempting to trigger domain validation", "cert_id", certID)
	_, err = c.client.VerifyIdentifiers(ctx, certID, zerossl.HTTPVerification, []string{})
	if err != nil {
		c.logger.Error("Failed to trigger domain validation", "error", err)
		// Don't return error immediately, let's check if we can get validation details
	}

	// Get updated certificate details after triggering validation
	c.logger.Info("Getting updated certificate details", "cert_id", certID)
	updatedCertDetails, err := c.client.GetCertificate(ctx, certID)
	if err != nil {
		return fmt.Errorf("failed to get updated certificate details: %w", err)
	}

	c.logger.Info("Updated certificate validation details", "validation", updatedCertDetails.Validation)

	// Now try to create validation files with updated details
	if updatedCertDetails.Validation != nil && updatedCertDetails.Validation.OtherMethods != nil {
		c.logger.Info("Found updated validation methods", "methods", updatedCertDetails.Validation.OtherMethods)

		for method, validation := range updatedCertDetails.Validation.OtherMethods {
			c.logger.Info("Processing updated validation method", "method", method, "validation", validation)

			if len(validation.FileValidationContent) > 0 {
				// For IP certificates, method is the IP address, not "http"
				// Combine all validation content parts (token, comodoca.com, hash)
				var validationContent string
				for i, content := range validation.FileValidationContent {
					if i > 0 {
						validationContent += "\n"
					}
					validationContent += content
				}

				// Extract filename from the validation URL
				filename := filepath.Base(validation.FileValidationURLHTTP)
				validationPath := filepath.Join(validationDir, ".well-known", "pki-validation", filename)

				c.logger.Info("Creating validation file", "path", validationPath, "content", validationContent)

				// Ensure directory exists
				if err := os.MkdirAll(filepath.Dir(validationPath), 0755); err != nil {
					return fmt.Errorf("failed to create validation directory: %w", err)
				}

				// Write validation file
				if err := os.WriteFile(validationPath, []byte(validationContent), 0644); err != nil {
					return fmt.Errorf("failed to write validation file: %w", err)
				}

				c.logger.Info("Validation file created successfully", "path", validationPath, "content", validationContent)
			} else {
				c.logger.Warn("Skipping updated validation method - no content", "method", method, "has_content", len(validation.FileValidationContent) > 0)
			}
		}
	} else {
		c.logger.Warn("No updated validation methods found", "validation_nil", updatedCertDetails.Validation == nil, "other_methods_nil", updatedCertDetails.Validation != nil && updatedCertDetails.Validation.OtherMethods == nil)
	}

	c.logger.Info("Domain validation process completed", "cert_id", certID)
	return nil
}
