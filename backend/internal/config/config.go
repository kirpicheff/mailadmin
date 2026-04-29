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
	MailRoot    string
	SieveRoot        string
	LogPath          string
	DovecotConfigDir string
}

func LoadConfig() *Config {
	// ... (loading .env)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	dsn := os.Getenv("DB_DSN")
	// ... (dsn logic)
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

	// Получаем пути к почте и sieve
	mailRoot := os.Getenv("MAIL_ROOT")
	if mailRoot == "" {
		mailRoot = "/data/mail"
	}
	sieveRoot := os.Getenv("SIEVE_ROOT")
	if sieveRoot == "" {
		sieveRoot = "/data/sieve"
	}
	logPath := os.Getenv("LOG_PATH")
	if logPath == "" {
		logPath = "/var/log/mail.log"
	}
	dovecotConfigDir := os.Getenv("DOVECOT_CONFIG_DIR")
	if dovecotConfigDir == "" {
		dovecotConfigDir = "/etc/dovecot"
	}

	// ... (cors logic)
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
		MailRoot:    mailRoot,
		SieveRoot:        sieveRoot,
		LogPath:          logPath,
		DovecotConfigDir: dovecotConfigDir,
	}
}



