package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/user/mailadmin/internal/audit"
	"github.com/user/mailadmin/internal/auth"
	"github.com/user/mailadmin/internal/db"
	"github.com/user/mailadmin/internal/models"
)

// RegisterNotificationHandlers регистрирует маршруты управления уведомлениями
func RegisterNotificationHandlers(g *echo.Group, secret string) {
	notifications := g.Group("/system/notifications")
	notifications.Use(auth.JWTMiddleware(secret))

	// Проверка на суперадмина
	notifications.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims := c.Get("user").(*auth.Claims)
			if !claims.SuperAdmin {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "superadmin access required"})
			}
			return next(c)
		}
	})

	// Получить все правила уведомлений
	notifications.GET("", func(c echo.Context) error {
		var rules []models.NotificationRule
		if err := db.DB.Order("email, domain").Find(&rules).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch rules: " + err.Error()})
		}
		return c.JSON(http.StatusOK, rules)
	})

	// Создать или обновить правила для email
	notifications.POST("", func(c echo.Context) error {
		type SaveRequest struct {
			Email   string   `json:"email" validate:"required,email"`
			Domains []string `json:"domains" validate:"required"`
			Active  bool     `json:"active"`
		}
		var req SaveRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}
		if err := c.Validate(&req); err != nil {
			return err
		}

		claims := c.Get("user").(*auth.Claims)
		tx := db.DB.Begin()

		// Удаляем старые правила для этого email
		if err := tx.Where("email = ?", req.Email).Delete(&models.NotificationRule{}).Error; err != nil {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to clear old rules: " + err.Error()})
		}

		// Создаем новые правила для каждого домена
		for _, d := range req.Domains {
			rule := models.NotificationRule{
				Email:  req.Email,
				Domain: d,
				Active: req.Active,
			}
			if err := tx.Create(&rule).Error; err != nil {
				tx.Rollback()
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to save rule: " + err.Error()})
			}
		}

		tx.Commit()
		audit.Log(db.DB, claims.Username, "system", "update notification rules", req.Email)

		return c.JSON(http.StatusOK, map[string]string{"message": "rules updated successfully"})
	})

	// Удалить правила для конкретного email
	notifications.DELETE("", func(c echo.Context) error {
		email := c.QueryParam("email")
		if email == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "email parameter required"})
		}

		claims := c.Get("user").(*auth.Claims)
		if err := db.DB.Where("email = ?", email).Delete(&models.NotificationRule{}).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete rules: " + err.Error()})
		}

		audit.Log(db.DB, claims.Username, "system", "delete notification rules", email)
		return c.NoContent(http.StatusNoContent)
	})
}
