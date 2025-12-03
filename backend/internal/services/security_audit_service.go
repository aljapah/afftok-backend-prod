package services

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================
// SECURITY AUDIT TYPES
// ============================================

// SecurityFindingSeverity represents the severity of a finding
type SecurityFindingSeverity string

const (
	FindingSeverityLow      SecurityFindingSeverity = "low"
	FindingSeverityMedium   SecurityFindingSeverity = "medium"
	FindingSeverityHigh     SecurityFindingSeverity = "high"
	FindingSeverityCritical SecurityFindingSeverity = "critical"
)

// SecurityFindingComponent represents the component with the finding
type SecurityFindingComponent string

const (
	ComponentAPIKeys   SecurityFindingComponent = "api_keys"
	ComponentWebhooks  SecurityFindingComponent = "webhooks"
	ComponentTenant    SecurityFindingComponent = "tenant"
	ComponentAuth      SecurityFindingComponent = "auth"
	ComponentEdge      SecurityFindingComponent = "edge"
	ComponentGeoRules  SecurityFindingComponent = "geo_rules"
	ComponentRateLimit SecurityFindingComponent = "rate_limit"
)

// ============================================
// SECURITY FINDING
// ============================================

// SecurityFinding represents a security audit finding
type SecurityFinding struct {
	ID             string                   `json:"id"`
	Severity       SecurityFindingSeverity  `json:"severity"`
	Component      SecurityFindingComponent `json:"component"`
	Title          string                   `json:"title"`
	Description    string                   `json:"description"`
	Recommendation string                   `json:"recommendation"`
	EntityID       string                   `json:"entity_id,omitempty"`
	EntityType     string                   `json:"entity_type,omitempty"`
	Metadata       map[string]interface{}   `json:"metadata,omitempty"`
	Timestamp      time.Time                `json:"timestamp"`
	Acknowledged   bool                     `json:"acknowledged"`
}

// ============================================
// SECURITY AUDIT REPORT
// ============================================

// SecurityAuditReport holds the complete audit report
type SecurityAuditReport struct {
	ID             string                     `json:"id"`
	Status         string                     `json:"status"` // passed, failed, warning
	StartTime      time.Time                  `json:"start_time"`
	EndTime        time.Time                  `json:"end_time"`
	DurationMs     int64                      `json:"duration_ms"`
	Summary        *SecurityAuditSummary      `json:"summary"`
	Findings       []SecurityFinding          `json:"findings"`
	ChecksRun      []string                   `json:"checks_run"`
	Metadata       map[string]interface{}     `json:"metadata,omitempty"`
}

// SecurityAuditSummary provides a summary of the audit
type SecurityAuditSummary struct {
	TotalChecks      int `json:"total_checks"`
	TotalFindings    int `json:"total_findings"`
	CriticalFindings int `json:"critical_findings"`
	HighFindings     int `json:"high_findings"`
	MediumFindings   int `json:"medium_findings"`
	LowFindings      int `json:"low_findings"`
	ByComponent      map[string]int `json:"by_component"`
}

// ============================================
// SECURITY AUDIT SERVICE
// ============================================

// SecurityAuditService performs security audits
type SecurityAuditService struct {
	mu            sync.RWMutex
	db            *gorm.DB
	observability *ObservabilityService
	
	// History
	reports       []*SecurityAuditReport
	maxReports    int
	
	// Metrics
	totalAudits   int64
	totalFindings int64
}

// NewSecurityAuditService creates a new security audit service
func NewSecurityAuditService(db *gorm.DB) *SecurityAuditService {
	return &SecurityAuditService{
		db:            db,
		observability: NewObservabilityService(),
		reports:       make([]*SecurityAuditReport, 0),
		maxReports:    50,
	}
}

// ============================================
// MAIN AUDIT RUNNER
// ============================================

// RunFullAudit runs all security audit checks
func (s *SecurityAuditService) RunFullAudit() (*SecurityAuditReport, error) {
	report := &SecurityAuditReport{
		ID:        uuid.New().String(),
		StartTime: time.Now(),
		Findings:  make([]SecurityFinding, 0),
		ChecksRun: make([]string, 0),
	}

	atomic.AddInt64(&s.totalAudits, 1)

	s.observability.Log(LogEvent{
		Category: LogCategorySystemEvent,
		Level:    LogLevelInfo,
		Message:  "Security audit started",
		Metadata: map[string]interface{}{
			"report_id": report.ID,
		},
	})

	// Run all checks
	s.auditAPIKeys(report)
	s.auditWebhooks(report)
	s.auditTenants(report)
	s.auditAuth(report)
	s.auditGeoRules(report)
	s.auditRateLimits(report)

	// Finalize report
	report.EndTime = time.Now()
	report.DurationMs = report.EndTime.Sub(report.StartTime).Milliseconds()
	report.Summary = s.calculateSummary(report)
	report.Status = s.determineStatus(report)

	// Store report
	s.storeReport(report)

	s.observability.Log(LogEvent{
		Category: LogCategorySystemEvent,
		Level:    LogLevelInfo,
		Message:  "Security audit completed",
		Metadata: map[string]interface{}{
			"report_id": report.ID,
			"status":    report.Status,
			"findings":  len(report.Findings),
		},
	})

	return report, nil
}

// RunQuickAudit runs a subset of critical checks
func (s *SecurityAuditService) RunQuickAudit() (*SecurityAuditReport, error) {
	report := &SecurityAuditReport{
		ID:        uuid.New().String(),
		StartTime: time.Now(),
		Findings:  make([]SecurityFinding, 0),
		ChecksRun: make([]string, 0),
	}

	// Run only critical checks
	s.auditAPIKeysQuick(report)
	s.auditAuthQuick(report)

	report.EndTime = time.Now()
	report.DurationMs = report.EndTime.Sub(report.StartTime).Milliseconds()
	report.Summary = s.calculateSummary(report)
	report.Status = s.determineStatus(report)

	return report, nil
}

// ============================================
// API KEY AUDITS
// ============================================

func (s *SecurityAuditService) auditAPIKeys(report *SecurityAuditReport) {
	report.ChecksRun = append(report.ChecksRun, "api_keys_full")

	// Check for keys without rotation > 90 days
	var oldKeys []models.AdvertiserAPIKey
	rotationThreshold := time.Now().AddDate(0, 0, -90)
	
	s.db.Where("created_at < ? AND status = ?", rotationThreshold, "active").
		Find(&oldKeys)

	for _, key := range oldKeys {
		report.Findings = append(report.Findings, SecurityFinding{
			ID:             uuid.New().String(),
			Severity:       FindingSeverityMedium,
			Component:      ComponentAPIKeys,
			Title:          "API Key Not Rotated",
			Description:    fmt.Sprintf("API key '%s' has not been rotated in over 90 days", key.Name),
			Recommendation: "Rotate this API key to maintain security best practices",
			EntityID:       key.ID.String(),
			EntityType:     "api_key",
			Metadata: map[string]interface{}{
				"key_name":   key.Name,
				"created_at": key.CreatedAt,
				"days_old":   int(time.Since(key.CreatedAt).Hours() / 24),
			},
			Timestamp: time.Now(),
		})
	}

	// Check for keys with high failure rate
	var failingKeys []struct {
		KeyID       uuid.UUID
		Name        string
		FailedCount int64
		TotalCount  int64
	}

	s.db.Raw(`
		SELECT ak.id as key_id, ak.name, 
			COUNT(CASE WHEN al.success = false THEN 1 END) as failed_count,
			COUNT(*) as total_count
		FROM advertiser_api_keys ak
		LEFT JOIN api_key_usage_logs al ON al.api_key_id = ak.id
		WHERE al.created_at > NOW() - INTERVAL '24 hours'
		GROUP BY ak.id, ak.name
		HAVING COUNT(CASE WHEN al.success = false THEN 1 END) > 10
	`).Scan(&failingKeys)

	for _, key := range failingKeys {
		failRate := float64(key.FailedCount) / float64(key.TotalCount) * 100
		if failRate > 20 {
			report.Findings = append(report.Findings, SecurityFinding{
				ID:             uuid.New().String(),
				Severity:       FindingSeverityHigh,
				Component:      ComponentAPIKeys,
				Title:          "High API Key Failure Rate",
				Description:    fmt.Sprintf("API key '%s' has %.1f%% failure rate in last 24h", key.Name, failRate),
				Recommendation: "Investigate failed requests - possible abuse or misconfiguration",
				EntityID:       key.KeyID.String(),
				EntityType:     "api_key",
				Metadata: map[string]interface{}{
					"failed_count": key.FailedCount,
					"total_count":  key.TotalCount,
					"failure_rate": failRate,
				},
				Timestamp: time.Now(),
			})
		}
	}

	// Check for keys with no IP restrictions
	var unrestrictedKeys []models.AdvertiserAPIKey
	s.db.Where("allowed_ips IS NULL OR allowed_ips = '[]' OR allowed_ips = 'null'").
		Where("status = ?", "active").
		Find(&unrestrictedKeys)

	for _, key := range unrestrictedKeys {
		report.Findings = append(report.Findings, SecurityFinding{
			ID:             uuid.New().String(),
			Severity:       FindingSeverityLow,
			Component:      ComponentAPIKeys,
			Title:          "API Key Without IP Restriction",
			Description:    fmt.Sprintf("API key '%s' has no IP restrictions", key.Name),
			Recommendation: "Consider adding IP allowlist for production keys",
			EntityID:       key.ID.String(),
			EntityType:     "api_key",
			Timestamp:      time.Now(),
		})
	}
}

func (s *SecurityAuditService) auditAPIKeysQuick(report *SecurityAuditReport) {
	report.ChecksRun = append(report.ChecksRun, "api_keys_quick")

	// Just check for compromised or revoked keys still being used
	var revokedButUsed []struct {
		KeyID     uuid.UUID
		Name      string
		UseCount  int64
	}

	s.db.Raw(`
		SELECT ak.id as key_id, ak.name, COUNT(al.id) as use_count
		FROM advertiser_api_keys ak
		JOIN api_key_usage_logs al ON al.api_key_id = ak.id
		WHERE ak.status = 'revoked'
		AND al.created_at > NOW() - INTERVAL '1 hour'
		GROUP BY ak.id, ak.name
	`).Scan(&revokedButUsed)

	for _, key := range revokedButUsed {
		report.Findings = append(report.Findings, SecurityFinding{
			ID:             uuid.New().String(),
			Severity:       FindingSeverityCritical,
			Component:      ComponentAPIKeys,
			Title:          "Revoked API Key Still Being Used",
			Description:    fmt.Sprintf("Revoked API key '%s' has %d attempted uses in last hour", key.Name, key.UseCount),
			Recommendation: "Investigate source of requests - possible key compromise",
			EntityID:       key.KeyID.String(),
			EntityType:     "api_key",
			Timestamp:      time.Now(),
		})
	}
}

// ============================================
// WEBHOOK AUDITS
// ============================================

func (s *SecurityAuditService) auditWebhooks(report *SecurityAuditReport) {
	report.ChecksRun = append(report.ChecksRun, "webhooks")

	// Check for webhooks not using HTTPS
	var insecureWebhooks []models.WebhookPipeline
	s.db.Where("status = ?", "active").Find(&insecureWebhooks)

	for _, webhook := range insecureWebhooks {
		// Check steps for non-HTTPS URLs
		for _, step := range webhook.Steps {
			if len(step.URL) > 0 && step.URL[:5] != "https" {
				report.Findings = append(report.Findings, SecurityFinding{
					ID:             uuid.New().String(),
					Severity:       FindingSeverityHigh,
					Component:      ComponentWebhooks,
					Title:          "Webhook Using HTTP",
					Description:    fmt.Sprintf("Webhook '%s' step uses insecure HTTP", webhook.Name),
					Recommendation: "Use HTTPS for all webhook endpoints",
					EntityID:       webhook.ID.String(),
					EntityType:     "webhook_pipeline",
					Timestamp:      time.Now(),
				})
			}
		}
	}

	// Check for webhooks without signing
	for _, webhook := range insecureWebhooks {
		for _, step := range webhook.Steps {
			if step.SignatureMode == "none" {
				report.Findings = append(report.Findings, SecurityFinding{
					ID:             uuid.New().String(),
					Severity:       FindingSeverityMedium,
					Component:      ComponentWebhooks,
					Title:          "Webhook Without Signing",
					Description:    fmt.Sprintf("Webhook '%s' step has no signature verification", webhook.Name),
					Recommendation: "Enable HMAC or JWT signing for webhook payloads",
					EntityID:       webhook.ID.String(),
					EntityType:     "webhook_pipeline",
					Timestamp:      time.Now(),
				})
			}
		}
	}
}

// ============================================
// TENANT AUDITS
// ============================================

func (s *SecurityAuditService) auditTenants(report *SecurityAuditReport) {
	report.ChecksRun = append(report.ChecksRun, "tenants")

	// Check for tenants with weak settings
	var tenants []models.Tenant
	s.db.Where("status = ?", "active").Find(&tenants)

	for _, tenant := range tenants {
		// Check rate limits
		if tenant.MaxClicksPerDay > 1000000 {
			report.Findings = append(report.Findings, SecurityFinding{
				ID:             uuid.New().String(),
				Severity:       FindingSeverityLow,
				Component:      ComponentTenant,
				Title:          "High Tenant Rate Limit",
				Description:    fmt.Sprintf("Tenant '%s' has very high rate limit: %d clicks/day", tenant.Name, tenant.MaxClicksPerDay),
				Recommendation: "Review if this rate limit is appropriate",
				EntityID:       tenant.ID.String(),
				EntityType:     "tenant",
				Timestamp:      time.Now(),
			})
		}

		// Check for disabled security features (parse from JSON)
		var settings models.TenantSettings
		if tenant.Settings != nil {
			_ = json.Unmarshal(tenant.Settings, &settings)
		}
		
		if !settings.EnableGeoRules {
			report.Findings = append(report.Findings, SecurityFinding{
				ID:             uuid.New().String(),
				Severity:       FindingSeverityMedium,
				Component:      ComponentTenant,
				Title:          "Geo Rules Disabled",
				Description:    fmt.Sprintf("Tenant '%s' has geo rules disabled", tenant.Name),
				Recommendation: "Enable geo rules for better traffic control",
				EntityID:       tenant.ID.String(),
				EntityType:     "tenant",
				Timestamp:      time.Now(),
			})
		}
	}
}

// ============================================
// AUTH AUDITS
// ============================================

func (s *SecurityAuditService) auditAuth(report *SecurityAuditReport) {
	report.ChecksRun = append(report.ChecksRun, "auth")

	// Check for failed login spikes
	var failedLogins int64
	s.db.Raw(`
		SELECT COUNT(*) FROM afftok_users 
		WHERE last_login_attempt < NOW() - INTERVAL '1 hour'
		AND login_failures > 5
	`).Scan(&failedLogins)

	if failedLogins > 10 {
		report.Findings = append(report.Findings, SecurityFinding{
			ID:             uuid.New().String(),
			Severity:       FindingSeverityHigh,
			Component:      ComponentAuth,
			Title:          "Multiple Failed Login Attempts",
			Description:    fmt.Sprintf("%d accounts have >5 failed logins in last hour", failedLogins),
			Recommendation: "Review for potential brute force attacks",
			Metadata: map[string]interface{}{
				"affected_accounts": failedLogins,
			},
			Timestamp: time.Now(),
		})
	}
}

func (s *SecurityAuditService) auditAuthQuick(report *SecurityAuditReport) {
	report.ChecksRun = append(report.ChecksRun, "auth_quick")

	// Check for active brute force
	threatDetector := GetThreatDetector()
	stats := threatDetector.GetThreatStats()

	if total, ok := stats["total_threats"].(int64); ok && total > 100 {
		report.Findings = append(report.Findings, SecurityFinding{
			ID:             uuid.New().String(),
			Severity:       FindingSeverityCritical,
			Component:      ComponentAuth,
			Title:          "High Threat Activity",
			Description:    fmt.Sprintf("%d threats detected recently", total),
			Recommendation: "Review threat logs and consider blocking suspicious IPs",
			Timestamp:      time.Now(),
		})
	}
}

// ============================================
// GEO RULES AUDITS
// ============================================

func (s *SecurityAuditService) auditGeoRules(report *SecurityAuditReport) {
	report.ChecksRun = append(report.ChecksRun, "geo_rules")

	// Check for overly permissive global rules
	var globalAllowAll []models.GeoRule
	s.db.Where("scope_type = ? AND mode = ? AND status = ?", "global", "allow", "active").
		Find(&globalAllowAll)

	for _, rule := range globalAllowAll {
		// Check if allowing all countries
		var countries []string
		if err := rule.Countries.Scan(&countries); err == nil {
			if len(countries) == 0 || len(countries) > 200 {
				report.Findings = append(report.Findings, SecurityFinding{
					ID:             uuid.New().String(),
					Severity:       FindingSeverityLow,
					Component:      ComponentGeoRules,
					Title:          "Overly Permissive Geo Rule",
					Description:    "Global geo rule allows all countries",
					Recommendation: "Consider restricting to target countries only",
					EntityID:       rule.ID.String(),
					EntityType:     "geo_rule",
					Timestamp:      time.Now(),
				})
			}
		}
	}
}

// ============================================
// RATE LIMIT AUDITS
// ============================================

func (s *SecurityAuditService) auditRateLimits(report *SecurityAuditReport) {
	report.ChecksRun = append(report.ChecksRun, "rate_limits")

	// Check if rate limiting is effectively configured
	// This would check actual rate limit configs
}

// ============================================
// HELPERS
// ============================================

func (s *SecurityAuditService) calculateSummary(report *SecurityAuditReport) *SecurityAuditSummary {
	summary := &SecurityAuditSummary{
		TotalChecks:   len(report.ChecksRun),
		TotalFindings: len(report.Findings),
		ByComponent:   make(map[string]int),
	}

	for _, finding := range report.Findings {
		switch finding.Severity {
		case FindingSeverityCritical:
			summary.CriticalFindings++
		case FindingSeverityHigh:
			summary.HighFindings++
		case FindingSeverityMedium:
			summary.MediumFindings++
		case FindingSeverityLow:
			summary.LowFindings++
		}

		summary.ByComponent[string(finding.Component)]++
	}

	return summary
}

func (s *SecurityAuditService) determineStatus(report *SecurityAuditReport) string {
	if report.Summary.CriticalFindings > 0 {
		return "failed"
	}
	if report.Summary.HighFindings > 0 {
		return "warning"
	}
	return "passed"
}

func (s *SecurityAuditService) storeReport(report *SecurityAuditReport) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.reports = append(s.reports, report)
	if len(s.reports) > s.maxReports {
		s.reports = s.reports[1:]
	}

	atomic.AddInt64(&s.totalFindings, int64(len(report.Findings)))
}

// ============================================
// GETTERS
// ============================================

// GetLatestReport returns the most recent report
func (s *SecurityAuditService) GetLatestReport() *SecurityAuditReport {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.reports) == 0 {
		return nil
	}
	return s.reports[len(s.reports)-1]
}

// GetAllFindings returns all findings from the latest report
func (s *SecurityAuditService) GetAllFindings() []SecurityFinding {
	report := s.GetLatestReport()
	if report == nil {
		return []SecurityFinding{}
	}
	return report.Findings
}

// GetFindingsBySeverity returns findings filtered by severity
func (s *SecurityAuditService) GetFindingsBySeverity(severity SecurityFindingSeverity) []SecurityFinding {
	findings := s.GetAllFindings()
	result := make([]SecurityFinding, 0)
	
	for _, f := range findings {
		if f.Severity == severity {
			result = append(result, f)
		}
	}
	return result
}

// GetStats returns service statistics
func (s *SecurityAuditService) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_audits":   atomic.LoadInt64(&s.totalAudits),
		"total_findings": atomic.LoadInt64(&s.totalFindings),
		"reports_stored": len(s.reports),
	}
}

