package config

import (
	"os"
)

type Config struct {
	DBDSN      string
	JWTSecret  string
	ListenAddr string
}

func LoadConfig() *Config {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		// Дефолт для разработки: user:pass@tcp(127.0.0.1:3306)/mailadmin?charset=utf8mb4&parseTime=True&loc=Local
		dsn = "postfix:password@tcp(192.168.35.4:3307)/postfix?charset=utf8mb4&parseTime=True&loc=Local"
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "mailadmin-super-secret-key-2026"
	}

	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	return &Config{
		DBDSN:      dsn,
		JWTSecret:  secret,
		ListenAddr: addr,
	}
}
