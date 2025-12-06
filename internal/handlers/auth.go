package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/aljapah/afftok-backend-prod/internal/services"
	"github.com/aljapah/afftok-backend-prod/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db                   *gorm.DB
	observabilityService *services.ObservabilityService
}

type GoogleClaims struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func NewAuthHandler(db *gorm.DB) *AuthHandler {
	return &AuthHandler{
		db:                   db,
		observabilityService: services.NewObservabilityService(),
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	type RegisterRequest struct {
		Username    string `json:"username" binding:"required,min=3,max=50"`
		Email       string `json:"email" binding:"required,email"`
		Password    string `json:"password" binding:"required,min=6"`
		FullName    string `json:"full_name"`
		Role        string `json:"role"`         // "promoter" or "advertiser"
		// Advertiser-specific fields
		CompanyName string `json:"company_name"`
		Phone       string `json:"phone"`
		Website     string `json:"website"`
		Country     string `json:"country"`
	}

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	
	// Security: Additional input validation
	if len(req.Username) > 50 || len(req.Email) > 255 || len(req.Password) > 100 || len(req.FullName) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Input too long"})
		return
	}
	
	// Security: Username format validation (alphanumeric and underscore only)
	for _, r := range req.Username {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username can only contain letters, numbers, and underscores"})
			return
		}
	}

	// Validate role
	role := req.Role
	if role == "" || role == "user" {
		role = "promoter" // Default to promoter
	}
	if role != "promoter" && role != "advertiser" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role. Must be 'promoter' or 'advertiser'"})
		return
	}

	// Advertiser-specific validation
	if role == "advertiser" {
		if req.CompanyName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Company name is required for advertisers"})
			return
		}
	}

	var existingUser models.AfftokUser
	if err := h.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	if err := h.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.AfftokUser{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		Role:         role,
		Status:       "active",
		Points:       0,
		Level:        1,
		UniqueCode:   models.GenerateUniqueCode(), // Generate unique referral code
		// Advertiser-specific fields
		CompanyName:  req.CompanyName,
		Phone:        req.Phone,
		Website:      req.Website,
		Country:      req.Country,
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	accessToken, err := utils.GenerateToken(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	user.PasswordHash = ""

	// Log successful registration
	h.observabilityService.LogAuth(user.ID.String(), user.Username, c.ClientIP(), "register", true, "")

	c.JSON(http.StatusCreated, gin.H{
		"message":       "User registered successfully",
		"user":          user,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	type LoginRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.AfftokUser
	if err := h.db.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error; err != nil {
		h.observabilityService.LogAuth("", req.Username, c.ClientIP(), "login", false, "user_not_found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if user.Status == "suspended" {
		h.observabilityService.LogAuth(user.ID.String(), user.Username, c.ClientIP(), "login", false, "account_suspended")
		c.JSON(http.StatusForbidden, gin.H{"error": "Account is suspended"})
		return
	}

	if !utils.CheckPassword(user.PasswordHash, req.Password) {
		h.observabilityService.LogAuth(user.ID.String(), user.Username, c.ClientIP(), "login", false, "invalid_password")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	accessToken, err := utils.GenerateToken(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	user.PasswordHash = ""

	// Log successful login
	h.observabilityService.LogAuth(user.ID.String(), user.Username, c.ClientIP(), "login", true, "")

	c.JSON(http.StatusOK, gin.H{
		"message":       "Login successful",
		"user":          user,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *AuthHandler) GoogleSignIn(c *gin.Context) {
	type GoogleSignInRequest struct {
		IDToken string `json:"idToken" binding:"required"`
	}

	var req GoogleSignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	googleClaims, err := h.verifyGoogleToken(req.IDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Google token"})
		return
	}

	var user models.AfftokUser
	if err := h.db.Where("email = ?", googleClaims.Email).First(&user).Error; err == nil {
		if user.Status == "suspended" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Account is suspended"})
			return
		}

		accessToken, err := utils.GenerateToken(user.ID, user.Username, user.Email, user.Role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		refreshToken, err := utils.GenerateRefreshToken(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
			return
		}

		user.PasswordHash = ""

		c.JSON(http.StatusOK, gin.H{
			"message":       "Login successful",
			"user":          user,
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		})
		return
	}

	username := generateUsernameFromEmail(googleClaims.Email)
	randomPassword := generateRandomPassword()
	hashedPassword, err := utils.HashPassword(randomPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	newUser := models.AfftokUser{
		ID:           uuid.New(),
		Username:     username,
		Email:        googleClaims.Email,
		PasswordHash: hashedPassword,
		FullName:     googleClaims.Name,
		AvatarURL:    googleClaims.Picture,
		Role:         "user",
		Status:       "active",
		Points:       0,
		Level:        1,
	}

	if err := h.db.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	accessToken, err := utils.GenerateToken(newUser.ID, newUser.Username, newUser.Email, newUser.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(newUser.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	newUser.PasswordHash = ""

	c.JSON(http.StatusCreated, gin.H{
		"message":       "User created and logged in successfully",
		"user":          newUser,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

func (h *AuthHandler) verifyGoogleToken(idToken string) (*GoogleClaims, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	payload := parts[1]
	payload += strings.Repeat("=", (4-len(payload)%4)%4)

	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return nil, err
	}

	var claims GoogleClaims
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return nil, err
	}

	return &claims, nil
}

func generateUsernameFromEmail(email string) string {
	parts := strings.Split(email, "@")
	username := parts[0]
	hash := sha256.Sum256([]byte(email))
	suffix := base64.URLEncoding.EncodeToString(hash[:])[:8]
	return username + "_" + suffix
}

func generateRandomPassword() string {
	return base64.StdEncoding.EncodeToString([]byte(uuid.New().String()))[:16]
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var user models.AfftokUser
	if err := h.db.Preload("UserBadges.Badge").First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var totalOffers int64
	h.db.Model(&models.UserOffer{}).
		Where("user_id = ? AND status = ?", user.ID, "active").
		Count(&totalOffers)

	totalClicks := user.TotalClicks
	totalConversions := user.TotalConversions

	conversionRate := 0.0
	if totalClicks > 0 {
		conversionRate = (float64(totalConversions) / float64(totalClicks)) * 100
	}

	var globalRank int64 = 1
	h.db.Model(&models.AfftokUser{}).
		Where("total_conversions > ?", user.TotalConversions).
		Count(&globalRank)
	globalRank += 1

	user.PasswordHash = ""

	// Get monthly stats
	monthlyClicks := 0
	monthlyConversions := 0
	// TODO: Calculate from clicks/conversions tables with date filter

	stats := gin.H{
		"total_clicks":            totalClicks,
		"total_conversions":       totalConversions,
		"total_earnings":          user.TotalEarnings,
		"total_registered_offers": totalOffers,
		"monthly_clicks":          monthlyClicks,
		"monthly_conversions":     monthlyConversions,
		"global_rank":             globalRank,
		"conversion_rate":         conversionRate,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user": user,
		"stats": stats,
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	type RefreshRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := utils.ValidateToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	var user models.AfftokUser
	if err := h.db.First(&user, "id = ?", claims.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	accessToken, err := utils.GenerateToken(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Logout successful",
	})
}