package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/user/mailadmin/internal/auth"
	"github.com/user/mailadmin/internal/db"
	"github.com/user/mailadmin/internal/models"
)

// RegisterStatsHandlers регистрирует маршруты для статистики
func RegisterStatsHandlers(g *echo.Group, secret string) {
	stats := g.Group("/stats")
	stats.Use(auth.JWTMiddleware(secret))

	stats.GET("/dashboard", func(c echo.Context) error {
		claims := c.Get("user").(*auth.Claims)

		var domainsCount, mailboxesCount, aliasesCount int64
		var quotaLimit, quotaUsed int64

		// 1. Считаем домены
		dQuery := db.DB.Model(&models.Domain{}).Where("domain != ?", "ALL")
		if !claims.SuperAdmin {
			dQuery = dQuery.Where("domain IN (?)", db.DB.Table("domain_admins").Select("domain").Where("username = ?", claims.Username))
		}
		dQuery.Count(&domainsCount)

		// 2. Считаем ящики и их суммарную квоту
		mQuery := db.DB.Model(&models.Mailbox{})
		if !claims.SuperAdmin {
			mQuery = mQuery.Where("domain IN (?)", db.DB.Table("domain_admins").Select("domain").Where("username = ?", claims.Username))
		}
		mQuery.Count(&mailboxesCount)
		
		// Считаем сумму квот, исключая 0 (безлимит), так как они искажают процент заполнения
		mQuery.Select("COALESCE(SUM(quota), 0)").Where("quota > 0").Scan(&quotaLimit)

		// 3. Считаем алиасы
		aQuery := db.DB.Model(&models.Alias{})
		if !claims.SuperAdmin {
			aQuery = aQuery.Where("domain IN (?)", db.DB.Table("domain_admins").Select("domain").Where("username = ?", claims.Username))
		}
		aQuery.Count(&aliasesCount)

		// 4. Расчет использованной квоты
		if claims.SuperAdmin {
			db.DB.Model(&models.Quota2{}).Select("COALESCE(SUM(bytes), 0)").Scan(&quotaUsed)
		} else {
			allowedDomains := db.DB.Table("domain_admins").Select("domain").Where("username = ?", claims.Username)
			db.DB.Table("quota2").
				Joins("JOIN mailbox ON mailbox.username = quota2.username").
				Where("mailbox.domain IN (?)", allowedDomains).
				Select("COALESCE(SUM(quota2.bytes), 0)").Scan(&quotaUsed)
		}

		// 5. Последние 10 логов
		var recentLogs []models.Log
		lQuery := db.DB.Model(&models.Log{})
		if !claims.SuperAdmin {
			lQuery = lQuery.Where("domain IN (?) OR domain = 'system'", db.DB.Table("domain_admins").Select("domain").Where("username = ?", claims.Username))
		}
		lQuery.Order("timestamp DESC").Limit(10).Find(&recentLogs)

		return c.JSON(http.StatusOK, map[string]interface{}{
			"domains_count":   domainsCount,
			"mailboxes_count": mailboxesCount,
			"aliases_count":   aliasesCount,
			"quota_limit":     quotaLimit,
			"quota_used":      quotaUsed,
			"recent_logs":     recentLogs,
		})
	})
}
