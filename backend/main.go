package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/user/mailadmin/internal/api"
	"github.com/user/mailadmin/internal/config"
	"github.com/user/mailadmin/internal/db"
	"net/http"
)

func main() {
	// Загружаем конфиг
	cfg := config.LoadConfig()

	// Инициализируем БД
	db.InitDB(cfg.DBDSN)

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:5174"}, // Для разработки
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

	// Группа API
	apiGroup := e.Group("/api")

	// Тестовый маршрут
	apiGroup.GET("/status", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "online",
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
	api.RegisterDomainHandlers(domainGroup)

	// Маршруты управления ящиками
	api.RegisterMailboxHandlers(apiGroup, cfg.JWTSecret)

	// Маршруты управления алиасами
	api.RegisterAliasHandlers(apiGroup, cfg.JWTSecret)

	// Запуск сервера
	e.Logger.Fatal(e.Start(cfg.ListenAddr))
}
