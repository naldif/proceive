package main

import (
	"log"

	"proceive/internal/config"
	"proceive/internal/database"
)

func main() {
	log.Println("🚀 Menjalankan migrasi database Proceive...")

	// 1. Muat konfigurasi .env
	cfg := config.LoadConfig()

	// 2. Buka koneksi ke PostgreSQL
	db := database.NewPostgresConnection(cfg)

	// 3. Eksekusi migrasi skema
	if err := database.Migrate(db); err != nil {
		log.Fatalf("❌ Migrasi database gagal: %v", err)
	}

	log.Println("🎉 Migrasi database Proceive selesai dengan sukses!")
}
