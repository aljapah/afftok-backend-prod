package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ============================================
// ADMIN LINK SIGNING HANDLER
// ============================================

// AdminLinkSigningHandler handles link signing management endpoints
type AdminLinkSigningHandler struct {
	linkSigningService *services.LinkSigningService
}

// NewAdminLinkSigningHandler creates a new admin link signing handler
func NewAdminLinkSigningHandler(linkSigningService *services.LinkSigningService) *AdminLinkSigningHandler {
	return &AdminLinkSigningHandler{
		linkSigningService: linkSigningService,
	}
}

// ============================================
// CONFIGURATION
// ============================================

// GetConfig returns the current link signing configuration
// GET /api/admin/link-signing/config
func (h *AdminLinkSigningHandler) GetConfig(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	config := h.linkSigningService.GetConfig()

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"ttl_seconds":        config.TTLSeconds,
			"ttl_human":          formatDuration(config.TTLSeconds),
			"allow_legacy_codes": config.AllowLegacyCodes,
			"secret_length":      config.SecretLength,
			"secret_configured":  config.SecretLength >= 32,
			"replay_cache_count": config.ReplayCacheCount,
		},
		"timestamp": time.Now().UTC(),
	})
}

// ============================================
// STATISTICS
// ============================================

// GetStats returns link signing statistics
// GET /api/admin/security/link-signing/stats
func (h *AdminLinkSigningHandler) GetStats(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	stats := h.linkSigningService.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           stats,
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// TEST LINK
// ============================================

// TestLink tests a link and returns validation details
// GET /api/admin/link-signing/test?link=...
func (h *AdminLinkSigningHandler) TestLink(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	link := c.Query("link")
	if link == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "link query parameter is required",
		})
		return
	}

	result := h.linkSigningService.TestLink(link)

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           result,
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// GENERATE SIGNED LINK
// ============================================

// GenerateSignedLink generates a new signed link for testing
// POST /api/admin/link-signing/generate
func (h *AdminLinkSigningHandler) GenerateSignedLink(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	var req struct {
		TrackingCode string `json:"tracking_code" binding:"required"`
		BaseURL      string `json:"base_url,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid request body: " + err.Error(),
		})
		return
	}

	signedCode := h.linkSigningService.GenerateSignedLink(req.TrackingCode)
	
	fullURL := "/c/" + signedCode
	if req.BaseURL != "" {
		fullURL = req.BaseURL + "/c/" + signedCode
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"tracking_code": req.TrackingCode,
			"signed_code":   signedCode,
			"full_url":      fullURL,
			"expires_in":    h.linkSigningService.GetConfig().TTLSeconds,
		},
		"timestamp": time.Now().UTC(),
	})
}

// ============================================
// SECRET ROTATION
// ============================================

// RotateSecret rotates the signing secret
// POST /api/admin/link-signing/rotate-secret
func (h *AdminLinkSigningHandler) RotateSecret(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	var req struct {
		NewSecret string `json:"new_secret,omitempty"` // Optional, will generate if not provided
	}
	c.ShouldBindJSON(&req)

	var newSecret string
	var err error

	if req.NewSecret != "" {
		if len(req.NewSecret) < 32 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":        false,
				"correlation_id": correlationID,
				"error":          "Secret must be at least 32 characters",
			})
			return
		}
		newSecret = req.NewSecret
	} else {
		newSecret, err = h.linkSigningService.GenerateNewSecret()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success":        false,
				"correlation_id": correlationID,
				"error":          "Failed to generate new secret: " + err.Error(),
			})
			return
		}
	}

	if err := h.linkSigningService.RotateSecret(newSecret); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to rotate secret: " + err.Error(),
		})
		return
	}

	// Note: We return the new secret only if it was generated
	// If provided by admin, we don't echo it back
	response := gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"message":        "Secret rotated successfully. All existing signed links are now invalid.",
		"timestamp":      time.Now().UTC(),
	}

	if req.NewSecret == "" {
		response["new_secret"] = newSecret
		response["warning"] = "⚠️ Save this secret now! It will NOT be shown again. Set it as LINK_SIGNING_SECRET environment variable."
	}

	c.JSON(http.StatusOK, response)
}

// ============================================
// REPLAY CACHE MANAGEMENT
// ============================================

// ClearReplayCache clears the replay protection cache
// POST /api/admin/link-signing/replay/clear
func (h *AdminLinkSigningHandler) ClearReplayCache(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	countBefore := h.linkSigningService.GetReplayCacheCount()

	if err := h.linkSigningService.ClearReplayCache(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to clear replay cache: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"cleared_count": countBefore,
		},
		"message":   "Replay cache cleared successfully",
		"timestamp": time.Now().UTC(),
	})
}

// GetReplayCacheStats returns replay cache statistics
// GET /api/admin/link-signing/replay/stats
func (h *AdminLinkSigningHandler) GetReplayCacheStats(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	count := h.linkSigningService.GetReplayCacheCount()
	config := h.linkSigningService.GetConfig()

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"nonce_count": count,
			"ttl_seconds": config.TTLSeconds,
		},
		"timestamp": time.Now().UTC(),
	})
}

// ============================================
// CONFIGURATION UPDATES
// ============================================

// UpdateConfig updates link signing configuration
// PUT /api/admin/link-signing/config
func (h *AdminLinkSigningHandler) UpdateConfig(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	var req struct {
		TTLSeconds       *int64 `json:"ttl_seconds,omitempty"`
		AllowLegacyCodes *bool  `json:"allow_legacy_codes,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid request body: " + err.Error(),
		})
		return
	}

	if req.TTLSeconds != nil {
		if *req.TTLSeconds < 60 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":        false,
				"correlation_id": correlationID,
				"error":          "TTL must be at least 60 seconds",
			})
			return
		}
		h.linkSigningService.SetTTL(*req.TTLSeconds)
	}

	if req.AllowLegacyCodes != nil {
		h.linkSigningService.SetAllowLegacy(*req.AllowLegacyCodes)
	}

	config := h.linkSigningService.GetConfig()

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           config,
		"message":        "Configuration updated successfully",
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// formatDuration formats seconds into human-readable duration
func formatDuration(seconds int64) string {
	if seconds < 60 {
		return fmt.Sprintf("%d seconds", seconds)
	}
	if seconds < 3600 {
		return fmt.Sprintf("%d minutes", seconds/60)
	}
	if seconds < 86400 {
		return fmt.Sprintf("%d hours", seconds/3600)
	}
	return fmt.Sprintf("%d days", seconds/86400)
}

