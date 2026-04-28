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

		// Вычисляем домен и проверяем доступ
		var domain string
		for i := len(username) - 1; i >= 0; i-- {
			if username[i] == '@' {
				domain = username[i+1:]
				break
			}
		}

		claims := c.Get("user").(*auth.Claims)
		if domain == "" || !hasDomainAccess(claims, domain) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied to this domain"})
		}

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

		type UpdateRequest struct {
			Subject      string    `json:"subject" validate:"required"`
			Body         string    `json:"body" validate:"required"`
			Active       bool      `json:"active"`
			ActiveFrom   time.Time `json:"activefrom"`
			ActiveUntil  time.Time `json:"activeuntil"`
			IntervalTime int       `json:"interval_time"`
		}

		var req UpdateRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}
		if err := c.Validate(&req); err != nil {
			return err
		}

		// Проверяем существование ящика
		var mailbox models.Mailbox
		if err := db.DB.Where("username = ?", username).First(&mailbox).Error; err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "mailbox not found"})
		}

		if !hasDomainAccess(claims, mailbox.Domain) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied to this domain"})
		}

		now := time.Now()
		// UPSERT логика
		var existing models.Vacation
		if err := db.DB.Where("email = ?", username).First(&existing).Error; err != nil {
			newVac := models.Vacation{
				Email:        username,
				Domain:       mailbox.Domain,
				Subject:      req.Subject,
				Body:         req.Body,
				Active:       req.Active,
				ActiveFrom:   req.ActiveFrom,
				ActiveUntil:  req.ActiveUntil,
				IntervalTime: req.IntervalTime,
				Created:      now,
				Modified:     now,
				Cache:        "",
			}
			if err := db.DB.Create(&newVac).Error; err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create vacation"})
			}
			audit.Log(db.DB, claims.Username, mailbox.Domain, "enable vacation", username)
			return c.JSON(http.StatusOK, newVac)
		} else {
			updates := map[string]interface{}{
				"subject":       req.Subject,
				"body":          req.Body,
				"active":        req.Active,
				"activefrom":    req.ActiveFrom,
				"activeuntil":   req.ActiveUntil,
				"interval_time": req.IntervalTime,
				"modified":      now,
			}
			if err := db.DB.Model(&existing).Updates(updates).Error; err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update vacation"})
			}
			audit.Log(db.DB, claims.Username, mailbox.Domain, "update vacation", username)
			return c.JSON(http.StatusOK, existing)
		}
	})
}
