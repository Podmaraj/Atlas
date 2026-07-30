package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"edgecore/internal/config"
	appLogger "edgecore/internal/logger"
	"edgecore/internal/models"
)

// InitDB creates PostgreSQL database connection with connection pooling and GORM AutoMigrate
func InitDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	gormLogLevel := logger.Warn
	if appLogger.L() != nil {
		// Log GORM queries if in debug mode
		gormLogLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgresql database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve sql.DB instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Perform automatic migrations for domain models
	if err := AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to auto migrate database schemas: %w", err)
	}

	appLogger.Info("PostgreSQL database connection established & schema migrated successfully")
	return db, nil
}

// AutoMigrate migrates all GORM entities
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Organization{},
		&models.Tenant{},
		&models.User{},
		&models.Role{},
		&models.ApiKey{},
		&models.Service{},
		&models.ServiceInstance{},
		&models.Route{},
		&models.Plugin{},
		&models.GatewayNode{},
		&models.Certificate{},
		&models.AuditLog{},
		&models.RateLimitRule{},
	)
}
