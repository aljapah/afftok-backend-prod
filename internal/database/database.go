package database

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/config"
	"github.com/aljapah/afftok-backend-prod/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// DBConfig holds database connection pool configuration
type DBConfig struct {
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultDBConfig returns optimized database configuration for high load
func DefaultDBConfig() DBConfig {
	numCPU := runtime.NumCPU()
	return DBConfig{
		MaxIdleConns:    numCPU * 5,          // Keep connections ready
		MaxOpenConns:    numCPU * 25,         // Max concurrent connections
		ConnMaxLifetime: 30 * time.Minute,    // Recycle connections
		ConnMaxIdleTime: 5 * time.Minute,     // Close idle connections
	}
}

func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.PostgresURL
	if dsn == "" {
		return nil, fmt.Errorf("POSTGRES_URL is not configured")
	}

	var gormLogger logger.Interface
	if cfg.IsProduction() {
		gormLogger = logger.Default.LogMode(logger.Error)
	} else {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	// Disable prepared statements to avoid "cached plan must not change result type" error
	// This happens when schema changes and PostgreSQL has cached query plans
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		// Performance optimizations - PrepareStmt disabled to prevent cached plan errors
		PrepareStmt:                              false,
		SkipDefaultTransaction:                   true,  // Skip default transaction for single queries
		DisableForeignKeyConstraintWhenMigrating: false, // Keep FK constraints
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Apply high-performance configuration
	dbConfig := DefaultDBConfig()
	sqlDB.SetMaxIdleConns(dbConfig.MaxIdleConns)
	sqlDB.SetMaxOpenConns(dbConfig.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(dbConfig.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(dbConfig.ConnMaxIdleTime)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db
	log.Printf("✅ PostgreSQL connected successfully (max_open=%d, max_idle=%d)",
		dbConfig.MaxOpenConns, dbConfig.MaxIdleConns)
	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database connection not established")
	}

	// Enable UUID extension
	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")

	// Run migrations
	err := db.AutoMigrate(
		&models.AdminUser{},
		&models.AfftokUser{},
		&models.Network{},
		&models.Offer{},
		&models.UserOffer{},
		&models.Click{},
		&models.Conversion{},
		&models.Team{},
		&models.TeamMember{},
		&models.Badge{},
		&models.UserBadge{},
		&models.TrackingEvent{},
		// Phase 8.2: API Keys
		&models.AdvertiserAPIKey{},
		&models.APIKeyUsageLog{},
		// Phase 8.3: Geo Rules
		&models.GeoRule{},
		// Phase 8.5: Webhooks
		&models.WebhookPipeline{},
		&models.WebhookStep{},
		&models.WebhookExecution{},
		&models.WebhookStepResult{},
		&models.WebhookDLQItem{},
		// Phase 8.6: Multi-Tenant
		&models.Tenant{},
		&models.TenantDomain{},
		&models.TenantAuditLog{},
		// Contests/Challenges
		&models.Contest{},
		&models.ContestParticipant{},
		// Invoices
		&models.Invoice{},
		&models.InvoiceItem{},
	)

	if err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	// Create additional indexes for performance
	createIndexes(db)
	
	// Generate unique codes for existing users who don't have one
	generateMissingUniqueCodes(db)

	log.Println("✅ Database migration completed successfully")
	return nil
}

// generateMissingUniqueCodes generates unique_code for users who don't have one
func generateMissingUniqueCodes(db *gorm.DB) {
	var users []models.AfftokUser
	db.Where("unique_code IS NULL OR unique_code = ''").Find(&users)
	
	for _, user := range users {
		user.UniqueCode = models.GenerateUniqueCode()
		if err := db.Model(&user).Update("unique_code", user.UniqueCode).Error; err != nil {
			log.Printf("⚠️ Failed to generate unique code for user %s: %v", user.Username, err)
		} else {
			log.Printf("✅ Generated unique code %s for user %s", user.UniqueCode, user.Username)
		}
	}
	
	if len(users) > 0 {
		log.Printf("✅ Generated unique codes for %d existing users", len(users))
	}
}

// createIndexes creates additional indexes for tracking performance
func createIndexes(db *gorm.DB) {
	// High-performance indexes for extreme load
	indexes := []string{
		// ============================================
		// CLICKS TABLE - Critical for high throughput
		// ============================================
		
		// Primary lookup: user_offer + time for range queries
		"CREATE INDEX IF NOT EXISTS idx_clicks_user_offer_time ON clicks(user_offer_id, clicked_at DESC)",
		
		// Fingerprint for deduplication (partial index for recent)
		"CREATE INDEX IF NOT EXISTS idx_clicks_fingerprint ON clicks(fingerprint) WHERE clicked_at > NOW() - INTERVAL '1 hour'",
		
		// IP for fraud detection
		"CREATE INDEX IF NOT EXISTS idx_clicks_ip ON clicks(ip)",
		
		// Country for analytics
		"CREATE INDEX IF NOT EXISTS idx_clicks_country ON clicks(country) WHERE country IS NOT NULL",
		
		// Device for analytics
		"CREATE INDEX IF NOT EXISTS idx_clicks_device ON clicks(device) WHERE device IS NOT NULL",
		
		// Date partitioning helper (for future partitioning)
		"CREATE INDEX IF NOT EXISTS idx_clicks_date ON clicks(DATE(clicked_at))",
		
		// ============================================
		// CONVERSIONS TABLE
		// ============================================
		
		// Status + time for filtering
		"CREATE INDEX IF NOT EXISTS idx_conversions_status_time ON conversions(status, converted_at DESC)",
		
		// External ID for deduplication (unique)
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_conversions_external_id ON conversions(external_conversion_id) WHERE external_conversion_id IS NOT NULL AND external_conversion_id != ''",
		
		// User offer for lookups
		"CREATE INDEX IF NOT EXISTS idx_conversions_user_offer ON conversions(user_offer_id)",
		
		// Network for filtering
		"CREATE INDEX IF NOT EXISTS idx_conversions_network ON conversions(network_id) WHERE network_id IS NOT NULL",
		
		// ============================================
		// USER OFFERS TABLE
		// ============================================
		
		// User + status for active offers
		"CREATE INDEX IF NOT EXISTS idx_user_offers_user_status ON user_offers(user_id, status)",
		
		// Short link for tracking resolution (critical path)
		"CREATE INDEX IF NOT EXISTS idx_user_offers_short_link ON user_offers(short_link) WHERE short_link IS NOT NULL AND short_link != ''",
		
		// Offer ID for stats aggregation
		"CREATE INDEX IF NOT EXISTS idx_user_offers_offer ON user_offers(offer_id)",
		
		// Stats lookup (covering index)
		"CREATE INDEX IF NOT EXISTS idx_user_offers_stats ON user_offers(user_id, total_clicks, total_conversions)",
		
		// ============================================
		// OFFERS TABLE
		// ============================================
		
		// Status + category for filtering
		"CREATE INDEX IF NOT EXISTS idx_offers_status_category ON offers(status, category)",
		
		// Network for filtering
		"CREATE INDEX IF NOT EXISTS idx_offers_network ON offers(network_id)",
		
		// ============================================
		// TRACKING EVENTS TABLE
		// ============================================
		
		// Type + time for analytics
		"CREATE INDEX IF NOT EXISTS idx_events_type_time ON tracking_events(event_type, created_at DESC)",
		
		// User for history
		"CREATE INDEX IF NOT EXISTS idx_events_user ON tracking_events(user_id) WHERE user_id IS NOT NULL",
		
		// ============================================
		// USERS TABLE
		// ============================================
		
		// Email for login (should be unique)
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON afftok_users(email)",
		
		// Username for lookup
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON afftok_users(username)",
		
		// Referral code
		"CREATE INDEX IF NOT EXISTS idx_users_referral ON afftok_users(referral_code) WHERE referral_code IS NOT NULL",
	}

	log.Println("📊 Creating performance indexes...")
	successCount := 0
	for _, idx := range indexes {
		if err := db.Exec(idx).Error; err != nil {
			log.Printf("⚠️ Index creation warning: %v", err)
		} else {
			successCount++
		}
	}
	log.Printf("📊 Created %d/%d indexes successfully", successCount, len(indexes))
}

func Close(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

func GetDB() *gorm.DB {
	return DB
}
