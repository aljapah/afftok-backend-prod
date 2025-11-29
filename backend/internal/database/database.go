package database

import (
	"fmt"
	"log"
	"time"

	"github.com/afftok/backend/internal/config"
	"github.com/afftok/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

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

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db
	log.Println("✅ PostgreSQL connected successfully")
	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database connection not established")
	}

	db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")

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
	)

	if err != nil {
		return fmt.Errorf("failed to auto migrate: %w", err)
	}

	log.Println("✅ Database migration completed successfully")
	return nil
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
