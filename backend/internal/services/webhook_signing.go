package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ============================================
// WEBHOOK SIGNING SERVICE
// ============================================

// WebhookSigningService handles payload signing for webhooks
type WebhookSigningService struct {
	defaultSecret []byte
	jwtSecret     []byte
	jwtIssuer     string
	jwtExpiry     time.Duration
}

// NewWebhookSigningService creates a new webhook signing service
func NewWebhookSigningService() *WebhookSigningService {
	service := &WebhookSigningService{
		defaultSecret: []byte("afftok-webhook-default-secret-change-in-production"),
		jwtSecret:     []byte("afftok-jwt-webhook-secret-change-in-production"),
		jwtIssuer:     "afftok-webhooks",
		jwtExpiry:     5 * time.Minute,
	}

	// Load from environment
	if secret := os.Getenv("WEBHOOK_SIGNING_SECRET"); secret != "" {
		service.defaultSecret = []byte(secret)
	}

	if secret := os.Getenv("WEBHOOK_JWT_SECRET"); secret != "" {
		service.jwtSecret = []byte(secret)
	}

	if issuer := os.Getenv("WEBHOOK_JWT_ISSUER"); issuer != "" {
		service.jwtIssuer = issuer
	}

	return service
}

// ============================================
// HMAC SIGNING
// ============================================

// SignHMAC creates an HMAC-SHA256 signature for the payload
func (s *WebhookSigningService) SignHMAC(payload []byte, secret string) string {
	key := s.defaultSecret
	if secret != "" {
		key = []byte(secret)
	}

	h := hmac.New(sha256.New, key)
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyHMAC verifies an HMAC-SHA256 signature
func (s *WebhookSigningService) VerifyHMAC(payload []byte, signature, secret string) bool {
	expected := s.SignHMAC(payload, secret)
	return hmac.Equal([]byte(expected), []byte(signature))
}

// CreateHMACHeaders creates headers for HMAC signed requests
func (s *WebhookSigningService) CreateHMACHeaders(payload []byte, secret string) map[string]string {
	signature := s.SignHMAC(payload, secret)
	timestamp := time.Now().Unix()

	return map[string]string{
		"X-Afftok-Signature":   signature,
		"X-Afftok-Timestamp":   fmt.Sprintf("%d", timestamp),
		"X-Afftok-Algorithm":   "HMAC-SHA256",
		"Content-Type":         "application/json",
	}
}

// ============================================
// JWT SIGNING
// ============================================

// WebhookJWTClaims represents the claims in a webhook JWT
type WebhookJWTClaims struct {
	TaskID       string     `json:"task_id"`
	AdvertiserID *uuid.UUID `json:"advertiser_id,omitempty"`
	PipelineID   uuid.UUID  `json:"pipeline_id"`
	ExecutionID  uuid.UUID  `json:"execution_id"`
	StepIndex    int        `json:"step_index"`
	Timestamp    int64      `json:"timestamp"`
	jwt.RegisteredClaims
}

// SignJWT creates a JWT token for the webhook
func (s *WebhookSigningService) SignJWT(taskID string, advertiserID *uuid.UUID, pipelineID, executionID uuid.UUID, stepIndex int, secret string) (string, error) {
	key := s.jwtSecret
	if secret != "" {
		key = []byte(secret)
	}

	now := time.Now()
	claims := WebhookJWTClaims{
		TaskID:       taskID,
		AdvertiserID: advertiserID,
		PipelineID:   pipelineID,
		ExecutionID:  executionID,
		StepIndex:    stepIndex,
		Timestamp:    now.Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.jwtIssuer,
			Subject:   taskID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.jwtExpiry)),
			NotBefore: jwt.NewNumericDate(now.Add(-1 * time.Minute)), // Allow 1 min clock skew
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(key)
}

// VerifyJWT verifies a JWT token
func (s *WebhookSigningService) VerifyJWT(tokenString, secret string) (*WebhookJWTClaims, error) {
	key := s.jwtSecret
	if secret != "" {
		key = []byte(secret)
	}

	token, err := jwt.ParseWithClaims(tokenString, &WebhookJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	if claims, ok := token.Claims.(*WebhookJWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid JWT token")
}

// CreateJWTHeaders creates headers for JWT signed requests
func (s *WebhookSigningService) CreateJWTHeaders(taskID string, advertiserID *uuid.UUID, pipelineID, executionID uuid.UUID, stepIndex int, secret string) (map[string]string, error) {
	token, err := s.SignJWT(taskID, advertiserID, pipelineID, executionID, stepIndex, secret)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
	}, nil
}

// ============================================
// UNIFIED SIGNING
// ============================================

// SigningMode represents the type of signing to use
type SigningMode string

const (
	SigningModeNone SigningMode = "none"
	SigningModeHMAC SigningMode = "hmac"
	SigningModeJWT  SigningMode = "jwt"
)

// SignRequest signs a webhook request based on the mode
type SignedRequest struct {
	Headers   map[string]string
	Signature string
	Token     string
}

// SignRequest creates a signed request based on the signing mode
func (s *WebhookSigningService) SignRequest(
	mode SigningMode,
	payload []byte,
	secret string,
	taskID string,
	advertiserID *uuid.UUID,
	pipelineID, executionID uuid.UUID,
	stepIndex int,
) (*SignedRequest, error) {
	result := &SignedRequest{
		Headers: make(map[string]string),
	}

	switch mode {
	case SigningModeNone:
		result.Headers["Content-Type"] = "application/json"
		return result, nil

	case SigningModeHMAC:
		result.Signature = s.SignHMAC(payload, secret)
		result.Headers = s.CreateHMACHeaders(payload, secret)
		return result, nil

	case SigningModeJWT:
		headers, err := s.CreateJWTHeaders(taskID, advertiserID, pipelineID, executionID, stepIndex, secret)
		if err != nil {
			return nil, err
		}
		result.Headers = headers
		result.Token = headers["Authorization"]
		return result, nil

	default:
		return nil, fmt.Errorf("unknown signing mode: %s", mode)
	}
}

// ============================================
// VERIFICATION
// ============================================

// VerifyRequest verifies a signed webhook request
func (s *WebhookSigningService) VerifyRequest(
	mode SigningMode,
	payload []byte,
	signature string,
	token string,
	secret string,
) (bool, error) {
	switch mode {
	case SigningModeNone:
		return true, nil

	case SigningModeHMAC:
		return s.VerifyHMAC(payload, signature, secret), nil

	case SigningModeJWT:
		// Remove "Bearer " prefix if present
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
		_, err := s.VerifyJWT(token, secret)
		return err == nil, err

	default:
		return false, fmt.Errorf("unknown signing mode: %s", mode)
	}
}

// ============================================
// WEBHOOK SIGNATURE HEADER CONSTANTS
// ============================================

const (
	HeaderSignature   = "X-Afftok-Signature"
	HeaderTimestamp   = "X-Afftok-Timestamp"
	HeaderAlgorithm   = "X-Afftok-Algorithm"
	HeaderWebhookID   = "X-Afftok-Webhook-ID"
	HeaderPipelineID  = "X-Afftok-Pipeline-ID"
	HeaderExecutionID = "X-Afftok-Execution-ID"
	HeaderStepIndex   = "X-Afftok-Step-Index"
	HeaderRetryCount  = "X-Afftok-Retry-Count"
)

// AddWebhookMetadataHeaders adds metadata headers to a request
func (s *WebhookSigningService) AddWebhookMetadataHeaders(
	headers map[string]string,
	taskID string,
	pipelineID, executionID uuid.UUID,
	stepIndex, retryCount int,
) map[string]string {
	if headers == nil {
		headers = make(map[string]string)
	}

	headers[HeaderWebhookID] = taskID
	headers[HeaderPipelineID] = pipelineID.String()
	headers[HeaderExecutionID] = executionID.String()
	headers[HeaderStepIndex] = fmt.Sprintf("%d", stepIndex)
	headers[HeaderRetryCount] = fmt.Sprintf("%d", retryCount)

	return headers
}

