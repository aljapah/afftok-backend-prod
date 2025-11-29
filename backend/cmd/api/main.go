package main

import (
	"log"

	"github.com/afftok/backend/internal/cache"
	"github.com/afftok/backend/internal/config"
	"github.com/afftok/backend/internal/database"
	"github.com/afftok/backend/internal/handlers"
	"github.com/afftok/backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
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

	if err := database.AutoMigrate(db); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	router.Use(middleware.CORSMiddleware())

	authHandler := handlers.NewAuthHandler(db)
	userHandler := handlers.NewUserHandler(db)
	offerHandler := handlers.NewOfferHandler(db)
	networkHandler := handlers.NewNetworkHandler(db)
	postbackHandler := handlers.NewPostbackHandler(db)
	teamHandler := handlers.NewTeamHandler(db)
	badgeHandler := handlers.NewBadgeHandler(db)
	clickHandler := handlers.NewClickHandler(db)
	promoterHandler := handlers.NewPromoterHandler(db)
		googleAuthHandler := handlers.NewGoogleAuthHandler(db)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "AffTok API is running",
		})
	})

	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authHandler.Logout)
				auth.GET("/google/login", googleAuthHandler.BeginLogin)
				auth.GET("/google/callback", googleAuthHandler.HandleCallback)
		}

		api.GET("/c/:id", clickHandler.TrackClick)
		api.GET("/promoter/:id", promoterHandler.GetPromoterPage)
		api.POST("/rate-promoter", promoterHandler.RatePromoter)

		api.POST("/postback", postbackHandler.HandlePostback)

		api.GET("/offers", offerHandler.GetAllOffers)
		api.GET("/offers/:id", offerHandler.GetOffer)

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.GET("/auth/me", authHandler.GetMe)
			protected.PUT("/profile", userHandler.UpdateProfile)

			protected.GET("/users", userHandler.GetAllUsers)
			protected.GET("/users/:id", userHandler.GetUser)

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
				teams.GET("/:id", teamHandler.GetTeam)
				teams.POST("", teamHandler.CreateTeam)
				teams.POST("/:id/join", teamHandler.JoinTeam)
				teams.POST("/:id/leave", teamHandler.LeaveTeam)
			}

			badges := protected.Group("/badges")
			{
				badges.GET("", badgeHandler.GetAllBadges)
				badges.GET("/my", badgeHandler.GetMyBadges)
			}

			clicks := protected.Group("/clicks")
			{
				clicks.GET("/my", clickHandler.GetMyClicks)
				clicks.GET("/:id/stats", clickHandler.GetClickStats)
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

				admin.POST("/conversions/:id/approve", postbackHandler.ApproveConversion)
				admin.POST("/conversions/:id/reject", postbackHandler.RejectConversion)

				admin.POST("/badges", badgeHandler.CreateBadge)
				admin.PUT("/badges/:id", badgeHandler.UpdateBadge)
				admin.DELETE("/badges/:id", badgeHandler.DeleteBadge)
			}
		}
	}

	port := cfg.Port
	log.Printf("🚀 Server starting on port %s in %s mode", port, cfg.Environment)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
