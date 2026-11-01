package database

import (
	"fmt"
	"log"
	"time"

	"proceive/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgresConnection membuat dan mengembalikan koneksi GORM ke PostgreSQL.
func NewPostgresConnection(cfg *config.Config) *gorm.DB {
	dbCfg := cfg.Database
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Jakarta",
		dbCfg.Host,
		dbCfg.Port,
		dbCfg.User,
		dbCfg.Password,
		dbCfg.Name,
		dbCfg.SSLMode,
	)

	logLevel := logger.Silent
	if cfg.App.Env == "development" {
		logLevel = logger.Info
	}

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		log.Fatalf("❌ Gagal koneksi ke database PostgreSQL [%s:%s/%s]: %v", dbCfg.Host, dbCfg.Port, dbCfg.Name, err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatalf("❌ Gagal mengambil instance sql.DB: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Printf("✅ Berhasil terhubung ke database PostgreSQL [%s:%s/%s]", dbCfg.Host, dbCfg.Port, dbCfg.Name)
	return gormDB
}
