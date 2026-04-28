package api

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/user/mailadmin/internal/auth"
	"github.com/user/mailadmin/internal/db"
	"github.com/user/mailadmin/internal/models"
)

// RegisterLogHandlers регистрирует маршруты для логов аудита
func RegisterLogHandlers(g *echo.Group, secret string) {
	logs := g.Group("/logs")
	logs.Use(auth.JWTMiddleware(secret))

	logs.GET("", func(c echo.Context) error {
		page, _ := strconv.Atoi(c.QueryParam("page"))
		limit, _ := strconv.Atoi(c.QueryParam("limit"))
		domain := c.QueryParam("domain")

		if page <= 0 {
			page = 1
		}
		if limit <= 0 {
			limit = 50
		}
		offset := (page - 1) * limit

		claims := c.Get("user").(*auth.Claims)
		dbQuery := db.DB.Model(&models.Log{})

		// Фильтрация по правам
		if !claims.SuperAdmin {
			dbQuery = dbQuery.Where("domain IN (?) OR domain = 'system'", db.DB.Table("domain_admins").Select("domain").Where("username = ?", claims.Username))
		}

		if domain != "" {
			dbQuery = dbQuery.Where("domain = ?", domain)
		}

		var total int64
		dbQuery.Count(&total)

		var logs []models.Log
		if err := dbQuery.Offset(offset).Limit(limit).Order("timestamp DESC").Find(&logs).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch logs"})
		}

		c.Response().Header().Set("X-Total-Count", strconv.FormatInt(total, 10))
		return c.JSON(http.StatusOK, logs)
	})
}
