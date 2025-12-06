package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/services"
	"github.com/aljapah/afftok-backend-prod/pkg/utils"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT token with enhanced security
func AuthMiddleware() gin.HandlerFunc {
	security := services.NewSecurityService()
	
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Validate header length (prevent DoS with huge headers)
		if len(authHeader) > 2000 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := strings.TrimSpace(parts[1])
		
		// Basic token format validation
		if len(tokenString) < 50 || len(tokenString) > 1000 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		// Validate token
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			// Log failed auth attempt
			security.LogAuditEvent(services.AuditEvent{
				Timestamp: time.Now(),
				EventType: "auth_failed",
				IP:        c.ClientIP(),
				UserAgent: c.Request.UserAgent(),
				Resource:  c.Request.URL.Path,
				Action:    "authenticate",
				Success:   false,
				Details: map[string]interface{}{
					"error": err.Error(),
				},
			})
			
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Check if token is about to expire (warn client)
		// This allows clients to refresh proactively
		if claims.ExpiresAt != nil {
			timeUntilExpiry := time.Until(claims.ExpiresAt.Time)
			if timeUntilExpiry < 5*time.Minute {
				c.Header("X-Token-Expiring-Soon", "true")
			}
		}

		// Set user info in context
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// AdminMiddleware checks if user is admin with audit logging
func AdminMiddleware() gin.HandlerFunc {
	security := services.NewSecurityService()
	
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			// Log unauthorized admin access attempt
			userID := ""
			if id, exists := c.Get("userID"); exists {
				userID = id.(string)
			}
			
			security.LogAuditEvent(services.AuditEvent{
				Timestamp: time.Now(),
				EventType: "admin_access_denied",
				UserID:    userID,
				IP:        c.ClientIP(),
				Resource:  c.Request.URL.Path,
				Action:    c.Request.Method,
				Success:   false,
				Details: map[string]interface{}{
					"role": role,
				},
			})
			
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		
		// Log successful admin access
		userID := ""
		if id, exists := c.Get("userID"); exists {
			userID = id.(string)
		}
		
		security.LogAuditEvent(services.AuditEvent{
			Timestamp: time.Now(),
			EventType: "admin_access",
			UserID:    userID,
			IP:        c.ClientIP(),
			Resource:  c.Request.URL.Path,
			Action:    c.Request.Method,
			Success:   true,
		})
		
		c.Next()
	}
}

// AuthRateLimitMiddleware applies rate limiting to auth endpoints
func AuthRateLimitMiddleware() gin.HandlerFunc {
	security := services.NewSecurityService()
	
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := "auth:" + ip
		
		result := security.CheckRateLimit(key, services.RateLimitAuthPerMinute, time.Minute)
		
		if !result.Allowed {
			security.LogAuditEvent(services.AuditEvent{
				Timestamp: time.Now(),
				EventType: "auth_rate_limited",
				IP:        ip,
				Resource:  c.Request.URL.Path,
				Action:    c.Request.Method,
				Success:   false,
			})
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many authentication attempts",
				"retry_after": result.ResetAt.Sub(time.Now()).Seconds(),
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}
