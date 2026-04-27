package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/user/mailadmin/internal/config"
	"github.com/user/mailadmin/internal/db"
	"net/http"
)

func main() {
	// Загружаем конфиг
	cfg := config.LoadConfig()

	// Инициализируем БД (пока без запуска, если базы нет - упадет)
	// db.InitDB(cfg.DBDSN)

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Группа API
	api := e.Group("/api")

	// Тестовый маршрут
	api.GET("/status", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "online",
			"message": "MailAdmin API is running",
		})
	})

	// Запуск сервера
	e.Logger.Fatal(e.Start(cfg.ListenAddr))
}
