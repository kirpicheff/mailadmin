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

func RegisterDomainHandlers(g *echo.Group, secret string) {
	g.Use(auth.JWTMiddleware(secret))

	g.GET("", func(c echo.Context) error {
		var domains []models.Domain

		claims := c.Get("user").(*auth.Claims)
		dbQuery := db.DB.Where("domain != ?", "ALL").Model(&models.Domain{})

		// Фильтр по поиску
		if search := c.QueryParam("search"); search != "" {
			dbQuery = dbQuery.Where("(domain LIKE ? OR description LIKE ?)", "%"+search+"%", "%"+search+"%")
		}

		// Фильтр по статусу
		if active := c.QueryParam("active"); active != "" {
			dbQuery = dbQuery.Where("active = ?", active == "true")
		}

		if !claims.SuperAdmin {
			dbQuery = dbQuery.Where("domain IN (?)", db.DB.Table("domain_admins").Select("domain").Where("username = ?", claims.Username))
		}

		if err := dbQuery.Order("domain").Find(&domains).Error; err != nil {
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
		type CreateRequest struct {
			Domain      string `json:"domain" validate:"required,fqdn"`
			Description string `json:"description"`
			Aliases     int    `json:"aliases" validate:"min=0"`
			Mailboxes   int    `json:"mailboxes" validate:"min=0"`
			MaxQuota    int64  `json:"maxquota" validate:"min=0"`
			Quota       int64  `json:"quota" validate:"min=0"`
			Transport   string `json:"transport"`
			BackupMX    bool   `json:"backupmx"`
			Active      bool   `json:"active"`
		}
		var req CreateRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}
		if err := c.Validate(&req); err != nil {
			return err
		}

		claims := c.Get("user").(*auth.Claims)
		if !claims.SuperAdmin {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "only superadmins can create domains"})
		}

		domain := models.Domain{
			Domain:      req.Domain,
			Description: req.Description,
			Aliases:     req.Aliases,
			Mailboxes:   req.Mailboxes,
			MaxQuota:    req.MaxQuota,
			Quota:       req.Quota,
			Transport:   req.Transport,
			BackupMX:    req.BackupMX,
			Created:     time.Now(),
			Modified:    time.Now(),
			Active:      req.Active,
		}

		// Логика создания с транзакцией
		tx := db.DB.Begin()

		if err := tx.Create(&domain).Error; err != nil {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create domain: " + err.Error()})
		}

		// Автосоздание стандартных RFC-алиасов для нового домена, пересылающих почту текущему админу
		defaultAliases := []string{"postmaster", "abuse", "hostmaster"}
		for _, prefix := range defaultAliases {
			alias := models.Alias{
				Address:  prefix + "@" + domain.Domain,
				Goto:     claims.Username,
				Domain:   domain.Domain,
				Created:  time.Now(),
				Modified: time.Now(),
				Active:   true,
			}
			tx.FirstOrCreate(&alias, models.Alias{Address: alias.Address})
		}

		tx.Commit()

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

		type UpdateRequest struct {
			Description string `json:"description"`
			Aliases     int    `json:"aliases" validate:"min=0"`
			Mailboxes   int    `json:"mailboxes" validate:"min=0"`
			MaxQuota    int64  `json:"maxquota" validate:"min=0"`
			Quota       int64  `json:"quota" validate:"min=0"`
			Transport   string `json:"transport"`
			BackupMX    bool   `json:"backupmx"`
			Active      bool   `json:"active"`
		}
		var req UpdateRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}
		if err := c.Validate(&req); err != nil {
			return err
		}

		claims := c.Get("user").(*auth.Claims)
		if !claims.SuperAdmin {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "only superadmins can edit domains"})
		}

		updates := map[string]interface{}{
			"description": req.Description,
			"aliases":     req.Aliases,
			"mailboxes":   req.Mailboxes,
			"maxquota":    req.MaxQuota,
			"quota":       req.Quota,
			"transport":   req.Transport,
			"backupmx":    req.BackupMX,
			"active":      req.Active,
			"modified":    time.Now(),
		}

		if err := db.DB.Model(&domain).Updates(updates).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update domain"})
		}

		audit.Log(db.DB, claims.Username, domain.Domain, "update domain", domain.Description)

		return c.JSON(http.StatusOK, domain)
	})

	// Удаление домена
	g.DELETE("/:id", func(c echo.Context) error {
		id := c.Param("id")
		claims := c.Get("user").(*auth.Claims)

		if !claims.SuperAdmin {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "only superadmins can delete domains"})
		}

		tx := db.DB.Begin()

		// Каскадное удаление "сирот"
		tx.Where("domain = ?", id).Delete(&models.Vacation{})
		tx.Where("domain = ?", id).Delete(&models.Mailbox{})
		tx.Where("domain = ?", id).Delete(&models.Alias{})
		tx.Where("target_domain = ? OR alias_domain = ?", id, id).Delete(&models.AliasDomain{})
		tx.Where("domain = ?", id).Delete(&models.DomainAdmin{})

		// Удаление самого домена
		if err := tx.Where("domain = ?", id).Delete(&models.Domain{}).Error; err != nil {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete domain"})
		}

		tx.Commit()

		audit.Log(db.DB, claims.Username, id, "delete domain", "")

		return c.JSON(http.StatusNoContent, nil)
	})
}
