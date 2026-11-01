package main

import (
	"log"

	"proceive/internal/config"
	"proceive/internal/database"
)

func main() {
	log.Println("🚀 Menjalankan seeding master data Proceive...")

	// 1. Muat konfigurasi .env
	cfg := config.LoadConfig()

	// 2. Buka koneksi ke PostgreSQL
	db := database.NewPostgresConnection(cfg)

	// 3. Eksekusi migrasi skema terlebih dahulu agar seluruh kolom tabel siap
	if err := database.Migrate(db); err != nil {
		log.Printf("⚠️ Warning migrasi skema: %v", err)
	}

	// 4. Eksekusi seeding data
	if err := database.Seed(db); err != nil {
		log.Fatalf("❌ Seeding database gagal: %v", err)
	}

	log.Println("🎉 Seeding master data Proceive selesai dengan sukses!")
}
