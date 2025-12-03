package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/aljapah/afftok-backend-prod/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================
// ADMIN TENANTS HANDLER
// ============================================

// AdminTenantsHandler handles tenant management endpoints
type AdminTenantsHandler struct {
	db            *gorm.DB
	tenantService *services.TenantService
}

// NewAdminTenantsHandler creates a new admin tenants handler
func NewAdminTenantsHandler(db *gorm.DB) *AdminTenantsHandler {
	return &AdminTenantsHandler{
		db:            db,
		tenantService: services.NewTenantService(db),
	}
}

// ============================================
// TENANT CRUD
// ============================================

// GetAllTenants returns all tenants
// GET /api/admin/tenants
func (h *AdminTenantsHandler) GetAllTenants(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	// Parse filters
	var status *models.TenantStatus
	var plan *models.TenantPlan

	if s := c.Query("status"); s != "" {
		st := models.TenantStatus(s)
		status = &st
	}

	if p := c.Query("plan"); p != "" {
		pl := models.TenantPlan(p)
		plan = &pl
	}

	search := c.Query("search")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	tenants, total, err := h.tenantService.ListTenants(status, plan, search, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to fetch tenants: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"tenants": tenants,
			"total":   total,
			"limit":   limit,
			"offset":  offset,
		},
		"timestamp": time.Now().UTC(),
	})
}

// GetTenant returns a single tenant
// GET /api/admin/tenants/:id
func (h *AdminTenantsHandler) GetTenant(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid tenant ID",
		})
		return
	}

	tenant, err := h.tenantService.GetTenant(tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Tenant not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           tenant,
		"timestamp":      time.Now().UTC(),
	})
}

// CreateTenant creates a new tenant
// POST /api/admin/tenants
func (h *AdminTenantsHandler) CreateTenant(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	var req struct {
		Name           string            `json:"name" binding:"required"`
		Slug           string            `json:"slug" binding:"required"`
		AdminEmail     string            `json:"admin_email" binding:"required,email"`
		Plan           models.TenantPlan `json:"plan"`
		LogoURL        string            `json:"logo_url"`
		PrimaryColor   string            `json:"primary_color"`
		SecondaryColor string            `json:"secondary_color"`
		CustomDomain   string            `json:"custom_domain"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid request: " + err.Error(),
		})
		return
	}

	// Set defaults
	if req.Plan == "" {
		req.Plan = models.TenantPlanFree
	}

	tenant := &models.Tenant{
		ID:             uuid.New(),
		Name:           req.Name,
		Slug:           req.Slug,
		AdminEmail:     req.AdminEmail,
		Status:         models.TenantStatusActive,
		Plan:           req.Plan,
		LogoURL:        req.LogoURL,
		PrimaryColor:   req.PrimaryColor,
		SecondaryColor: req.SecondaryColor,
		CustomDomain:   req.CustomDomain,
	}

	if err := h.tenantService.CreateTenant(tenant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to create tenant: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           tenant,
		"message":        "Tenant created successfully",
		"timestamp":      time.Now().UTC(),
	})
}

// UpdateTenant updates a tenant
// PUT /api/admin/tenants/:id
func (h *AdminTenantsHandler) UpdateTenant(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid tenant ID",
		})
		return
	}

	tenant, err := h.tenantService.GetTenant(tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Tenant not found",
		})
		return
	}

	var req struct {
		Name           *string `json:"name"`
		AdminEmail     *string `json:"admin_email"`
		LogoURL        *string `json:"logo_url"`
		FaviconURL     *string `json:"favicon_url"`
		PrimaryColor   *string `json:"primary_color"`
		SecondaryColor *string `json:"secondary_color"`
		CustomDomain   *string `json:"custom_domain"`
		BillingEmail   *string `json:"billing_email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid request: " + err.Error(),
		})
		return
	}

	// Update fields
	if req.Name != nil {
		tenant.Name = *req.Name
	}
	if req.AdminEmail != nil {
		tenant.AdminEmail = *req.AdminEmail
	}
	if req.LogoURL != nil {
		tenant.LogoURL = *req.LogoURL
	}
	if req.FaviconURL != nil {
		tenant.FaviconURL = *req.FaviconURL
	}
	if req.PrimaryColor != nil {
		tenant.PrimaryColor = *req.PrimaryColor
	}
	if req.SecondaryColor != nil {
		tenant.SecondaryColor = *req.SecondaryColor
	}
	if req.CustomDomain != nil {
		tenant.CustomDomain = *req.CustomDomain
	}
	if req.BillingEmail != nil {
		tenant.BillingEmail = *req.BillingEmail
	}

	if err := h.tenantService.UpdateTenant(tenant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to update tenant: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           tenant,
		"message":        "Tenant updated successfully",
		"timestamp":      time.Now().UTC(),
	})
}

// DeleteTenant soft-deletes a tenant
// DELETE /api/admin/tenants/:id
func (h *AdminTenantsHandler) DeleteTenant(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid tenant ID",
		})
		return
	}

	// Prevent deletion of default tenant
	if tenantID == models.DefaultTenantID {
		c.JSON(http.StatusForbidden, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Cannot delete the default tenant",
		})
		return
	}

	if err := h.tenantService.DeleteTenant(tenantID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to delete tenant: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"message":        "Tenant deleted successfully",
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// TENANT STATUS MANAGEMENT
// ============================================

// SuspendTenant suspends a tenant
// POST /api/admin/tenants/:id/suspend
func (h *AdminTenantsHandler) SuspendTenant(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid tenant ID",
		})
		return
	}

	// Prevent suspension of default tenant
	if tenantID == models.DefaultTenantID {
		c.JSON(http.StatusForbidden, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Cannot suspend the default tenant",
		})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)

	if err := h.tenantService.SuspendTenant(tenantID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to suspend tenant: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"message":        "Tenant suspended successfully",
		"timestamp":      time.Now().UTC(),
	})
}

// ActivateTenant activates a tenant
// POST /api/admin/tenants/:id/activate
func (h *AdminTenantsHandler) ActivateTenant(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid tenant ID",
		})
		return
	}

	if err := h.tenantService.ActivateTenant(tenantID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to activate tenant: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"message":        "Tenant activated successfully",
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// TENANT STATS
// ============================================

// GetTenantStats returns stats for a tenant
// GET /api/admin/tenants/:id/stats
func (h *AdminTenantsHandler) GetTenantStats(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid tenant ID",
		})
		return
	}

	stats, err := h.tenantService.GetStats(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to fetch stats: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           stats,
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// TENANT DOMAINS
// ============================================

// GetTenantDomains returns domains for a tenant
// GET /api/admin/tenants/:id/domains
func (h *AdminTenantsHandler) GetTenantDomains(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid tenant ID",
		})
		return
	}

	domains, err := h.tenantService.GetDomains(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to fetch domains: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           domains,
		"timestamp":      time.Now().UTC(),
	})
}

// AddTenantDomain adds a domain to a tenant
// POST /api/admin/tenants/:id/domains
func (h *AdminTenantsHandler) AddTenantDomain(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid tenant ID",
		})
		return
	}

	var req struct {
		Domain    string `json:"domain" binding:"required"`
		IsPrimary bool   `json:"is_primary"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid request: " + err.Error(),
		})
		return
	}

	if err := h.tenantService.AddDomain(tenantID, req.Domain, req.IsPrimary); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to add domain: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"message":        "Domain added successfully",
		"timestamp":      time.Now().UTC(),
	})
}

// RemoveTenantDomain removes a domain from a tenant
// DELETE /api/admin/tenants/:id/domains/:domain
func (h *AdminTenantsHandler) RemoveTenantDomain(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid tenant ID",
		})
		return
	}

	domain := c.Param("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Domain is required",
		})
		return
	}

	if err := h.tenantService.RemoveDomain(tenantID, domain); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to remove domain: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"message":        "Domain removed successfully",
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// TENANT BRANDING
// ============================================

// GetTenantBranding returns branding for a tenant
// GET /api/admin/tenants/:id/branding
func (h *AdminTenantsHandler) GetTenantBranding(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid tenant ID",
		})
		return
	}

	tenant, err := h.tenantService.GetTenant(tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Tenant not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"logo_url":        tenant.LogoURL,
			"favicon_url":     tenant.FaviconURL,
			"primary_color":   tenant.PrimaryColor,
			"secondary_color": tenant.SecondaryColor,
		},
		"timestamp": time.Now().UTC(),
	})
}

// UpdateTenantBranding updates branding for a tenant
// PUT /api/admin/tenants/:id/branding
func (h *AdminTenantsHandler) UpdateTenantBranding(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid tenant ID",
		})
		return
	}

	tenant, err := h.tenantService.GetTenant(tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Tenant not found",
		})
		return
	}

	var req struct {
		LogoURL        *string `json:"logo_url"`
		FaviconURL     *string `json:"favicon_url"`
		PrimaryColor   *string `json:"primary_color"`
		SecondaryColor *string `json:"secondary_color"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid request: " + err.Error(),
		})
		return
	}

	if req.LogoURL != nil {
		tenant.LogoURL = *req.LogoURL
	}
	if req.FaviconURL != nil {
		tenant.FaviconURL = *req.FaviconURL
	}
	if req.PrimaryColor != nil {
		tenant.PrimaryColor = *req.PrimaryColor
	}
	if req.SecondaryColor != nil {
		tenant.SecondaryColor = *req.SecondaryColor
	}

	if err := h.tenantService.UpdateTenant(tenant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to update branding: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"message":        "Branding updated successfully",
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// TENANT PLAN MANAGEMENT
// ============================================

// ChangeTenantPlan changes a tenant's plan
// POST /api/admin/tenants/:id/plan
func (h *AdminTenantsHandler) ChangeTenantPlan(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid tenant ID",
		})
		return
	}

	var req struct {
		Plan models.TenantPlan `json:"plan" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid request: " + err.Error(),
		})
		return
	}

	// Validate plan
	if req.Plan != models.TenantPlanFree && req.Plan != models.TenantPlanPro && req.Plan != models.TenantPlanEnterprise {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid plan. Must be: free, pro, or enterprise",
		})
		return
	}

	if err := h.tenantService.ChangePlan(tenantID, req.Plan); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to change plan: " + err.Error(),
		})
		return
	}

	// Get updated tenant
	tenant, _ := h.tenantService.GetTenant(tenantID)

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           tenant,
		"message":        "Plan changed successfully",
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// TENANT SETTINGS
// ============================================

// GetTenantSettings returns settings for a tenant
// GET /api/admin/tenants/:id/settings
func (h *AdminTenantsHandler) GetTenantSettings(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid tenant ID",
		})
		return
	}

	settings, err := h.tenantService.GetSettings(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to fetch settings: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           settings,
		"timestamp":      time.Now().UTC(),
	})
}

// UpdateTenantSettings updates settings for a tenant
// PUT /api/admin/tenants/:id/settings
func (h *AdminTenantsHandler) UpdateTenantSettings(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid tenant ID",
		})
		return
	}

	var settings models.TenantSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid request: " + err.Error(),
		})
		return
	}

	if err := h.tenantService.UpdateSettings(tenantID, &settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to update settings: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"message":        "Settings updated successfully",
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// TENANT REPORT
// ============================================

// GetTenantsReport returns a summary report of all tenants
// GET /api/admin/tenants/report
func (h *AdminTenantsHandler) GetTenantsReport(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	// Count by status
	var activeCount, suspendedCount, pendingCount int64
	h.db.Model(&models.Tenant{}).Where("status = ?", models.TenantStatusActive).Count(&activeCount)
	h.db.Model(&models.Tenant{}).Where("status = ?", models.TenantStatusSuspended).Count(&suspendedCount)
	h.db.Model(&models.Tenant{}).Where("status = ?", models.TenantStatusPending).Count(&pendingCount)

	// Count by plan
	var freeCount, proCount, enterpriseCount int64
	h.db.Model(&models.Tenant{}).Where("plan = ?", models.TenantPlanFree).Count(&freeCount)
	h.db.Model(&models.Tenant{}).Where("plan = ?", models.TenantPlanPro).Count(&proCount)
	h.db.Model(&models.Tenant{}).Where("plan = ?", models.TenantPlanEnterprise).Count(&enterpriseCount)

	// Recent tenants
	var recentTenants []models.Tenant
	h.db.Order("created_at DESC").Limit(10).Find(&recentTenants)

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"by_status": gin.H{
				"active":    activeCount,
				"suspended": suspendedCount,
				"pending":   pendingCount,
			},
			"by_plan": gin.H{
				"free":       freeCount,
				"pro":        proCount,
				"enterprise": enterpriseCount,
			},
			"total":          activeCount + suspendedCount + pendingCount,
			"recent_tenants": recentTenants,
		},
		"timestamp": time.Now().UTC(),
	})
}

// ============================================
// TENANT AUDIT LOGS
// ============================================

// GetTenantAuditLogs returns audit logs for a tenant
// GET /api/admin/tenants/:id/audit-logs
func (h *AdminTenantsHandler) GetTenantAuditLogs(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid tenant ID",
		})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	logs, err := h.tenantService.GetAuditLogs(tenantID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to fetch audit logs: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           logs,
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// TENANT FEATURES
// ============================================

// GetTenantFeatures returns features for a tenant
// GET /api/admin/tenants/:id/features
func (h *AdminTenantsHandler) GetTenantFeatures(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	tenantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid tenant ID",
		})
		return
	}

	tenant, err := h.tenantService.GetTenant(tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Tenant not found",
		})
		return
	}

	var features models.TenantFeatures
	if tenant.Features != nil {
		json.Unmarshal(tenant.Features, &features)
	} else {
		features = models.GetFeaturesForPlan(tenant.Plan)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"plan":     tenant.Plan,
			"features": features,
		},
		"timestamp": time.Now().UTC(),
	})
}

// ============================================
// PLAN REFERENCE
// ============================================

// GetPlans returns available plans
// GET /api/admin/tenants/plans
func (h *AdminTenantsHandler) GetPlans(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	plans := []map[string]interface{}{
		{
			"name":     "Free",
			"value":    string(models.TenantPlanFree),
			"features": models.GetFeaturesForPlan(models.TenantPlanFree),
			"limits": func() map[string]int {
				u, o, c, a, w := models.GetLimitsForPlan(models.TenantPlanFree)
				return map[string]int{"users": u, "offers": o, "clicks_per_day": c, "api_keys": a, "webhooks": w}
			}(),
		},
		{
			"name":     "Pro",
			"value":    string(models.TenantPlanPro),
			"features": models.GetFeaturesForPlan(models.TenantPlanPro),
			"limits": func() map[string]int {
				u, o, c, a, w := models.GetLimitsForPlan(models.TenantPlanPro)
				return map[string]int{"users": u, "offers": o, "clicks_per_day": c, "api_keys": a, "webhooks": w}
			}(),
		},
		{
			"name":     "Enterprise",
			"value":    string(models.TenantPlanEnterprise),
			"features": models.GetFeaturesForPlan(models.TenantPlanEnterprise),
			"limits": func() map[string]int {
				u, o, c, a, w := models.GetLimitsForPlan(models.TenantPlanEnterprise)
				return map[string]int{"users": u, "offers": o, "clicks_per_day": c, "api_keys": a, "webhooks": w}
			}(),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           plans,
		"timestamp":      time.Now().UTC(),
	})
}

