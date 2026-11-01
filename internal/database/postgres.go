package database

import (
	"fmt"
	"log"
	"strings"
	"time"

	"proceive/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ensureDatabaseExists memeriksa apakah database target sudah ada di server PostgreSQL.
// Jika belum ada, fungsi ini otomatis membuat database tersebut (misal database 'proceive').
func ensureDatabaseExists(dbCfg *config.DatabaseConfig) {
	// Lewati jika menggunakan DATABASE_URL atau jika nama database adalah bawaan PostgreSQL/Railway
	if dbCfg.URL != "" || strings.EqualFold(dbCfg.Name, "postgres") || strings.EqualFold(dbCfg.Name, "railway") {
		return
	}

	// Buka koneksi maintenance ke database default 'postgres'
	maintenanceDSN := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s TimeZone=Asia/Jakarta",
		dbCfg.Host,
		dbCfg.Port,
		dbCfg.User,
		dbCfg.Password,
		dbCfg.SSLMode,
	)

	maintDB, err := gorm.Open(postgres.Open(maintenanceDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		// Jika koneksi maintenance gagal (misal di cloud terkelola tanpa akses db 'postgres'), lewati secara graceful
		return
	}

	sqlDB, err := maintDB.DB()
	if err == nil {
		defer sqlDB.Close()
	}

	// Cek apakah database target sudah ada
	var exists bool
	maintDB.Raw("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = ?)", dbCfg.Name).Scan(&exists)

	if !exists {
		log.Printf("📦 Database '%s' belum ditemukan di server. Membuat database otomatis...", dbCfg.Name)
		createSQL := fmt.Sprintf(`CREATE DATABASE "%s"`, strings.ReplaceAll(dbCfg.Name, `"`, `""`))
		if err := maintDB.Exec(createSQL).Error; err != nil {
			log.Printf("⚠️ Info: Pembuatan database otomatis dilewati (%v)", err)
		} else {
			log.Printf("🎉 Database '%s' berhasil dibuat secara otomatis!", dbCfg.Name)
		}
	}
}

// NewPostgresConnection membuat dan mengembalikan koneksi GORM ke PostgreSQL.
// Mendukung koneksi via DATABASE_URL (Railway) maupun parameter individual (DB_HOST/PGHOST, dll).
func NewPostgresConnection(cfg *config.Config) *gorm.DB {
	dbCfg := cfg.Database

	logLevel := logger.Silent
	if cfg.App.Env == "development" {
		logLevel = logger.Info
	}

	var dsn string
	if dbCfg.URL != "" {
		// Menggunakan DATABASE_URL bawaan Railway atau cloud provider
		dsn = dbCfg.URL
		log.Println("🔌 Terhubung menggunakan DATABASE_URL (Railway/Cloud Connection String)")
	} else {
		// Pastikan database target sudah ada sebelum mencoba menghubungkan
		ensureDatabaseExists(&dbCfg)

		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Jakarta",
			dbCfg.Host,
			dbCfg.Port,
			dbCfg.User,
			dbCfg.Password,
			dbCfg.Name,
			dbCfg.SSLMode,
		)
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
