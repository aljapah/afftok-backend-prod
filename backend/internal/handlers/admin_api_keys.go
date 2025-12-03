package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/middleware"
	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/aljapah/afftok-backend-prod/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ============================================
// ADMIN API KEYS HANDLER
// ============================================

// AdminAPIKeysHandler handles API key management endpoints
type AdminAPIKeysHandler struct {
	apiKeyService *services.APIKeyService
}

// NewAdminAPIKeysHandler creates a new admin API keys handler
func NewAdminAPIKeysHandler(apiKeyService *services.APIKeyService) *AdminAPIKeysHandler {
	return &AdminAPIKeysHandler{
		apiKeyService: apiKeyService,
	}
}

// ============================================
// LIST API KEYS
// ============================================

// GetAllAPIKeys returns all API keys (admin only)
// GET /api/admin/api-keys?page=1&limit=20
func (h *AdminAPIKeysHandler) GetAllAPIKeys(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	keys, total, err := h.apiKeyService.GetAllAPIKeys(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to fetch API keys: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"api_keys":    keys,
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
		"timestamp": time.Now().UTC(),
	})
}

// GetAPIKeyByID returns a single API key by ID (masked)
// GET /api/admin/api-keys/:id
func (h *AdminAPIKeysHandler) GetAPIKeyByID(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid API key ID",
		})
		return
	}

	key, err := h.apiKeyService.GetAPIKeyByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "API key not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           key,
		"timestamp":      time.Now().UTC(),
	})
}

// GetAPIKeysByAdvertiser returns all API keys for a specific advertiser
// GET /api/admin/advertisers/:id/api-keys
func (h *AdminAPIKeysHandler) GetAPIKeysByAdvertiser(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	advertiserIDStr := c.Param("id")
	advertiserID, err := uuid.Parse(advertiserIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid advertiser ID",
		})
		return
	}

	keys, err := h.apiKeyService.GetAPIKeysByAdvertiser(advertiserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to fetch API keys: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"api_keys":      keys,
			"count":         len(keys),
			"advertiser_id": advertiserID,
		},
		"timestamp": time.Now().UTC(),
	})
}

// ============================================
// CREATE API KEY
// ============================================

// CreateAPIKey creates a new API key for an advertiser
// POST /api/admin/advertisers/:id/api-keys
// Body: { "name": "MyKey", "permissions": [], "allowed_ips": [], "expires_in_days": 0 }
func (h *AdminAPIKeysHandler) CreateAPIKey(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	advertiserIDStr := c.Param("id")
	advertiserID, err := uuid.Parse(advertiserIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid advertiser ID",
		})
		return
	}

	var req models.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid request body: " + err.Error(),
		})
		return
	}

	response, err := h.apiKeyService.GenerateAPIKey(advertiserID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to create API key: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           response,
		"message":        "⚠️ IMPORTANT: Copy the API key now. It will NOT be shown again!",
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// ROTATE API KEY
// ============================================

// RotateAPIKey rotates an existing API key
// POST /api/admin/api-keys/:id/rotate
func (h *AdminAPIKeysHandler) RotateAPIKey(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid API key ID",
		})
		return
	}

	response, err := h.apiKeyService.RotateAPIKey(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to rotate API key: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           response,
		"message":        "⚠️ Old key revoked. Copy the new API key now. It will NOT be shown again!",
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// REVOKE API KEY
// ============================================

// RevokeAPIKey revokes an API key
// POST /api/admin/api-keys/:id/revoke
func (h *AdminAPIKeysHandler) RevokeAPIKey(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid API key ID",
		})
		return
	}

	if err := h.apiKeyService.RevokeAPIKey(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to revoke API key: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"message":        "API key revoked successfully",
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// IP MANAGEMENT
// ============================================

// AddAllowedIP adds an IP to the API key's allowlist
// POST /api/admin/api-keys/:id/allow-ip
// Body: { "ip": "192.168.1.1" }
func (h *AdminAPIKeysHandler) AddAllowedIP(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid API key ID",
		})
		return
	}

	var req struct {
		IP string `json:"ip" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid request body",
		})
		return
	}

	if err := h.apiKeyService.AddAllowedIP(id, req.IP); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to add IP: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"message":        "IP added to allowlist",
		"ip":             req.IP,
		"timestamp":      time.Now().UTC(),
	})
}

// RemoveAllowedIP removes an IP from the API key's allowlist
// POST /api/admin/api-keys/:id/deny-ip
// Body: { "ip": "192.168.1.1" }
func (h *AdminAPIKeysHandler) RemoveAllowedIP(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid API key ID",
		})
		return
	}

	var req struct {
		IP string `json:"ip" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid request body",
		})
		return
	}

	if err := h.apiKeyService.RemoveAllowedIP(id, req.IP); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to remove IP: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"message":        "IP removed from allowlist",
		"ip":             req.IP,
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// STATISTICS & REPORTS
// ============================================

// GetAPIKeyStats returns API key statistics
// GET /api/admin/security/api-keys/report
func (h *AdminAPIKeysHandler) GetAPIKeyStats(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	stats, err := h.apiKeyService.GetAPIKeyStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to get stats: " + err.Error(),
		})
		return
	}

	// Get middleware metrics
	middlewareMetrics := middleware.GetAPIKeyMetrics()

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"keys": stats,
			"authentication": gin.H{
				"total_attempts":   middlewareMetrics.TotalAttempts,
				"success_attempts": middlewareMetrics.SuccessAttempts,
				"failed_attempts":  middlewareMetrics.FailedAttempts,
				"blocked_attempts": middlewareMetrics.BlockedAttempts,
				"rate_limit_blocks": middlewareMetrics.RateLimitBlocks,
				"ip_violations":    middlewareMetrics.IPViolations,
			},
		},
		"timestamp": time.Now().UTC(),
	})
}

