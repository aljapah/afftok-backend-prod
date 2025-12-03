package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
	TokenID  string    `json:"jti,omitempty"` // Unique token ID for revocation
	jwt.RegisteredClaims
}

var jwtSecret []byte

// Token expiration times
const (
	AccessTokenExpiry  = 1 * time.Hour   // Reduced from 24h for security
	RefreshTokenExpiry = 7 * 24 * time.Hour
)

func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your_jwt_secret_key_change_in_production"
	}
	// Ensure minimum secret length
	if len(secret) < 32 {
		fmt.Println("WARNING: JWT_SECRET should be at least 32 characters")
	}
	jwtSecret = []byte(secret)
}

// generateTokenID creates a unique token ID
func generateTokenID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GenerateToken generates a JWT token for a user with enhanced security
func GenerateToken(userID uuid.UUID, username, email, role string) (string, error) {
	now := time.Now()
	expirationTime := now.Add(AccessTokenExpiry)
	tokenID := generateTokenID()

	claims := &Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Role:     role,
		TokenID:  tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "afftok",
			Subject:   userID.String(),
			ID:        tokenID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GenerateRefreshToken generates a refresh token with enhanced security
func GenerateRefreshToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	expirationTime := now.Add(RefreshTokenExpiry)
	tokenID := generateTokenID()

	claims := &Claims{
		UserID:  userID,
		TokenID: tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "afftok-refresh",
			Subject:   userID.String(),
			ID:        tokenID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims with enhanced security checks
func ValidateToken(tokenString string) (*Claims, error) {
	// Basic length check to prevent DoS
	if len(tokenString) < 50 || len(tokenString) > 1000 {
		return nil, fmt.Errorf("invalid token length")
	}

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Specifically check for HS256
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing algorithm: %v", token.Method.Alg())
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Additional validation: check issuer
	issuer := claims.Issuer
	if issuer != "afftok" && issuer != "afftok-refresh" {
		return nil, fmt.Errorf("invalid token issuer")
	}

	// Check if token is not yet valid (nbf claim)
	if claims.NotBefore != nil && time.Now().Before(claims.NotBefore.Time) {
		return nil, fmt.Errorf("token not yet valid")
	}

	// Check if user ID is valid
	if claims.UserID == uuid.Nil {
		return nil, fmt.Errorf("invalid user ID in token")
	}

	return claims, nil
}

// IsRefreshToken checks if a token is a refresh token
func IsRefreshToken(claims *Claims) bool {
	return claims.Issuer == "afftok-refresh"
}

// GetTokenExpiry returns when the token expires
func GetTokenExpiry(claims *Claims) time.Time {
	if claims.ExpiresAt != nil {
		return claims.ExpiresAt.Time
	}
	return time.Time{}
}
