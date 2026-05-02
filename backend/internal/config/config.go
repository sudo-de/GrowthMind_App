package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	APIVersion         string
	DatabaseURL        string
	RedisURL           string
	JWTSecret          string
	JWTRefreshSecret   string
	AllowOrigins       string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURI  string
	SMTPHost           string
	SMTPPort           string
	SMTPUsername       string
	SMTPPassword       string
	SMTPFrom           string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}
	return &Config{
		Port:               getEnv("PORT", ""),
		APIVersion:         getEnv("API_VERSION", ""),
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		RedisURL:           getEnv("REDIS_URL", ""),
		JWTSecret:          getEnv("JWT_SECRET", "change-me-in-production"),
		JWTRefreshSecret:   getEnv("JWT_REFRESH_SECRET", "change-refresh-secret-in-production"),
		AllowOrigins:       getEnv("ALLOW_ORIGINS", "*"),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURI:  getEnv("GOOGLE_REDIRECT_URI", ""),
		SMTPHost:           getEnv("SMTP_HOST", ""),
		SMTPPort:           getEnv("SMTP_PORT", "587"),
		SMTPUsername:       getEnv("SMTP_USERNAME", ""),
		SMTPPassword:       getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:           getEnv("SMTP_FROM", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
