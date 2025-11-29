package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/afftok/backend/internal/models"
	"github.com/afftok/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"
)

const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

// GoogleOauthConfig is the configuration for Google OAuth2
var GoogleOauthConfig = &oauth2.Config{
	RedirectURL:  "http://localhost:8080/api/auth/google/callback",
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	Endpoint:     google.Endpoint,
}

// GoogleAuthHandler handles Google OAuth2 authentication
type GoogleAuthHandler struct {
	db *gorm.DB
}

// NewGoogleAuthHandler creates a new GoogleAuthHandler
func NewGoogleAuthHandler(db *gorm.DB) *GoogleAuthHandler {
	return &GoogleAuthHandler{db: db}
}

// BeginLogin redirects the user to Google's login page
func (h *GoogleAuthHandler) BeginLogin(c *gin.Context) {
	// Generate a random state string
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)

	// Save the state in a cookie
	c.SetCookie("oauthstate", state, 3600, "/", "localhost", false, true)

	url := GoogleOauthConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// HandleCallback handles the callback from Google
func (h *GoogleAuthHandler) HandleCallback(c *gin.Context) {
	// Check the state
	oauthState, _ := c.Cookie("oauthstate")
	if c.Request.FormValue("state") != oauthState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
		return
	}

	// Exchange the code for a token
	token, err := GoogleOauthConfig.Exchange(context.Background(), c.Request.FormValue("code"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to exchange token"})
		return
	}

	// Get user info from Google
	response, err := http.Get(oauthGoogleUrlAPI + token.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user info"})
		return
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read user info"})
		return
	}

	// Parse user info
	var userInfo struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.Unmarshal(contents, &userInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse user info"})
		return
	}

	// Check if user exists
	var user models.AfftokUser
	if err := h.db.Where("email = ?", userInfo.Email).First(&user).Error; err != nil {
		// User does not exist, create a new one
		newUser := models.AfftokUser{
			ID:           uuid.New(),
			Username:     userInfo.Email, // Or generate a unique username
			Email:        userInfo.Email,
			PasswordHash: "", // No password for OAuth users
			FullName:     userInfo.Name,
			AvatarURL:    userInfo.Picture,
			Role:         "user",
			Status:       "active",
			Points:       0,
			Level:        1,
		}

		if err := h.db.Create(&newUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}

		user = newUser
	} else {
		// User exists, update avatar and full name if needed
		user.AvatarURL = userInfo.Picture
		user.FullName = userInfo.Name
		h.db.Save(&user)
	}

	// Generate tokens
	accessToken, err := utils.GenerateToken(user.ID, user.Username, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate refresh token"})
		return
	}

	user.PasswordHash = ""

	c.JSON(http.StatusOK, gin.H{
		"message":       "Login successful",
		"user":          user,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}
_
