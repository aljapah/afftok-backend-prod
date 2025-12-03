package services

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ============================================
// TEMPLATE ENGINE SERVICE
// ============================================

// TemplateEngine handles template rendering with placeholders
type TemplateEngine struct {
	placeholderRegex *regexp.Regexp
}

// NewTemplateEngine creates a new template engine
func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		// Match {{key}} or {{object.key}} or {{object.nested.key}}
		placeholderRegex: regexp.MustCompile(`\{\{([a-zA-Z0-9_.]+)\}\}`),
	}
}

// ============================================
// TEMPLATE CONTEXT
// ============================================

// TemplateContext holds all available data for template rendering
type TemplateContext struct {
	// Click data
	Click map[string]interface{} `json:"click,omitempty"`
	
	// Conversion data
	Conversion map[string]interface{} `json:"conversion,omitempty"`
	
	// User offer data
	UserOffer map[string]interface{} `json:"user_offer,omitempty"`
	
	// Offer data
	Offer map[string]interface{} `json:"offer,omitempty"`
	
	// User data
	User map[string]interface{} `json:"user,omitempty"`
	
	// Postback data
	Postback map[string]interface{} `json:"postback,omitempty"`
	
	// System data
	Timestamp     int64  `json:"timestamp"`
	TimestampISO  string `json:"timestamp_iso"`
	CorrelationID string `json:"correlation_id"`
	TaskID        string `json:"task_id"`
	
	// Custom data
	Custom map[string]interface{} `json:"custom,omitempty"`
}

// NewTemplateContext creates a new template context with defaults
func NewTemplateContext() *TemplateContext {
	now := time.Now()
	return &TemplateContext{
		Click:        make(map[string]interface{}),
		Conversion:   make(map[string]interface{}),
		UserOffer:    make(map[string]interface{}),
		Offer:        make(map[string]interface{}),
		User:         make(map[string]interface{}),
		Postback:     make(map[string]interface{}),
		Custom:       make(map[string]interface{}),
		Timestamp:    now.Unix(),
		TimestampISO: now.UTC().Format(time.RFC3339),
	}
}

// ToMap converts the context to a flat map for easy lookup
func (ctx *TemplateContext) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	
	// Add click data with prefix
	for k, v := range ctx.Click {
		result["click."+k] = v
	}
	
	// Add conversion data with prefix
	for k, v := range ctx.Conversion {
		result["conversion."+k] = v
	}
	
	// Add user_offer data with prefix
	for k, v := range ctx.UserOffer {
		result["user_offer."+k] = v
	}
	
	// Add offer data with prefix
	for k, v := range ctx.Offer {
		result["offer."+k] = v
	}
	
	// Add user data with prefix
	for k, v := range ctx.User {
		result["user."+k] = v
	}
	
	// Add postback data with prefix
	for k, v := range ctx.Postback {
		result["postback."+k] = v
	}
	
	// Add custom data with prefix
	for k, v := range ctx.Custom {
		result["custom."+k] = v
	}
	
	// Add system data
	result["timestamp"] = ctx.Timestamp
	result["timestamp_iso"] = ctx.TimestampISO
	result["correlation_id"] = ctx.CorrelationID
	result["task_id"] = ctx.TaskID
	
	return result
}

// ============================================
// TEMPLATE RENDERING
// ============================================

// Render renders a template string with the given context
func (e *TemplateEngine) Render(template string, ctx *TemplateContext) (string, error) {
	if template == "" {
		return "", nil
	}
	
	data := ctx.ToMap()
	
	result := e.placeholderRegex.ReplaceAllStringFunc(template, func(match string) string {
		// Extract key from {{key}}
		key := strings.Trim(match, "{}")
		
		// Look up value
		if value, ok := data[key]; ok {
			return e.formatValue(value)
		}
		
		// Try nested lookup for complex paths
		if value := e.getNestedValue(data, key); value != nil {
			return e.formatValue(value)
		}
		
		// Return empty string for missing values
		return ""
	})
	
	return result, nil
}

// RenderURL renders a URL template
func (e *TemplateEngine) RenderURL(urlTemplate string, ctx *TemplateContext) (string, error) {
	return e.Render(urlTemplate, ctx)
}

// RenderBody renders a body template
func (e *TemplateEngine) RenderBody(bodyTemplate string, ctx *TemplateContext) (string, error) {
	return e.Render(bodyTemplate, ctx)
}

// RenderHeaders renders header templates
func (e *TemplateEngine) RenderHeaders(headers map[string]string, ctx *TemplateContext) (map[string]string, error) {
	result := make(map[string]string)
	
	for key, value := range headers {
		renderedValue, err := e.Render(value, ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to render header %s: %w", key, err)
		}
		result[key] = renderedValue
	}
	
	return result, nil
}

// RenderJSON renders a JSON template and returns parsed JSON
func (e *TemplateEngine) RenderJSON(jsonTemplate string, ctx *TemplateContext) (map[string]interface{}, error) {
	rendered, err := e.Render(jsonTemplate, ctx)
	if err != nil {
		return nil, err
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(rendered), &result); err != nil {
		return nil, fmt.Errorf("failed to parse rendered JSON: %w", err)
	}
	
	return result, nil
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// formatValue converts a value to string representation
func (e *TemplateEngine) formatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case time.Time:
		return v.UTC().Format(time.RFC3339)
	case nil:
		return ""
	default:
		// Try JSON encoding for complex types
		if bytes, err := json.Marshal(v); err == nil {
			return string(bytes)
		}
		return fmt.Sprintf("%v", v)
	}
}

// getNestedValue gets a nested value from a map using dot notation
func (e *TemplateEngine) getNestedValue(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	
	var current interface{} = data
	
	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			var ok bool
			current, ok = v[part]
			if !ok {
				return nil
			}
		default:
			return nil
		}
	}
	
	return current
}

// ============================================
// CONTEXT BUILDERS
// ============================================

// BuildClickContext builds a template context from click data
func BuildClickContext(clickData map[string]interface{}) *TemplateContext {
	ctx := NewTemplateContext()
	ctx.Click = clickData
	return ctx
}

// BuildConversionContext builds a template context from conversion data
func BuildConversionContext(conversionData map[string]interface{}) *TemplateContext {
	ctx := NewTemplateContext()
	ctx.Conversion = conversionData
	return ctx
}

// BuildPostbackContext builds a template context from postback data
func BuildPostbackContext(postbackData map[string]interface{}) *TemplateContext {
	ctx := NewTemplateContext()
	ctx.Postback = postbackData
	return ctx
}

// MergeContexts merges multiple contexts into one
func MergeContexts(contexts ...*TemplateContext) *TemplateContext {
	result := NewTemplateContext()
	
	for _, ctx := range contexts {
		if ctx == nil {
			continue
		}
		
		// Merge maps
		for k, v := range ctx.Click {
			result.Click[k] = v
		}
		for k, v := range ctx.Conversion {
			result.Conversion[k] = v
		}
		for k, v := range ctx.UserOffer {
			result.UserOffer[k] = v
		}
		for k, v := range ctx.Offer {
			result.Offer[k] = v
		}
		for k, v := range ctx.User {
			result.User[k] = v
		}
		for k, v := range ctx.Postback {
			result.Postback[k] = v
		}
		for k, v := range ctx.Custom {
			result.Custom[k] = v
		}
		
		// Use latest non-empty values
		if ctx.CorrelationID != "" {
			result.CorrelationID = ctx.CorrelationID
		}
		if ctx.TaskID != "" {
			result.TaskID = ctx.TaskID
		}
	}
	
	return result
}

// ============================================
// VALIDATION
// ============================================

// ValidateTemplate validates a template string
func (e *TemplateEngine) ValidateTemplate(template string) error {
	if template == "" {
		return nil
	}
	
	// Check for balanced braces
	openCount := strings.Count(template, "{{")
	closeCount := strings.Count(template, "}}")
	
	if openCount != closeCount {
		return fmt.Errorf("unbalanced placeholders: %d opening, %d closing", openCount, closeCount)
	}
	
	// Extract all placeholders
	matches := e.placeholderRegex.FindAllString(template, -1)
	
	// Validate each placeholder format
	for _, match := range matches {
		key := strings.Trim(match, "{}")
		if !isValidPlaceholderKey(key) {
			return fmt.Errorf("invalid placeholder: %s", match)
		}
	}
	
	return nil
}

// isValidPlaceholderKey checks if a placeholder key is valid
func isValidPlaceholderKey(key string) bool {
	if key == "" {
		return false
	}
	
	// Valid keys: alphanumeric, underscore, dot
	validKeyRegex := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_.]*$`)
	return validKeyRegex.MatchString(key)
}

// ExtractPlaceholders extracts all placeholders from a template
func (e *TemplateEngine) ExtractPlaceholders(template string) []string {
	matches := e.placeholderRegex.FindAllString(template, -1)
	
	result := make([]string, len(matches))
	for i, match := range matches {
		result[i] = strings.Trim(match, "{}")
	}
	
	return result
}

// ============================================
// PREDEFINED TEMPLATES
// ============================================

// DefaultConversionTemplate is the default template for conversion webhooks
const DefaultConversionTemplate = `{
	"event": "conversion",
	"conversion_id": "{{conversion.id}}",
	"external_id": "{{conversion.external_id}}",
	"amount": {{conversion.amount}},
	"currency": "{{conversion.currency}}",
	"status": "{{conversion.status}}",
	"user_offer_id": "{{user_offer.id}}",
	"offer_id": "{{offer.id}}",
	"timestamp": {{timestamp}},
	"correlation_id": "{{correlation_id}}"
}`

// DefaultClickTemplate is the default template for click webhooks
const DefaultClickTemplate = `{
	"event": "click",
	"click_id": "{{click.id}}",
	"ip": "{{click.ip}}",
	"user_agent": "{{click.user_agent}}",
	"country": "{{click.country}}",
	"device": "{{click.device}}",
	"user_offer_id": "{{user_offer.id}}",
	"offer_id": "{{offer.id}}",
	"timestamp": {{timestamp}},
	"correlation_id": "{{correlation_id}}"
}`

// DefaultPostbackTemplate is the default template for postback webhooks
const DefaultPostbackTemplate = `{
	"event": "postback",
	"postback_id": "{{postback.id}}",
	"external_id": "{{postback.external_id}}",
	"amount": {{postback.amount}},
	"status": "{{postback.status}}",
	"network_id": "{{postback.network_id}}",
	"timestamp": {{timestamp}},
	"correlation_id": "{{correlation_id}}"
}`

