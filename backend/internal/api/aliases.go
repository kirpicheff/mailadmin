package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/user/mailadmin/internal/audit"
	"github.com/user/mailadmin/internal/auth"
	"github.com/user/mailadmin/internal/db"
	"github.com/user/mailadmin/internal/models"
)

// RegisterAliasHandlers регистрирует маршруты управления алиасами
func RegisterAliasHandlers(g *echo.Group, secret string) {
	aliasGroup := g.Group("/aliases")
	aliasGroup.Use(auth.JWTMiddleware(secret))

	// Список алиасов
	aliasGroup.GET("", func(c echo.Context) error {
		domain := c.QueryParam("domain")
		search := c.QueryParam("search")
		page, _ := strconv.Atoi(c.QueryParam("page"))
		limit, _ := strconv.Atoi(c.QueryParam("limit"))

		if page <= 0 {
			page = 1
		}
		if limit <= 0 {
			limit = 50
		}
		offset := (page - 1) * limit

		claims := c.Get("user").(*auth.Claims)
		dbQuery := db.DB.Model(&models.Alias{}).Where("address != goto")

		if !claims.SuperAdmin {
			dbQuery = dbQuery.Where("domain IN (?)", db.DB.Table("domain_admins").Select("domain").Where("username = ?", claims.Username))
		}

		if search != "" {
			s := "%" + search + "%"
			dbQuery = dbQuery.Where("address LIKE ? OR goto LIKE ?", s, s)
		} else if domain != "" {
			dbQuery = dbQuery.Where("domain = ?", domain)
		} else {
			return c.JSON(http.StatusOK, []models.Alias{})
		}

		var total int64
		dbQuery.Count(&total)

		var aliases []models.Alias
		if err := dbQuery.Offset(offset).Limit(limit).Order("address ASC").Find(&aliases).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch aliases"})
		}

		c.Response().Header().Set("X-Total-Count", strconv.FormatInt(total, 10))
		return c.JSON(http.StatusOK, aliases)
	})

	// Создание алиаса
	aliasGroup.POST("", func(c echo.Context) error {
		type CreateRequest struct {
			Address string `json:"address" validate:"required,email_or_catchall"`
			Goto    string `json:"goto" validate:"required"` // Может быть списком через запятую
			Domain  string `json:"domain" validate:"required,fqdn"`
			Active  bool   `json:"active"`
		}
		var req CreateRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}
		if err := c.Validate(&req); err != nil {
			return err
		}

		// Проверяем, что домен в адресе совпадает с указанным доменом
		parts := strings.Split(req.Address, "@")
		if len(parts) != 2 || parts[1] != req.Domain {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "address domain must match request domain"})
		}

		claims := c.Get("user").(*auth.Claims)
		if !hasDomainAccess(claims, req.Domain) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied to this domain"})
		}

		if req.Goto == "[ALL_MAILBOXES]" && !claims.SuperAdmin {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "only superadmins can manage global aliases"})
		}

		alias := models.Alias{
			Address:  req.Address,
			Goto:     req.Goto,
			Domain:   req.Domain,
			Active:   req.Active,
			Created:  time.Now(),
			Modified: time.Now(),
		}

		if err := db.DB.Create(&alias).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create alias"})
		}

		audit.Log(db.DB, claims.Username, alias.Domain, "create alias", alias.Address)

		return c.JSON(http.StatusCreated, alias)
	})

	// Обновление алиаса
	aliasGroup.PUT("/:address", func(c echo.Context) error {
		address := c.Param("address")
		var existing models.Alias
		if err := db.DB.Where("address = ?", address).First(&existing).Error; err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "alias not found"})
		}

		claims := c.Get("user").(*auth.Claims)
		if !hasDomainAccess(claims, existing.Domain) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied to this domain"})
		}

		type UpdateRequest struct {
			Goto   string `json:"goto" validate:"required"`
			Active bool   `json:"active"`
		}
		var req UpdateRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}
		if err := c.Validate(&req); err != nil {
			return err
		}

		if (existing.Goto == "[ALL_MAILBOXES]" || req.Goto == "[ALL_MAILBOXES]") && !claims.SuperAdmin {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "only superadmins can manage global aliases"})
		}

		updates := map[string]interface{}{
			"goto":     req.Goto,
			"active":   req.Active,
			"modified": time.Now(),
		}

		if err := db.DB.Model(&existing).Updates(updates).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update alias"})
		}

		audit.Log(db.DB, claims.Username, existing.Domain, "update alias", existing.Address)

		return c.JSON(http.StatusOK, existing)
	})

	// Удаление алиаса
	aliasGroup.DELETE("/:address", func(c echo.Context) error {
		address := c.Param("address")
		var existing models.Alias
		if err := db.DB.Where("address = ?", address).First(&existing).Error; err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "alias not found"})
		}

		claims := c.Get("user").(*auth.Claims)
		if existing.Goto == "[ALL_MAILBOXES]" && !claims.SuperAdmin {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "only superadmins can manage global aliases"})
		}

		if !hasDomainAccess(claims, existing.Domain) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied to this domain"})
		}

		if err := db.DB.Where("address = ?", address).Delete(&models.Alias{}).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete alias"})
		}

		audit.Log(db.DB, claims.Username, existing.Domain, "delete alias", address)

		return c.NoContent(http.StatusNoContent)
	})

	// --- Алиасы доменов ---

	domainAliasGroup := aliasGroup.Group("/domain-aliases")

	domainAliasGroup.GET("", func(c echo.Context) error {
		domain := c.QueryParam("domain")
		search := c.QueryParam("search")

		claims := c.Get("user").(*auth.Claims)
		dbQuery := db.DB.Model(&models.AliasDomain{})

		if !claims.SuperAdmin {
			dbQuery = dbQuery.Where("target_domain IN (?)", db.DB.Table("domain_admins").Select("domain").Where("username = ?", claims.Username))
		}

		if search != "" {
			s := "%" + search + "%"
			dbQuery = dbQuery.Where("alias_domain LIKE ? OR target_domain LIKE ?", s, s)
		} else if domain != "" {
			dbQuery = dbQuery.Where("target_domain = ?", domain)
		} else {
			return c.JSON(http.StatusOK, []models.AliasDomain{})
		}

		var aliases []models.AliasDomain
		if err := dbQuery.Find(&aliases).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch domain aliases"})
		}
		return c.JSON(http.StatusOK, aliases)
	})

	domainAliasGroup.POST("", func(c echo.Context) error {
		var alias models.AliasDomain
		if err := c.Bind(&alias); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}

		claims := c.Get("user").(*auth.Claims)
		if !claims.SuperAdmin {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "only superadmins can create domain aliases"})
		}

		alias.Created = time.Now()
		alias.Modified = time.Now()
		if err := db.DB.Create(&alias).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create domain alias"})
		}

		audit.Log(db.DB, claims.Username, alias.TargetDomain, "create domain alias", alias.AliasDomain)

		return c.JSON(http.StatusCreated, alias)
	})

	domainAliasGroup.DELETE("/:alias_domain", func(c echo.Context) error {
		aliasDomain := c.Param("alias_domain")
		var existing models.AliasDomain
		db.DB.Select("target_domain").Where("alias_domain = ?", aliasDomain).First(&existing)

		claims := c.Get("user").(*auth.Claims)
		if !claims.SuperAdmin {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "only superadmins can delete domain aliases"})
		}

		if err := db.DB.Where("alias_domain = ?", aliasDomain).Delete(&models.AliasDomain{}).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete domain alias"})
		}

		audit.Log(db.DB, claims.Username, existing.TargetDomain, "delete domain alias", aliasDomain)

		return c.NoContent(http.StatusNoContent)
	})
}
