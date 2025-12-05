package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type InviteHandler struct {
	db *gorm.DB
}

func NewInviteHandler(db *gorm.DB) *InviteHandler {
	return &InviteHandler{db: db}
}

// GetInviteInfo returns information about an invite code (public)
func (h *InviteHandler) GetInviteInfo(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invite code is required"})
		return
	}

	// TODO: Look up team by invite code
	c.JSON(http.StatusOK, gin.H{
		"code":    code,
		"valid":   true,
		"message": "Invite link is valid",
	})
}

// RecordInviteVisit records a visit to an invite link
func (h *InviteHandler) RecordInviteVisit(c *gin.Context) {
	code := c.Param("code")
	// TODO: Record visit statistics
	c.JSON(http.StatusOK, gin.H{
		"message": "Visit recorded",
		"code":    code,
	})
}

// GetMyInviteLink returns the authenticated user's personal invite link
func (h *InviteHandler) GetMyInviteLink(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// TODO: Generate or retrieve user's invite link
	c.JSON(http.StatusOK, gin.H{
		"user_id":     userID,
		"invite_link": "https://afftokapp.com/invite/user123",
		"invite_code": "user123",
	})
}

// CheckPendingInvite checks if user has a pending invite to claim
func (h *InviteHandler) CheckPendingInvite(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// TODO: Check for pending invites
	c.JSON(http.StatusOK, gin.H{
		"user_id":        userID,
		"pending_invite": nil,
		"has_pending":    false,
	})
}

// AutoJoinByInvite automatically joins user to team based on stored invite
func (h *InviteHandler) AutoJoinByInvite(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// TODO: Auto-join logic
	c.JSON(http.StatusOK, gin.H{
		"message": "No pending invite to auto-join",
		"user_id": userID,
	})
}

// ClaimInvite claims a specific invite by ID
func (h *InviteHandler) ClaimInvite(c *gin.Context) {
	inviteID := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// TODO: Claim invite logic
	c.JSON(http.StatusOK, gin.H{
		"message":   "Invite claimed successfully",
		"invite_id": inviteID,
		"user_id":   userID,
	})
}

// ClaimInviteByCode claims an invite using the invite code
func (h *InviteHandler) ClaimInviteByCode(c *gin.Context) {
	code := c.Param("code")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// TODO: Claim by code logic
	c.JSON(http.StatusOK, gin.H{
		"message": "Invite claimed successfully",
		"code":    code,
		"user_id": userID,
	})
}
