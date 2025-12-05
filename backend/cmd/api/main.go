package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/aljapah/afftok-backend-prod/internal/alerting"
	"github.com/aljapah/afftok-backend-prod/internal/cache"
	"github.com/aljapah/afftok-backend-prod/internal/config"
	"github.com/aljapah/afftok-backend-prod/internal/database"
	"github.com/aljapah/afftok-backend-prod/internal/handlers"
	"github.com/aljapah/afftok-backend-prod/internal/middleware"
	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/aljapah/afftok-backend-prod/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Optimize Go runtime for high throughput
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.Printf("🚀 Starting AffTok API with %d CPUs", runtime.NumCPU())

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}
	defer database.Close(db)

	redisClient, err := cache.ConnectRedis(cfg)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer cache.CloseRedis(redisClient)

	// Skip AutoMigrate - tables already exist in production
	// This prevents "insufficient arguments" error from schema conflicts
	if os.Getenv("SKIP_MIGRATION") != "true" {
		if err := database.AutoMigrate(db); err != nil {
			log.Printf("⚠️ Migration warning (non-fatal): %v", err)
			// Don't fail - tables likely already exist
		}
	} else {
		log.Println("⏭️ Skipping database migration (SKIP_MIGRATION=true)")
	}

	// Graceful shutdown handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("🛑 Shutting down gracefully...")
		services.StopAllPools()
		os.Exit(0)
	}()

	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Security middlewares
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.SecurityHeadersMiddleware())
	router.Use(middleware.SecureErrorMiddleware())
	router.Use(middleware.AuditLogMiddleware())

	router.Static("/public", "./public")

	authHandler := handlers.NewAuthHandler(db)
	userHandler := handlers.NewUserHandler(db)
	offerHandler := handlers.NewOfferHandler(db)
	networkHandler := handlers.NewNetworkHandler(db)
	postbackHandler := handlers.NewPostbackHandler(db)
	teamHandler := handlers.NewTeamHandler(db)
	badgeHandler := handlers.NewBadgeHandler(db)
	clickHandler := handlers.NewClickHandler(db)
	contestHandler := handlers.NewContestHandler(db)
	promoterHandler := handlers.NewPromoterHandler(db)
	advertiserHandler := handlers.NewAdvertiserHandler(db)
	observabilityHandler := handlers.NewObservabilityHandler()
	
	// Phase 7: Admin Observability Handlers
	adminDashboardHandler := handlers.NewAdminDashboardHandler()
	adminMetricsHandler := handlers.NewAdminMetricsHandler()
	adminHealthHandler := handlers.NewAdminHealthHandler()
	adminLogsHandler := handlers.NewAdminLogsHandler()
	adminFraudHandler := handlers.NewAdminFraudHandler()
	adminDiagnosticsHandler := handlers.NewAdminDiagnosticsHandler()
	adminStressHandler := handlers.NewAdminStressHandler()

	// Phase 8.1: Database Hardening Handlers
	dbStatsService := services.NewDBStatsService(db)
	partitionService := services.NewPartitionService(db)
	adminDBHandler := handlers.NewAdminDBHandler(dbStatsService, partitionService)
	
	// Initialize DB Router for Read Replicas
	if _, err := database.InitDBRouter(cfg); err != nil {
		log.Printf("⚠️ DB Router init warning: %v", err)
	}

	// Phase 8.2: API Key System
	apiKeyService := services.NewAPIKeyService(db)
	observabilityService := services.NewObservabilityService()
	adminAPIKeysHandler := handlers.NewAdminAPIKeysHandler(apiKeyService)
	
	// Initialize API Key Middleware with services
	middleware.InitAPIKeyMiddleware(apiKeyService, observabilityService)
	
	// Set API Key service on postback handler
	postbackHandler.SetAPIKeyService(apiKeyService)

	// Phase 8.3: Geo Rules System
	geoRuleService := services.NewGeoRuleService(db)
	adminGeoRulesHandler := handlers.NewAdminGeoRulesHandler(geoRuleService)
	
	// Set Geo Rule service on handlers
	clickHandler.SetGeoRuleService(geoRuleService)
	postbackHandler.SetGeoRuleService(geoRuleService)

	// Phase 8.4: Link Signing System
	linkSigningService := services.NewLinkSigningService()
	adminLinkSigningHandler := handlers.NewAdminLinkSigningHandler(linkSigningService)
	linkService := services.NewLinkService()
	
	// Set Link Signing service on handlers
	clickHandler.SetLinkSigningService(linkSigningService)
	offerHandler.SetLinkSigningService(linkSigningService)

	// Phase 8.5: Advanced Webhooks Engine
	webhookService := services.GetWebhookService(db)
	adminWebhooksHandler := handlers.NewAdminWebhooksHandler(db)
	
	// Start webhook workers
	webhookService.Start()
	defer webhookService.Stop()

	// Phase 8.6: Multi-Tenant System
	middleware.InitTenantMiddleware(db)
	adminTenantsHandler := handlers.NewAdminTenantsHandler(db)
	
	// Create default tenant if not exists
	tenantService := services.NewTenantService(db)
	if _, err := tenantService.GetTenant(models.DefaultTenantID); err != nil {
		defaultTenant := models.DefaultTenant()
		tenantService.CreateTenant(defaultTenant)
		log.Println("✅ Default tenant created")
	}

	// Phase 8.7: Edge CDN Layer
	edgeIngestHandler := handlers.NewEdgeIngestHandler(db)
	edgeIngestHandler.SetLinkService(linkService)
	adminEdgeHandler := handlers.NewAdminEdgeHandler(db)
	
	// Start edge ingest workers
	edgeIngestService := services.GetEdgeIngestService(db)
	edgeIngestService.SetLinkService(linkService)
	edgeIngestService.Start(4) // 4 workers
	defer edgeIngestService.Stop()
	log.Println("✅ Edge ingest workers started")

	// Phase 8.8: Zero-Drop Tracking Mode
	adminZeroDropHandler := handlers.NewAdminZeroDropHandler(db)
	
	// Initialize Zero-Drop system (WAL, Failover Queue, Crash Recovery, Streams)
	if err := services.InitZeroDrop(db); err != nil {
		log.Printf("⚠️ Zero-Drop init warning: %v", err)
	} else {
		log.Println("✅ Zero-Drop system initialized")
	}
	defer services.ShutdownZeroDrop()
	
	// Start Postback Queue service
	postbackQueueService := services.GetPostbackQueueService()
	postbackQueueService.Start()
	defer postbackQueueService.Stop()
	log.Println("✅ Postback queue service started")

	// Phase 8.9: Launch Mode - Production Hardening
	adminLaunchHandler := handlers.NewAdminLaunchHandler(db)
	
	// Initialize Alert Manager
	alertManager := alerting.GetAlertManager()
	_ = alertManager // Used in handlers
	log.Println("✅ Alert manager initialized")
	
	// Initialize Threat Detector
	threatDetector := services.GetThreatDetector()
	_ = threatDetector // Used in handlers
	log.Println("✅ Threat detector initialized")
	
	// Initialize Logging Mode Service
	loggingModeService := services.GetLoggingModeService()
	_ = loggingModeService // Used in handlers
	log.Println("✅ Logging mode service initialized")

	// Initialize worker pools for async processing
	services.StartAllPools()

	// Public health endpoint
	router.GET("/health", observabilityHandler.GetHealth)

	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		auth.Use(middleware.AuthRateLimitMiddleware())
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)
		}

		// Advertiser Registration (public - no auth required)
		api.POST("/advertiser/register", advertiserHandler.RegisterAdvertiser)

		// Click tracking with bot detection and rate limiting
		api.GET("/c/:id", middleware.BotDetectionMiddleware(), clickHandler.TrackClick)
		api.GET("/promoter/:id", promoterHandler.GetPromoterPage)
		api.GET("/promoter/user/:username", promoterHandler.GetPromoterPageByUsername) // Public - landing page
		api.POST("/rate-promoter", promoterHandler.RatePromoter)

		// Postback with security validation + API Key or JWT auth
		api.POST("/postback", middleware.PostbackSecurityMiddleware(), middleware.APIKeyOrJWTMiddleware(), postbackHandler.HandlePostback)

		api.GET("/offers", offerHandler.GetAllOffers)
		api.GET("/offers/:id", offerHandler.GetOffer)

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.GET("/auth/me", authHandler.GetMe)
			protected.PUT("/profile", userHandler.UpdateProfile)

			protected.GET("/users", userHandler.GetAllUsers)
			protected.GET("/users/:id", userHandler.GetUser)
			
			// Stats endpoints
			protected.GET("/stats/me", userHandler.GetMyStats)
			protected.GET("/stats/daily", userHandler.GetDailyStats)

			// Leaderboard
			protected.GET("/leaderboard", userHandler.GetLeaderboard)

			protected.POST("/offers/:id/join", offerHandler.JoinOffer)
			protected.GET("/offers/my", offerHandler.GetMyOffers)

			networks := protected.Group("/networks")
			{
				networks.GET("", networkHandler.GetAllNetworks)
				networks.GET("/:id", networkHandler.GetNetwork)
			}

			teams := protected.Group("/teams")
			{
				teams.GET("", teamHandler.GetAllTeams)
				teams.GET("/my", teamHandler.GetMyTeam)
				teams.GET("/:id", teamHandler.GetTeam)
				teams.POST("", teamHandler.CreateTeam)
				teams.POST("/:id/join", teamHandler.JoinTeam)
				teams.POST("/:id/leave", teamHandler.LeaveTeam)
				teams.POST("/join/:code", teamHandler.JoinTeamByInviteCode)
				teams.POST("/:id/approve/:memberId", teamHandler.ApproveMember)
				teams.POST("/:id/reject/:memberId", teamHandler.RejectMember)
				teams.DELETE("/:id/members/:memberId", teamHandler.RemoveMember)
				teams.GET("/:id/pending", teamHandler.GetPendingRequests)
				teams.POST("/:id/regenerate-invite", teamHandler.RegenerateInviteCode)
				teams.DELETE("/:id", teamHandler.DeleteTeam)
			}

			badges := protected.Group("/badges")
			{
				badges.GET("", badgeHandler.GetAllBadges)
				badges.GET("/my", badgeHandler.GetMyBadges)
			}

			// Contests / Challenges
			contests := protected.Group("/contests")
			{
				contests.GET("", contestHandler.GetActiveContests)
				contests.GET("/my", contestHandler.GetMyContests)
				contests.GET("/:id", contestHandler.GetContest)
				contests.GET("/:id/leaderboard", contestHandler.GetContestLeaderboard)
				contests.POST("/:id/join", contestHandler.JoinContest)
			}

			clicks := protected.Group("/clicks")
			{
				clicks.GET("/my", clickHandler.GetMyClicks)
				clicks.GET("/:id/stats", clickHandler.GetClickStats)
				clicks.GET("/by-offer", clickHandler.GetClicksByOffer)
			}

			// ========== Advertiser Routes ==========
			advertiser := protected.Group("/advertiser")
			{
				advertiser.GET("/dashboard", advertiserHandler.GetDashboard)
				advertiser.GET("/offers", advertiserHandler.GetMyOffers)
				advertiser.POST("/offers", advertiserHandler.CreateOffer)
				advertiser.PUT("/offers/:id", advertiserHandler.UpdateOffer)
				advertiser.DELETE("/offers/:id", advertiserHandler.DeleteOffer)
				advertiser.POST("/offers/:id/pause", advertiserHandler.PauseOffer)
				advertiser.GET("/offers/:id/stats", advertiserHandler.GetOfferStats)
			}

			admin := protected.Group("/admin")
			admin.Use(middleware.AdminMiddleware())
			{
				admin.PUT("/users/:id", userHandler.UpdateUser)
				admin.DELETE("/users/:id", userHandler.DeleteUser)

				admin.POST("/offers", offerHandler.CreateOffer)
				admin.PUT("/offers/:id", offerHandler.UpdateOffer)
				admin.DELETE("/offers/:id", offerHandler.DeleteOffer)

				admin.POST("/networks", networkHandler.CreateNetwork)
				admin.PUT("/networks/:id", networkHandler.UpdateNetwork)
				admin.DELETE("/networks/:id", networkHandler.DeleteNetwork)

				admin.GET("/conversions", postbackHandler.GetConversions)
				admin.POST("/conversions/:id/approve", postbackHandler.ApproveConversion)
				admin.POST("/conversions/:id/reject", postbackHandler.RejectConversion)

				// Pending Offers Management (for advertiser submissions)
				admin.GET("/offers/pending", advertiserHandler.GetPendingOffers)
				admin.POST("/offers/:id/approve", advertiserHandler.ApproveOffer)
				admin.POST("/offers/:id/reject", advertiserHandler.RejectOffer)

				// Contests / Challenges Management
				admin.GET("/contests", contestHandler.AdminGetAllContests)
				admin.POST("/contests", contestHandler.AdminCreateContest)
				admin.GET("/contests/:id", contestHandler.GetContest)
				admin.PUT("/contests/:id", contestHandler.AdminUpdateContest)
				admin.DELETE("/contests/:id", contestHandler.AdminDeleteContest)
				admin.GET("/contests/:id/participants", contestHandler.AdminGetContestParticipants)

				admin.POST("/badges", badgeHandler.CreateBadge)
			admin.PUT("/badges/:id", badgeHandler.UpdateBadge)
			admin.DELETE("/badges/:id", badgeHandler.DeleteBadge)

			// ============================================
			// PHASE 7: SYSTEM OBSERVABILITY API LAYER
			// ============================================

			// 1. System Dashboard endpoint
			admin.GET("/dashboard", adminDashboardHandler.GetDashboard)

			// 2. Metrics endpoints
			admin.GET("/metrics", adminMetricsHandler.GetMetrics)

			// 3. Metrics Export endpoint
			admin.GET("/metrics/export", adminMetricsHandler.ExportMetrics)

			// 4. Health endpoints
			admin.GET("/health", adminHealthHandler.GetHealth)
			admin.GET("/connections", adminHealthHandler.GetConnections)

			// 5. Logs endpoints
			admin.GET("/logs/recent", adminLogsHandler.GetRecentLogs)
			admin.GET("/logs/errors", adminLogsHandler.GetErrorLogs)
			admin.GET("/logs/fraud", adminLogsHandler.GetFraudLogs)
			admin.GET("/logs/categories", adminLogsHandler.GetLogCategories)
			admin.GET("/logs/category/:category", adminLogsHandler.GetLogsByCategory)
			admin.GET("/logs/ip/:ip", adminLogsHandler.GetLogsByIP)
			admin.GET("/logs/user/:user_id", adminLogsHandler.GetLogsByUser)

			// 6. Fraud insights endpoint
			admin.GET("/fraud/insights", adminFraudHandler.GetFraudInsights)
			admin.POST("/fraud/block-ip", adminFraudHandler.BlockIP)
			admin.POST("/fraud/unblock-ip", adminFraudHandler.UnblockIP)
			admin.GET("/fraud/blocked-ips", adminFraudHandler.GetBlockedIPs)

			// 7. Diagnostics endpoints
			admin.GET("/diagnostics/redis", adminDiagnosticsHandler.GetRedisDiagnostics)
			admin.GET("/diagnostics/db", adminDiagnosticsHandler.GetDBDiagnostics)
			admin.GET("/diagnostics/system", adminDiagnosticsHandler.GetSystemDiagnostics)

			// 8. Stress test endpoints
			admin.GET("/stress/clicks", adminStressHandler.SimulateClicks)
			admin.GET("/stress/postbacks", adminStressHandler.SimulatePostbacks)
			admin.GET("/stress/full", adminStressHandler.RunFullStressTest)
			admin.GET("/stress/pools", adminStressHandler.GetWorkerPoolStats)
			admin.GET("/stress/cache", adminStressHandler.GetCacheStats)

			// ============================================
			// PHASE 8.1: DATABASE HARDENING API LAYER
			// ============================================

			// 1. Backup/PITR Info
			admin.GET("/db/backup-info", adminDBHandler.GetBackupInfo)

			// 2. Vacuum/Analyze Plan
			admin.GET("/db/vacuum-plan", adminDBHandler.GetVacuumPlan)
			admin.GET("/db/stats", adminDBHandler.GetTableStats)

			// 3. Index Profiling
			admin.GET("/db/indexes", adminDBHandler.GetIndexes)

			// 4. Partitioning
			admin.GET("/db/partitions", adminDBHandler.GetPartitionStatus)
			admin.POST("/db/partition/create", adminDBHandler.CreatePartition)
			admin.POST("/db/partitions/ensure", adminDBHandler.EnsurePartitions)
			admin.GET("/db/partition/migration-plan", adminDBHandler.GetMigrationPlan)

			// 5. Connection Pool
			admin.GET("/db/pool", adminDBHandler.GetConnectionPool)

			// 6. Latency & Performance
			admin.GET("/db/latency", adminDBHandler.GetDBLatency)
			admin.GET("/db/slow-queries", adminDBHandler.GetSlowQueries)

			// 7. Size
			admin.GET("/db/size", adminDBHandler.GetDBSize)

			// 8. Full Report
			admin.GET("/db/report", adminDBHandler.GetDBReport)

			// ============================================
			// PHASE 8.2: ADVERTISER API KEYS
			// ============================================

			// 1. List all API keys
			admin.GET("/api-keys", adminAPIKeysHandler.GetAllAPIKeys)

			// 2. Get single API key (masked)
			admin.GET("/api-keys/:id", adminAPIKeysHandler.GetAPIKeyByID)

			// 3. Get API keys by advertiser
			admin.GET("/advertisers/:id/api-keys", adminAPIKeysHandler.GetAPIKeysByAdvertiser)

			// 4. Create API key for advertiser
			admin.POST("/advertisers/:id/api-keys", adminAPIKeysHandler.CreateAPIKey)

			// 5. Rotate API key
			admin.POST("/api-keys/:id/rotate", adminAPIKeysHandler.RotateAPIKey)

			// 6. Revoke API key
			admin.POST("/api-keys/:id/revoke", adminAPIKeysHandler.RevokeAPIKey)

			// 7. IP management
			admin.POST("/api-keys/:id/allow-ip", adminAPIKeysHandler.AddAllowedIP)
			admin.POST("/api-keys/:id/deny-ip", adminAPIKeysHandler.RemoveAllowedIP)

			// 8. API Key stats report
			admin.GET("/security/api-keys/report", adminAPIKeysHandler.GetAPIKeyStats)

			// ============================================
			// PHASE 8.3: GEO RULES
			// ============================================

			// 1. List all geo rules
			admin.GET("/geo-rules", adminGeoRulesHandler.GetAllGeoRules)

			// 2. Get single geo rule
			admin.GET("/geo-rules/:id", adminGeoRulesHandler.GetGeoRuleByID)

			// 3. Get geo rules by offer
			admin.GET("/offers/:id/geo-rules", adminGeoRulesHandler.GetGeoRulesByOffer)

			// 4. Get geo rules by advertiser (reuse existing route pattern)
			admin.GET("/advertisers/:id/geo-rules", adminGeoRulesHandler.GetGeoRulesByAdvertiser)

			// 5. Create geo rule
			admin.POST("/geo-rules", adminGeoRulesHandler.CreateGeoRule)

			// 6. Update geo rule
			admin.PUT("/geo-rules/:id", adminGeoRulesHandler.UpdateGeoRule)

			// 7. Delete geo rule
			admin.DELETE("/geo-rules/:id", adminGeoRulesHandler.DeleteGeoRule)

			// 8. Geo rule statistics
			admin.GET("/geo-rules/stats", adminGeoRulesHandler.GetGeoRuleStats)

			// 9. Country codes reference
			admin.GET("/geo-rules/countries", adminGeoRulesHandler.GetCountryCodes)

			// 10. Test geo rule
			admin.POST("/geo-rules/test", adminGeoRulesHandler.TestGeoRule)

			// ============================================
			// PHASE 8.4: LINK SIGNING & TTL VALIDATION
			// ============================================

			// 1. Link signing configuration
			admin.GET("/link-signing/config", adminLinkSigningHandler.GetConfig)
			admin.PUT("/link-signing/config", adminLinkSigningHandler.UpdateConfig)

			// 2. Test link validation
			admin.GET("/link-signing/test", adminLinkSigningHandler.TestLink)

			// 3. Generate signed link (for testing)
			admin.POST("/link-signing/generate", adminLinkSigningHandler.GenerateSignedLink)

			// 4. Secret rotation
			admin.POST("/link-signing/rotate-secret", adminLinkSigningHandler.RotateSecret)

			// 5. Replay cache management
			admin.POST("/link-signing/replay/clear", adminLinkSigningHandler.ClearReplayCache)
			admin.GET("/link-signing/replay/stats", adminLinkSigningHandler.GetReplayCacheStats)

			// 6. Link signing statistics
			admin.GET("/security/link-signing/stats", adminLinkSigningHandler.GetStats)

			// ============================================
			// PHASE 8.5: ADVANCED WEBHOOKS ENGINE
			// ============================================

			// 1. Webhook Pipelines
			admin.GET("/webhooks/pipelines", adminWebhooksHandler.GetAllPipelines)
			admin.GET("/webhooks/pipelines/:id", adminWebhooksHandler.GetPipeline)
			admin.POST("/webhooks/pipelines", adminWebhooksHandler.CreatePipeline)
			admin.PUT("/webhooks/pipelines/:id", adminWebhooksHandler.UpdatePipeline)
			admin.DELETE("/webhooks/pipelines/:id", adminWebhooksHandler.DeletePipeline)

			// 2. Execution Logs
			admin.GET("/webhooks/logs/recent", adminWebhooksHandler.GetRecentLogs)
			admin.GET("/webhooks/logs/:task_id", adminWebhooksHandler.GetExecutionLog)
			admin.GET("/webhooks/logs/failures", adminWebhooksHandler.GetFailureLogs)

			// 3. Dead Letter Queue (DLQ)
			admin.GET("/webhooks/dlq", adminWebhooksHandler.GetDLQ)
			admin.POST("/webhooks/dlq/retry/:id", adminWebhooksHandler.RetryDLQItem)
			admin.DELETE("/webhooks/dlq/:id", adminWebhooksHandler.DeleteDLQItem)

			// 4. Testing
			admin.POST("/webhooks/test/pipeline", adminWebhooksHandler.TestPipeline)
			admin.POST("/webhooks/test/step", adminWebhooksHandler.TestStep)

			// 5. Stats & Reference
			admin.GET("/webhooks/stats", adminWebhooksHandler.GetStats)
			admin.GET("/webhooks/trigger-types", adminWebhooksHandler.GetTriggerTypes)
			admin.GET("/webhooks/signature-modes", adminWebhooksHandler.GetSignatureModes)

			// ============================================
			// PHASE 8.6: MULTI-TENANT SYSTEM
			// ============================================

			// 1. Tenant CRUD
			admin.GET("/tenants", adminTenantsHandler.GetAllTenants)
			admin.GET("/tenants/report", adminTenantsHandler.GetTenantsReport)
			admin.GET("/tenants/plans", adminTenantsHandler.GetPlans)
			admin.GET("/tenants/:id", adminTenantsHandler.GetTenant)
			admin.POST("/tenants", adminTenantsHandler.CreateTenant)
			admin.PUT("/tenants/:id", adminTenantsHandler.UpdateTenant)
			admin.DELETE("/tenants/:id", adminTenantsHandler.DeleteTenant)

			// 2. Tenant Status
			admin.POST("/tenants/:id/suspend", adminTenantsHandler.SuspendTenant)
			admin.POST("/tenants/:id/activate", adminTenantsHandler.ActivateTenant)

			// 3. Tenant Stats
			admin.GET("/tenants/:id/stats", adminTenantsHandler.GetTenantStats)

			// 4. Tenant Domains
			admin.GET("/tenants/:id/domains", adminTenantsHandler.GetTenantDomains)
			admin.POST("/tenants/:id/domains", adminTenantsHandler.AddTenantDomain)
			admin.DELETE("/tenants/:id/domains/:domain", adminTenantsHandler.RemoveTenantDomain)

			// 5. Tenant Branding
			admin.GET("/tenants/:id/branding", adminTenantsHandler.GetTenantBranding)
			admin.PUT("/tenants/:id/branding", adminTenantsHandler.UpdateTenantBranding)

			// 6. Tenant Plan
			admin.POST("/tenants/:id/plan", adminTenantsHandler.ChangeTenantPlan)

			// 7. Tenant Settings
			admin.GET("/tenants/:id/settings", adminTenantsHandler.GetTenantSettings)
			admin.PUT("/tenants/:id/settings", adminTenantsHandler.UpdateTenantSettings)

			// 8. Tenant Features
			admin.GET("/tenants/:id/features", adminTenantsHandler.GetTenantFeatures)

			// 9. Tenant Audit Logs
			admin.GET("/tenants/:id/audit-logs", adminTenantsHandler.GetTenantAuditLogs)

			// ============================================
			// PHASE 8.7: EDGE CDN LAYER
			// ============================================

			// 1. Edge Status
			admin.GET("/edge/status", adminEdgeHandler.GetEdgeStatus)
			admin.GET("/edge/regions", adminEdgeHandler.GetEdgeRegions)
			admin.GET("/edge/router", adminEdgeHandler.GetEdgeRouter)
			admin.GET("/edge/stats", adminEdgeHandler.GetEdgeFullStats)

			// 2. Edge Queue
			admin.GET("/edge/queue", adminEdgeHandler.GetEdgeQueue)
			admin.POST("/edge/queue/flush", adminEdgeHandler.FlushEdgeQueue)

			// 3. Edge Failover
			admin.GET("/edge/failover", adminEdgeHandler.GetEdgeFailover)

			// 4. Edge Cache
			admin.POST("/edge/cache/refresh", adminEdgeHandler.RefreshEdgeCache)

			// ============================================
			// PHASE 8.8: ZERO-DROP TRACKING MODE
			// ============================================

			// 1. Zero-Drop Status
			admin.GET("/zero-drop/status", adminZeroDropHandler.GetStatus)
			admin.GET("/zero-drop/metrics", adminZeroDropHandler.GetMetrics)

			// 2. WAL Management
			admin.GET("/zero-drop/wal", adminZeroDropHandler.GetWALStatus)
			admin.GET("/zero-drop/wal/pending", adminZeroDropHandler.GetWALPending)
			admin.POST("/zero-drop/wal/compact", adminZeroDropHandler.CompactWAL)

			// 3. Replay & Recovery
			admin.POST("/zero-drop/replay", adminZeroDropHandler.TriggerReplay)
			admin.POST("/zero-drop/fix-inconsistencies", adminZeroDropHandler.FixInconsistencies)

			// 4. Redis Streams
			admin.GET("/zero-drop/streams", adminZeroDropHandler.GetStreamsStatus)

			// 5. Failover Queue
			admin.GET("/zero-drop/failover-queue", adminZeroDropHandler.GetFailoverQueueStatus)
			admin.POST("/zero-drop/failover-queue/flush", adminZeroDropHandler.FlushFailoverQueue)

			// 6. Zero-Drop Mode Control
			admin.POST("/zero-drop/enable", adminZeroDropHandler.EnableZeroDropMode)
			admin.POST("/zero-drop/disable", adminZeroDropHandler.DisableZeroDropMode)
			admin.POST("/zero-drop/tenant/:id/enable", adminZeroDropHandler.EnableZeroDropForTenant)
			admin.POST("/zero-drop/tenant/:id/disable", adminZeroDropHandler.DisableZeroDropForTenant)

			// 7. Postback Queue
			admin.GET("/postbacks/queue", adminZeroDropHandler.GetPostbackQueue)
			admin.GET("/postbacks/dlq", adminZeroDropHandler.GetPostbackDLQ)
			admin.POST("/postbacks/dlq/:id/retry", adminZeroDropHandler.RetryPostbackDLQItem)
			admin.POST("/postbacks/dlq/retry-all", adminZeroDropHandler.RetryAllPostbackDLQ)
			admin.DELETE("/postbacks/dlq/:id", adminZeroDropHandler.DeletePostbackDLQItem)

			// ============================================
			// PHASE 8.9: LAUNCH MODE (PRODUCTION HARDENING)
			// ============================================

			// 1. Launch Dashboard
			admin.GET("/launch-dashboard", adminLaunchHandler.GetLaunchDashboard)
			admin.GET("/live-metrics", adminLaunchHandler.GetLiveMetrics)

			// 2. Logging Mode
			admin.GET("/logging/mode", adminLaunchHandler.GetLoggingMode)
			admin.POST("/logging/mode", adminLaunchHandler.SetLoggingMode)
			admin.GET("/logging/state", adminLaunchHandler.GetLoggingState)

			// 3. Threat Protection
			admin.GET("/security/threats", adminLaunchHandler.GetThreats)
			admin.GET("/security/anomalies", adminLaunchHandler.GetAnomalies)
			admin.GET("/security/ip-blocks", adminLaunchHandler.GetIPBlocks)
			admin.POST("/security/ip-blocks", adminLaunchHandler.BlockIPAddress)
			admin.DELETE("/security/ip-blocks/:ip", adminLaunchHandler.UnblockIPAddress)

			// 4. Alerts
			admin.GET("/alerts/active", adminLaunchHandler.GetActiveAlerts)
			admin.GET("/alerts/history", adminLaunchHandler.GetAlertHistory)
			admin.POST("/alerts/:id/acknowledge", adminLaunchHandler.AcknowledgeAlert)
			admin.GET("/alerts/thresholds", adminLaunchHandler.GetAlertThresholds)
			admin.PUT("/alerts/thresholds", adminLaunchHandler.UpdateAlertThresholds)

			// 5. Load Testing
			admin.POST("/loadtest/run", adminLaunchHandler.RunLoadTest)
			admin.GET("/loadtest/report", adminLaunchHandler.GetLoadTestReport)

			// ============================================
			// PHASE 9: QA, SECURITY AUDIT & BENCHMARKS
			// ============================================
			
			// Initialize QA Handler
			adminQAHandler := handlers.NewAdminQAHandler(db)

			// 1. E2E Tests
			admin.POST("/e2e-tests/run", adminQAHandler.RunE2ETest)
			admin.GET("/e2e-tests/scenarios", adminQAHandler.GetE2EScenarios)
			admin.GET("/e2e-tests/history", adminQAHandler.GetE2ETestHistory)
			admin.GET("/e2e-tests/:id", adminQAHandler.GetE2ETestRun)

			// 2. Consistency Checks
			admin.GET("/consistency/run", adminQAHandler.RunConsistencyCheck)
			admin.GET("/consistency/report", adminQAHandler.GetConsistencyReport)
			admin.GET("/consistency/issues", adminQAHandler.GetConsistencyIssues)
			admin.POST("/consistency/fix", adminQAHandler.FixConsistencyIssues)

			// 3. Security Audit
			admin.GET("/security/audit/run", adminQAHandler.RunSecurityAudit)
			admin.GET("/security/audit/report", adminQAHandler.GetSecurityAuditReport)
			admin.GET("/security/audit/findings", adminQAHandler.GetSecurityFindings)

			// 4. Benchmarks
			admin.POST("/benchmarks/run", adminQAHandler.RunBenchmark)
			admin.GET("/benchmarks/report", adminQAHandler.GetBenchmarkReport)
			admin.GET("/benchmarks/history", adminQAHandler.GetBenchmarkHistory)

			// 5. Preflight Check
			admin.GET("/preflight/check", adminQAHandler.PreflightCheck)

			// 6. QA Stats
			admin.GET("/qa/stats", adminQAHandler.GetQAStats)
		}
		}

		// Internal endpoints (for monitoring systems - no auth required)
		internal := api.Group("/internal")
		{
			internal.GET("/health", adminHealthHandler.GetHealth)
			internal.GET("/health/full", adminLaunchHandler.GetFullHealth)
			internal.GET("/metrics", adminMetricsHandler.GetMetrics)
			
			// Edge ingestion endpoints (no auth - called by edge workers)
			internal.POST("/edge-click", edgeIngestHandler.IngestEdgeClick)
			internal.GET("/edge/offer/:trackingCode", edgeIngestHandler.GetOfferConfig)
			internal.GET("/edge/stats", edgeIngestHandler.GetEdgeStats)
		}
	}

	port := cfg.Port
	log.Printf("🚀 Server starting on port %s in %s mode", port, cfg.Environment)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
