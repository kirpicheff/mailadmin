package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DBDSN       string
	JWTSecret   string
	ListenAddr  string
	CORSOrigins []string
}

func LoadConfig() *Config {
	// Загружаем .env файл если он есть
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "user:password@tcp(127.0.0.1:3306)/mailadmin?charset=utf8mb4&parseTime=True&loc=Local"
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "change-me-in-production-min-32-bytes!!"
	}

	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	// Читаем CORS_ORIGIN и разбиваем его по запятой
	corsRaw := os.Getenv("CORS_ORIGIN")
	var origins []string
	if corsRaw != "" {
		parts := strings.Split(corsRaw, ",")
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				origins = append(origins, trimmed)
			}
		}
	}

	// Дефолт, если ничего не указано
	if len(origins) == 0 {
		origins = []string{"http://localhost:5173"}
	}

	// В production (не localhost) секреты обязаны быть заданы через env
	if addr != ":8080" && addr != "127.0.0.1:8080" {
		if os.Getenv("JWT_SECRET") == "" {
			log.Fatal("SECURITY: JWT_SECRET must be set via environment variable in production")
		}
		if os.Getenv("DB_DSN") == "" {
			log.Fatal("SECURITY: DB_DSN must be set via environment variable in production")
		}
		if os.Getenv("CORS_ORIGIN") == "" {
			log.Println("WARNING: CORS_ORIGIN not set, using default:", origins)
		}
	}

	return &Config{
		DBDSN:       dsn,
		JWTSecret:   secret,
		ListenAddr:  addr,
		CORSOrigins: origins,
	}
}



