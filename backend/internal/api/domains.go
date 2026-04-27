package api

import (
	"net/http"
	"github.com/labstack/echo/v4"
	"github.com/user/mailadmin/internal/audit"
	"github.com/user/mailadmin/internal/auth"
	"github.com/user/mailadmin/internal/db"
	"github.com/user/mailadmin/internal/models"
)

func RegisterDomainHandlers(g *echo.Group, secret string) {
	g.Use(auth.JWTMiddleware(secret))

	// Список доменов с реальной статистикой
	g.GET("", func(c echo.Context) error {
		var domains []models.Domain
		// Скрываем 'ALL', если не запрошено иначе
		if err := db.DB.Where("domain != ?", "ALL").Order("domain").Find(&domains).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch domains"})
		}

		type DomainStats struct {
			models.Domain
			MailboxesCount int   `json:"mailboxes_count"`
			AliasesCount   int   `json:"aliases_count"`
			QuotaUsed      int64 `json:"quota_used"`
		}

		var result []DomainStats
		for _, d := range domains {
			var mCount, aCount int64
			var qUsed int64

			// Считаем ящики
			db.DB.Model(&models.Mailbox{}).Where("domain = ?", d.Domain).Count(&mCount)
			// Считаем алиасы
			db.DB.Model(&models.Alias{}).Where("domain = ?", d.Domain).Count(&aCount)
			// Считаем занятую квоту (сумма всех ящиков)
			db.DB.Model(&models.Mailbox{}).Where("domain = ?", d.Domain).Select("SUM(quota)").Scan(&qUsed)

			result = append(result, DomainStats{
				Domain:         d,
				MailboxesCount: int(mCount),
				AliasesCount:   int(aCount),
				QuotaUsed:      qUsed,
			})
		}

		return c.JSON(http.StatusOK, result)
	})

	// Создание домена
	g.POST("", func(c echo.Context) error {
		var domain models.Domain
		if err := c.Bind(&domain); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}
		
		claims := c.Get("user").(*auth.Claims)
		
		// Логика создания
		if err := db.DB.Create(&domain).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create domain: " + err.Error()})
		}

		audit.Log(db.DB, claims.Username, domain.Domain, "create domain", domain.Description)

		return c.JSON(http.StatusCreated, domain)
	})

	// Редактирование домена
	g.PUT("/:id", func(c echo.Context) error {
		id := c.Param("id")
		var domain models.Domain
		if err := db.DB.Where("domain = ?", id).First(&domain).Error; err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "domain not found"})
		}

		if err := c.Bind(&domain); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}

		claims := c.Get("user").(*auth.Claims)

		if err := db.DB.Save(&domain).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update domain"})
		}

		audit.Log(db.DB, claims.Username, domain.Domain, "update domain", domain.Description)

		return c.JSON(http.StatusOK, domain)
	})

	// Удаление домена
	g.DELETE("/:id", func(c echo.Context) error {
		id := c.Param("id")
		claims := c.Get("user").(*auth.Claims)

		if err := db.DB.Where("domain = ?", id).Delete(&models.Domain{}).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete domain"})
		}

		audit.Log(db.DB, claims.Username, id, "delete domain", "")

		return c.JSON(http.StatusNoContent, nil)
	})
}
