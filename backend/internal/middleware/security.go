package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/services"
	"github.com/gin-gonic/gin"
)

var securityService = services.NewSecurityService()

// RateLimitMiddleware applies rate limiting to routes
func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		path := c.Request.URL.Path
		
		// Create rate limit key based on IP and path
		key := ip + ":" + path
		
		result := securityService.CheckRateLimit(key, limit, window)
		
		// Set rate limit headers
		c.Header("X-RateLimit-Limit", string(rune(limit)))
		c.Header("X-RateLimit-Remaining", string(rune(result.Remaining)))
		c.Header("X-RateLimit-Reset", result.ResetAt.Format(time.RFC3339))
		
		if !result.Allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":    "Rate limit exceeded",
				"retry_after": result.ResetAt.Sub(time.Now()).Seconds(),
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// ClickRateLimitMiddleware applies click-specific rate limiting
func ClickRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		
		// General click rate limit
		key := "click:" + ip
		result := securityService.CheckRateLimit(key, services.RateLimitClicksPerMinute, time.Minute)
		
		if !result.Allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many click requests",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// BotDetectionMiddleware detects and blocks bot traffic
func BotDetectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		result := securityService.DetectBot(c)
		
		// Store detection result in context
		c.Set("botDetection", result)
		
		// Block high-confidence bots on click endpoints
		if result.IsBot && result.Confidence > 0.8 {
			path := c.Request.URL.Path
			if strings.HasPrefix(path, "/api/c/") {
				// Log the blocked request
				securityService.LogAuditEvent(services.AuditEvent{
					Timestamp: time.Now(),
					EventType: "bot_blocked",
					IP:        c.ClientIP(),
					UserAgent: c.Request.UserAgent(),
					Resource:  path,
					Action:    "click",
					Success:   false,
					Details: map[string]interface{}{
						"reason":     result.Reason,
						"confidence": result.Confidence,
						"risk_score": result.RiskScore,
					},
				})
				
				// Return 403 for obvious bots
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Access denied",
				})
				c.Abort()
				return
			}
		}
		
		c.Next()
	}
}

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")
		
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")
		
		// XSS Protection
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// Referrer Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Content Security Policy
		// Allow inline styles/scripts for HTML pages, restrict for API
		path := c.Request.URL.Path
		if strings.Contains(path, "/promoter/") || strings.Contains(path, "/invite/") || strings.Contains(path, "/join/") || strings.Contains(path, "/r/") || strings.Contains(path, "/@") {
			// HTML pages need inline styles, scripts, fonts and images from any source
			c.Header("Content-Security-Policy", "default-src * 'unsafe-inline' 'unsafe-eval' data: blob:; img-src * data: blob: https:; font-src * data: https://fonts.googleapis.com https://fonts.gstatic.com; style-src * 'unsafe-inline'; script-src * 'unsafe-inline' 'unsafe-eval'; frame-ancestors 'none'")
		} else {
			// API endpoints - strict policy
			c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		}
		
		// Remove server header
		c.Header("Server", "")
		
		c.Next()
	}
}

// RequestValidationMiddleware validates incoming requests
func RequestValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate request
		if err := securityService.ValidateRequest(c); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// AuditLogMiddleware logs all requests for audit purposes
func AuditLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Process request
		c.Next()
		
		// Log after request completes
		userID := ""
		if id, exists := c.Get("userID"); exists {
			userID = id.(string)
		}
		
		// Only log sensitive endpoints
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/admin") ||
			strings.HasPrefix(path, "/api/auth") ||
			strings.HasPrefix(path, "/api/postback") {
			
			securityService.LogAuditEvent(services.AuditEvent{
				Timestamp: start,
				EventType: "api_request",
				UserID:    userID,
				IP:        c.ClientIP(),
				UserAgent: c.Request.UserAgent(),
				Resource:  path,
				Action:    c.Request.Method,
				Success:   c.Writer.Status() < 400,
				Details: map[string]interface{}{
					"status":   c.Writer.Status(),
					"duration": time.Since(start).Milliseconds(),
				},
			})
		}
	}
}

// PostbackSecurityMiddleware validates postback requests
func PostbackSecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		result := securityService.ValidatePostback(c, "")
		
		if !result.Valid {
			securityService.LogAuditEvent(services.AuditEvent{
				Timestamp: time.Now(),
				EventType: "postback_rejected",
				IP:        c.ClientIP(),
				Resource:  c.Request.URL.Path,
				Action:    "postback",
				Success:   false,
				Details: map[string]interface{}{
					"reason": result.Reason,
				},
			})
			
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Invalid postback request",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// SanitizeInputMiddleware sanitizes common input fields
func SanitizeInputMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize query parameters
		for key, values := range c.Request.URL.Query() {
			for i, v := range values {
				values[i] = securityService.SanitizeString(v, 1000)
			}
			c.Request.URL.Query()[key] = values
		}
		
		c.Next()
	}
}

// SecureErrorMiddleware hides internal error details
func SecureErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		// Check for errors
		if len(c.Errors) > 0 {
			// Log the actual error
			for _, err := range c.Errors {
				securityService.LogAuditEvent(services.AuditEvent{
					Timestamp: time.Now(),
					EventType: "error",
					IP:        c.ClientIP(),
					Resource:  c.Request.URL.Path,
					Action:    c.Request.Method,
					Success:   false,
					Details: map[string]interface{}{
						"error": err.Error(),
					},
				})
			}
			
			// Return generic error to client
			if c.Writer.Status() >= 500 {
				c.JSON(c.Writer.Status(), gin.H{
					"error": "An internal error occurred",
				})
			}
		}
	}
}

