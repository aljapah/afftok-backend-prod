package handlers

import (
    "net/http"

    "github.com/aljapah/afftok-backend-prod/internal/models"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type NetworkHandler struct {
    db *gorm.DB
}

func NewNetworkHandler(db *gorm.DB) *NetworkHandler {
    return &NetworkHandler{db: db}
}

func (h *NetworkHandler) GetAllNetworks(c *gin.Context) {
    var networks []models.Network

    query := h.db.Select("id, name, description, logo_url, status, created_at")

    status := c.Query("status")
    if status != "" {
        query = query.Where("status = ?", status)
    }

    if err := query.Find(&networks).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch networks"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "networks": networks,
    })
}

func (h *NetworkHandler) GetNetwork(c *gin.Context) {
    networkID := c.Param("id")

    var network models.Network
    if err := h.db.Select("id, name, description, logo_url, status, created_at").First(&network, "id = ?", networkID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Network not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "network": network,
    })
}

func (h *NetworkHandler) CreateNetwork(c *gin.Context) {
    type CreateNetworkRequest struct {
        Name        string `json:"name" binding:"required"`
        Description string `json:"description"`
        LogoURL     string `json:"logo_url"`
    }

    var req CreateNetworkRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    network := models.Network{
        ID:          uuid.New(),
        Name:        req.Name,
        Description: req.Description,
        LogoURL:     req.LogoURL,
        Status:      "active",
    }

    if err := h.db.Create(&network).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create network"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "message": "Network created successfully",
        "network": network,
    })
}

func (h *NetworkHandler) UpdateNetwork(c *gin.Context) {
    networkID := c.Param("id")

    type UpdateNetworkRequest struct {
        Name        string `json:"name"`
        Description string `json:"description"`
        LogoURL     string `json:"logo_url"`
        Status      string `json:"status"`
    }

    var req UpdateNetworkRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    updates := map[string]interface{}{}
    if req.Name != "" {
        updates["name"] = req.Name
    }
    if req.Description != "" {
        updates["description"] = req.Description
    }
    if req.LogoURL != "" {
        updates["logo_url"] = req.LogoURL
    }
    if req.Status != "" {
        updates["status"] = req.Status
    }

    if err := h.db.Model(&models.Network{}).Where("id = ?", networkID).Updates(updates).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update network"})
        return
    }

    var network models.Network
    h.db.First(&network, "id = ?", networkID)

    c.JSON(http.StatusOK, gin.H{
        "message": "Network updated successfully",
        "network": network,
    })
}

func (h *NetworkHandler) DeleteNetwork(c *gin.Context) {
    networkID := c.Param("id")

    id, err := uuid.Parse(networkID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid network ID"})
        return
    }

    if err := h.db.Delete(&models.Network{}, "id = ?", id).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete network"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "Network deleted successfully",
    })
}
