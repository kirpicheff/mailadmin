package api

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/user/mailadmin/internal/audit"
	"github.com/user/mailadmin/internal/auth"
	"github.com/user/mailadmin/internal/db"
	"github.com/user/mailadmin/internal/models"
)

// RegisterVacationHandlers регистрирует маршруты управления автоответчиком
func RegisterVacationHandlers(g *echo.Group, secret string) {
	vac := g.Group("/mailboxes/:username/vacation")
	vac.Use(auth.JWTMiddleware(secret))

	// Получение настроек автоответчика
	vac.GET("", func(c echo.Context) error {
		username := c.Param("username")
		var vacation models.Vacation
		
		result := db.DB.Where("email = ?", username).First(&vacation)
		if result.Error != nil {
			// Если записи нет - возвращаем пустую структуру со статусом деактивации
			return c.JSON(http.StatusOK, models.Vacation{
				Email:  username,
				Active: false,
			})
		}
		
		return c.JSON(http.StatusOK, vacation)
	})

	// Обновление настроек
	vac.PUT("", func(c echo.Context) error {
		username := c.Param("username")
		claims := c.Get("user").(*auth.Claims)
		
		var req models.Vacation
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}

		// Проверяем существование ящика
		var mailbox models.Mailbox
		if err := db.DB.Where("username = ?", username).First(&mailbox).Error; err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "mailbox not found"})
		}

		req.Email = username
		req.Domain = mailbox.Domain
		req.Modified = time.Now()
		
		// UPSERT логика
		var existing models.Vacation
		if err := db.DB.Where("email = ?", username).First(&existing).Error; err != nil {
			req.Created = time.Now()
			req.Cache = "" // Инициализируем кеш
			if err := db.DB.Create(&req).Error; err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create vacation"})
			}
		} else {
			if err := db.DB.Model(&existing).Updates(map[string]interface{}{
				"subject":       req.Subject,
				"body":          req.Body,
				"active":        req.Active,
				"activefrom":    req.ActiveFrom,
				"activeuntil":   req.ActiveUntil,
				"interval_time": req.IntervalTime,
				"modified":      req.Modified,
			}).Error; err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update vacation"})
			}
		}

		audit.Log(db.DB, claims.Username, mailbox.Domain, "update vacation", username)
		
		return c.JSON(http.StatusOK, req)
	})
}
