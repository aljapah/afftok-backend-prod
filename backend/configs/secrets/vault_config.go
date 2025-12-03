package secrets

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// ============================================
// SECRETS MANAGER
// ============================================

// SecretType represents different types of secrets
type SecretType string

const (
	SecretTypeAPIKey      SecretType = "api_key"
	SecretTypeDBPassword  SecretType = "db_password"
	SecretTypeJWTSecret   SecretType = "jwt_secret"
	SecretTypeHMACKey     SecretType = "hmac_key"
	SecretTypeEncryption  SecretType = "encryption_key"
	SecretTypeWebhook     SecretType = "webhook_secret"
	SecretTypeTenant      SecretType = "tenant_secret"
)

// Secret represents a managed secret
type Secret struct {
	ID          string     `json:"id"`
	Type        SecretType `json:"type"`
	Name        string     `json:"name"`
	TenantID    string     `json:"tenant_id,omitempty"`
	Version     int        `json:"version"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	RotatedAt   *time.Time `json:"rotated_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	value       string     // Never exposed
}

// ============================================
// VAULT CONFIG
// ============================================

// VaultConfig holds Vault connection configuration
type VaultConfig struct {
	Address   string
	Token     string
	Namespace string
	MountPath string
	Enabled   bool
}

// LoadVaultConfig loads Vault configuration from environment
func LoadVaultConfig() *VaultConfig {
	return &VaultConfig{
		Address:   os.Getenv("VAULT_ADDR"),
		Token:     os.Getenv("VAULT_TOKEN"),
		Namespace: os.Getenv("VAULT_NAMESPACE"),
		MountPath: getEnvOrDefault("VAULT_MOUNT_PATH", "secret"),
		Enabled:   os.Getenv("VAULT_ENABLED") == "true",
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ============================================
// SECRETS MANAGER SERVICE
// ============================================

// SecretsManager manages application secrets
type SecretsManager struct {
	mu           sync.RWMutex
	vaultConfig  *VaultConfig
	secrets      map[string]*Secret
	rotationDays int
	
	// Secret revealer blocking
	blockedPatterns []string
}

// NewSecretsManager creates a new secrets manager
func NewSecretsManager() *SecretsManager {
	return &SecretsManager{
		vaultConfig:  LoadVaultConfig(),
		secrets:      make(map[string]*Secret),
		rotationDays: 30,
		blockedPatterns: []string{
			"password",
			"secret",
			"token",
			"key",
			"credential",
			"auth",
		},
	}
}

// ============================================
// SECRET OPERATIONS
// ============================================

// GenerateSecret generates a new secret
func (m *SecretsManager) GenerateSecret(secretType SecretType, name, tenantID string, expirationDays int) (*Secret, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Generate random value
	value, err := generateSecureRandom(32)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate secret: %w", err)
	}

	now := time.Now()
	var expiresAt *time.Time
	if expirationDays > 0 {
		exp := now.Add(time.Duration(expirationDays) * 24 * time.Hour)
		expiresAt = &exp
	}

	secret := &Secret{
		ID:        fmt.Sprintf("%s_%s_%d", secretType, name, now.UnixNano()),
		Type:      secretType,
		Name:      name,
		TenantID:  tenantID,
		Version:   1,
		CreatedAt: now,
		ExpiresAt: expiresAt,
		value:     value,
	}

	// Store secret
	m.secrets[secret.ID] = secret

	// Store in Vault if enabled
	if m.vaultConfig.Enabled {
		if err := m.storeInVault(secret); err != nil {
			// Log warning but don't fail
			fmt.Printf("Warning: Failed to store secret in Vault: %v\n", err)
		}
	}

	// Return plaintext value only once
	return secret, value, nil
}

// GetSecret retrieves a secret (value is never returned)
func (m *SecretsManager) GetSecret(secretID string) (*Secret, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	secret, exists := m.secrets[secretID]
	if !exists {
		return nil, fmt.Errorf("secret not found: %s", secretID)
	}

	// Update last used
	now := time.Now()
	secret.LastUsedAt = &now

	return secret, nil
}

// ValidateSecret validates a secret value
func (m *SecretsManager) ValidateSecret(secretID, value string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	secret, exists := m.secrets[secretID]
	if !exists {
		return false
	}

	// Check expiration
	if secret.ExpiresAt != nil && time.Now().After(*secret.ExpiresAt) {
		return false
	}

	return secret.value == value
}

// RotateSecret rotates a secret
func (m *SecretsManager) RotateSecret(secretID string) (*Secret, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	secret, exists := m.secrets[secretID]
	if !exists {
		return nil, "", fmt.Errorf("secret not found: %s", secretID)
	}

	// Generate new value
	newValue, err := generateSecureRandom(32)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate new secret: %w", err)
	}

	now := time.Now()
	secret.value = newValue
	secret.Version++
	secret.RotatedAt = &now

	// Update expiration if set
	if secret.ExpiresAt != nil {
		exp := now.Add(time.Duration(m.rotationDays) * 24 * time.Hour)
		secret.ExpiresAt = &exp
	}

	// Update in Vault if enabled
	if m.vaultConfig.Enabled {
		if err := m.storeInVault(secret); err != nil {
			fmt.Printf("Warning: Failed to update secret in Vault: %v\n", err)
		}
	}

	return secret, newValue, nil
}

// DeleteSecret deletes a secret
func (m *SecretsManager) DeleteSecret(secretID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.secrets[secretID]; !exists {
		return fmt.Errorf("secret not found: %s", secretID)
	}

	delete(m.secrets, secretID)

	// Delete from Vault if enabled
	if m.vaultConfig.Enabled {
		m.deleteFromVault(secretID)
	}

	return nil
}

// ============================================
// VAULT INTEGRATION
// ============================================

// storeInVault stores a secret in Vault
func (m *SecretsManager) storeInVault(secret *Secret) error {
	// In production, this would use the Vault API
	// For now, we'll use a placeholder
	if !m.vaultConfig.Enabled {
		return nil
	}

	// Vault API call would go here
	// client.Logical().Write(path, data)
	
	return nil
}

// deleteFromVault deletes a secret from Vault
func (m *SecretsManager) deleteFromVault(secretID string) error {
	if !m.vaultConfig.Enabled {
		return nil
	}

	// Vault API call would go here
	// client.Logical().Delete(path)
	
	return nil
}

// ============================================
// LOG SANITIZATION
// ============================================

// SanitizeForLogging sanitizes data for safe logging
func (m *SecretsManager) SanitizeForLogging(data map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})
	
	for key, value := range data {
		sanitized[key] = m.sanitizeValue(key, value)
	}
	
	return sanitized
}

// sanitizeValue sanitizes a single value
func (m *SecretsManager) sanitizeValue(key string, value interface{}) interface{} {
	keyLower := strings.ToLower(key)
	
	// Check if key matches blocked patterns
	for _, pattern := range m.blockedPatterns {
		if strings.Contains(keyLower, pattern) {
			return "[REDACTED]"
		}
	}
	
	// Handle nested maps
	if nested, ok := value.(map[string]interface{}); ok {
		return m.SanitizeForLogging(nested)
	}
	
	// Handle strings that look like secrets
	if str, ok := value.(string); ok {
		if m.looksLikeSecret(str) {
			return "[REDACTED]"
		}
	}
	
	return value
}

// looksLikeSecret checks if a string looks like a secret
func (m *SecretsManager) looksLikeSecret(s string) bool {
	// Check for common secret patterns
	if len(s) >= 32 && !strings.Contains(s, " ") {
		// Long string without spaces might be a secret
		return true
	}
	
	// Check for base64-encoded secrets
	if len(s) >= 20 && isBase64(s) {
		return true
	}
	
	// Check for JWT tokens
	if strings.Count(s, ".") == 2 && len(s) > 50 {
		return true
	}
	
	return false
}

// isBase64 checks if a string is base64 encoded
func isBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

// ============================================
// TENANT SECRETS
// ============================================

// GetTenantSecrets returns secrets for a specific tenant
func (m *SecretsManager) GetTenantSecrets(tenantID string) []*Secret {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*Secret
	for _, secret := range m.secrets {
		if secret.TenantID == tenantID {
			result = append(result, secret)
		}
	}
	return result
}

// ============================================
// ROTATION SCHEDULER
// ============================================

// CheckRotation checks if any secrets need rotation
func (m *SecretsManager) CheckRotation() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var needsRotation []string
	rotationThreshold := time.Duration(m.rotationDays) * 24 * time.Hour

	for id, secret := range m.secrets {
		var lastRotation time.Time
		if secret.RotatedAt != nil {
			lastRotation = *secret.RotatedAt
		} else {
			lastRotation = secret.CreatedAt
		}

		if time.Since(lastRotation) > rotationThreshold {
			needsRotation = append(needsRotation, id)
		}
	}

	return needsRotation
}

// RotateExpired rotates all expired secrets
func (m *SecretsManager) RotateExpired() ([]string, error) {
	needsRotation := m.CheckRotation()
	var rotated []string

	for _, secretID := range needsRotation {
		if _, _, err := m.RotateSecret(secretID); err == nil {
			rotated = append(rotated, secretID)
		}
	}

	return rotated, nil
}

// ============================================
// STATS
// ============================================

// GetStats returns secrets manager statistics
func (m *SecretsManager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	byType := make(map[string]int)
	expiringSoon := 0
	expired := 0

	for _, secret := range m.secrets {
		byType[string(secret.Type)]++
		
		if secret.ExpiresAt != nil {
			if time.Now().After(*secret.ExpiresAt) {
				expired++
			} else if time.Until(*secret.ExpiresAt) < 7*24*time.Hour {
				expiringSoon++
			}
		}
	}

	return map[string]interface{}{
		"total_secrets":    len(m.secrets),
		"by_type":          byType,
		"expiring_soon":    expiringSoon,
		"expired":          expired,
		"vault_enabled":    m.vaultConfig.Enabled,
		"rotation_days":    m.rotationDays,
		"needs_rotation":   len(m.CheckRotation()),
	}
}

// ============================================
// HELPERS
// ============================================

// generateSecureRandom generates a secure random string
func generateSecureRandom(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// ============================================
// GLOBAL INSTANCE
// ============================================

var (
	secretsManagerInstance *SecretsManager
	secretsManagerOnce     sync.Once
)

// GetSecretsManager returns the global secrets manager
func GetSecretsManager() *SecretsManager {
	secretsManagerOnce.Do(func() {
		secretsManagerInstance = NewSecretsManager()
	})
	return secretsManagerInstance
}

// ============================================
// ROTATION CRON
// ============================================

// StartRotationCron starts the secret rotation cron job
func StartRotationCron(ctx context.Context) {
	manager := GetSecretsManager()
	
	ticker := time.NewTicker(24 * time.Hour) // Check daily
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rotated, err := manager.RotateExpired()
			if err != nil {
				fmt.Printf("Secret rotation error: %v\n", err)
			} else if len(rotated) > 0 {
				fmt.Printf("Rotated %d secrets: %v\n", len(rotated), rotated)
			}
		}
	}
}

// MarshalJSON implements custom JSON marshaling to hide sensitive data
func (s *Secret) MarshalJSON() ([]byte, error) {
	type SecretPublic struct {
		ID         string     `json:"id"`
		Type       SecretType `json:"type"`
		Name       string     `json:"name"`
		TenantID   string     `json:"tenant_id,omitempty"`
		Version    int        `json:"version"`
		CreatedAt  time.Time  `json:"created_at"`
		ExpiresAt  *time.Time `json:"expires_at,omitempty"`
		RotatedAt  *time.Time `json:"rotated_at,omitempty"`
		LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	}

	public := SecretPublic{
		ID:         s.ID,
		Type:       s.Type,
		Name:       s.Name,
		TenantID:   s.TenantID,
		Version:    s.Version,
		CreatedAt:  s.CreatedAt,
		ExpiresAt:  s.ExpiresAt,
		RotatedAt:  s.RotatedAt,
		LastUsedAt: s.LastUsedAt,
	}

	return json.Marshal(public)
}

