package main

import (
	"fmt"
	"log"
	"net/http"

	"proceive/internal/config"
	"proceive/internal/database"
	"proceive/internal/routes"
)

func main() {
	// Load konfigurasi dari .env
	cfg := config.LoadConfig()

	// Inisialisasi koneksi database
	db := database.NewPostgresConnection(cfg)

	// Auto-migrate jika diaktifkan di konfigurasi
	if cfg.Database.AutoMigrate {
		if err := database.Migrate(db); err != nil {
			log.Fatalf("❌ Auto migrate gagal: %v", err)
		}
	}

	// Inisialisasi master data Proceive (idempotent seeder)
	if err := database.Seed(db); err != nil {
		log.Printf("⚠️ Warning saat seeding master data: %v", err)
	}

	// Setup Gin router dengan semua modul
	router := routes.Setup(db, cfg)

	// Jalankan server
	addr := fmt.Sprintf(":%s", cfg.App.Port)
	log.Printf("🚀 Proceive API berjalan di http://localhost%s (env: %s)", addr, cfg.App.Env)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("❌ Gagal menjalankan server: %v", err)
	}
}
