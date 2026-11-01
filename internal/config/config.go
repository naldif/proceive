package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	App        AppConfig
	Database   DatabaseConfig
	JWT        JWTConfig
	Cloudinary CloudinaryConfig
}

type AppConfig struct {
	Name string
	Env  string
	Port string
}

type DatabaseConfig struct {
	URL         string
	Host        string
	Port        string
	User        string
	Password    string
	Name        string
	SSLMode     string
	AutoMigrate bool
}

type JWTConfig struct {
	Secret          string
	ExpirationHours int
}

type CloudinaryConfig struct {
	CloudName string
	APIKey    string
	APISecret string
}

// LoadConfig membaca environment variable dari file .env (jika ada) dan OS environment
func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Info: File .env tidak ditemukan, menggunakan environment variable sistem atau default")
	}

	expHours, err := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "24"))
	if err != nil {
		expHours = 24
	}

	return &Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "proceive"),
			Env:  getEnv("APP_ENV", "development"),
			Port: getEnv("PORT", getEnv("APP_PORT", "8080")), // Otomatis mendeteksi PORT dinamis dari Railway
		},
		Database: DatabaseConfig{
			URL:         getEnv("DATABASE_URL", getEnv("DATABASE_PUBLIC_URL", "")),
			Host:        getEnv("DB_HOST", getEnv("PGHOST", "localhost")),
			Port:        getEnv("DB_PORT", getEnv("PGPORT", "5432")),
			User:        getEnv("DB_USER", getEnv("PGUSER", "postgres")),
			Password:    getEnv("DB_PASSWORD", getEnv("PGPASSWORD", "postgres")),
			Name:        getEnv("DB_NAME", getEnv("PGDATABASE", "proceive")),
			SSLMode:     getEnv("DB_SSLMODE", "disable"),
			AutoMigrate: getEnv("DB_AUTO_MIGRATE", "true") == "true", // Default true agar otomatis migrate & seeder di Railway
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "proceive_super_secret_jwt_key_2026"),
			ExpirationHours: expHours,
		},
		Cloudinary: CloudinaryConfig{
			CloudName: getEnv("CLOUDINARY_CLOUD_NAME", ""),
			APIKey:    getEnv("CLOUDINARY_API_KEY", ""),
			APISecret: getEnv("CLOUDINARY_API_SECRET", ""),
		},
	}
}

func getEnv(key, fallback string) string {
	if val, exists := os.LookupEnv(key); exists && val != "" {
		return val
	}
	return fallback
}
