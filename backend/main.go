package main

import (
	"flag"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/user/mailadmin/internal/agent"
	"github.com/user/mailadmin/internal/api"
	"github.com/user/mailadmin/internal/config"
	"github.com/user/mailadmin/internal/db"
	"golang.org/x/time/rate"
)

func main() {
	isWeb := flag.Bool("web", false, "Run as web node")
	isAgent := flag.Bool("agent", false, "Run as agent node")
	flag.Parse()

	_ = isWeb // Игнорируем неиспользуемую переменную, она служит как флаг документации

	if *isAgent {
		agent.Start()
		return
	}

	// По умолчанию или если указан --web запускаем веб-интерфейс
	
	// Загружаем конфиг
	cfg := config.LoadConfig()

	// Инициализируем БД
	db.InitDB(cfg.DBDSN)

	e := echo.New()

	// Регистрация валидатора
	e.Validator = api.NewValidator()

	// Middleware: базовые
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// MED-2: Заголовки безопасности
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:      "1; mode=block",
		ContentTypeNosniff: "nosniff",
		XFrameOptions:      "DENY",
		HSTSMaxAge:         31536000,
	}))

	// HIGH-1: CORS-origins из конфигурации (поддерживает список доменов)
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     cfg.CORSOrigins,
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowCredentials: true,
		ExposeHeaders:    []string{"X-Total-Count"},
	}))

	// HIGH-2: Глобальный rate limit (20 запросов/сек на IP)
	e.Use(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{
				Rate:      rate.Limit(20),
				Burst:     40,
				ExpiresIn: 3 * time.Minute,
			},
		),
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			return ctx.RealIP(), nil
		},
		ErrorHandler: func(context echo.Context, err error) error {
			return context.JSON(http.StatusForbidden, map[string]string{"error": "rate limit exceeded"})
		},
		DenyHandler: func(context echo.Context, identifier string, err error) error {
			return context.JSON(http.StatusTooManyRequests, map[string]string{"error": "too many requests, slow down"})
		},
	}))

	// Группа API
	apiGroup := e.Group("/api")

	// Тестовый маршрут (без auth)
	apiGroup.GET("/status", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "online",
			"message": "MailAdmin API is running",
		})
	})

	// Маршруты авторизации
	authGroup := apiGroup.Group("/auth")
	api.RegisterAuthHandlers(authGroup, cfg)

	// Маршруты управления админами
	api.RegisterAdminHandlers(apiGroup, cfg.JWTSecret)

	// Маршруты управления доменами
	domainGroup := apiGroup.Group("/domains")
	api.RegisterDomainHandlers(domainGroup, cfg.JWTSecret)

	// Маршруты управления ящиками
	api.RegisterMailboxHandlers(apiGroup, cfg.JWTSecret)

	// Маршруты управления алиасами
	api.RegisterAliasHandlers(apiGroup, cfg.JWTSecret)

	// Логи и статистика
	api.RegisterLogHandlers(apiGroup, cfg.JWTSecret)
	api.RegisterStatsHandlers(apiGroup, cfg.JWTSecret)
	api.RegisterSystemHandlers(apiGroup, cfg.JWTSecret)

	// Инструменты и Автоответчик
	api.RegisterToolsHandlers(apiGroup, cfg.JWTSecret)
	api.RegisterVacationHandlers(apiGroup, cfg.JWTSecret)

	// Диагностика
	diagGroup := apiGroup.Group("/diagnostics")
	api.RegisterDiagnosticsHandlers(diagGroup, cfg.JWTSecret)

	// Управление Sieve-фильтрами
	sieveGroup := apiGroup.Group("/sieve")
	api.RegisterSieveHandlers(sieveGroup, cfg.JWTSecret, cfg)

	// Запуск сервера
	e.Logger.Fatal(e.Start(cfg.ListenAddr))
}


