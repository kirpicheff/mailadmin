package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/GehirnInc/crypt/sha512_crypt"
	"github.com/labstack/echo/v4"
	"github.com/user/mailadmin/internal/audit"
	"github.com/user/mailadmin/internal/auth"
	"github.com/user/mailadmin/internal/db"
	"github.com/user/mailadmin/internal/models"
)

// RegisterMailboxHandlers регистрирует маршруты управления почтовыми ящиками
func RegisterMailboxHandlers(g *echo.Group, secret string) {
	mailboxes := g.Group("/mailboxes")
	mailboxes.Use(auth.JWTMiddleware(secret))

	// Список ящиков (с фильтром по домену)
	mailboxes.GET("", func(c echo.Context) error {
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
		dbQuery := db.DB.Model(&models.Mailbox{})

		// Фильтр по правам админа
		if !claims.SuperAdmin {
			dbQuery = dbQuery.Where("domain IN (?)", db.DB.Table("domain_admins").Select("domain").Where("username = ?", claims.Username))
		}

		// Если есть поиск - ищем по всем доступным доменам
		if search != "" {
			s := "%" + search + "%"
			dbQuery = dbQuery.Where("username LIKE ? OR name LIKE ?", s, s)
		} else if domain != "" {
			dbQuery = dbQuery.Where("domain = ?", domain)
		} else {
			// Если домен не указан и поиска нет - ничего не возвращаем или первый домен
			return c.JSON(http.StatusOK, []models.Mailbox{})
		}

		var total int64
		dbQuery.Count(&total)

		type MailboxWithQuota struct {
			models.Mailbox
			QuotaUsed int64 `json:"quota_used"`
			Messages  int   `json:"messages"`
		}

		var mailboxes []MailboxWithQuota

		// Используем алиасы m и q для ясности
		err := dbQuery.
			Select("mailbox.*, COALESCE(quota2.bytes, 0) as quota_used, COALESCE(quota2.messages, 0) as messages").
			Joins("LEFT JOIN quota2 ON quota2.username = mailbox.username").
			Offset(offset).Limit(limit).
			Order("mailbox.username ASC").
			Scan(&mailboxes).Error

		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch mailboxes: "+err.Error())
		}

		c.Response().Header().Set("X-Total-Count", strconv.FormatInt(total, 10))
		return c.JSON(http.StatusOK, mailboxes)
	})

	// Создание ящика
	mailboxes.POST("", func(c echo.Context) error {
		type CreateRequest struct {
			Username    string `json:"username" validate:"required,email"`
			Password    string `json:"password" validate:"required,min=8"`
			Name        string `json:"name"`
			Quota       int64  `json:"quota" validate:"min=0"`
			Active      bool   `json:"active"`
			Phone       string `json:"phone"`
			EmailOther  string `json:"email_other" validate:"omitempty,email"`
		}
		var req CreateRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}
		if err := c.Validate(&req); err != nil {
			return err
		}

		// Валидация username (должен быть email) - мы уже проверили тегом email, 
		// но нам нужны части для maildir
		parts := strings.Split(req.Username, "@")
		localPart := parts[0]
		domain := parts[1]

		claims := c.Get("user").(*auth.Claims)
		if !hasDomainAccess(claims, domain) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied to this domain"})
		}

		// Проверка лимита ящиков на домене
		var domainRec models.Domain
		if err := db.DB.Where("domain = ?", domain).First(&domainRec).Error; err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "domain not found"})
		}

		if domainRec.Mailboxes > 0 {
			var count int64
			db.DB.Model(&models.Mailbox{}).Where("domain = ?", domain).Count(&count)
			if count >= int64(domainRec.Mailboxes) {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "domain mailbox limit reached"})
			}
		}

		// Хеширование пароля
		hash, err := auth.GenerateHash(req.Password)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "hashing failed"})
		}

		box := models.Mailbox{
			Username:   req.Username,
			Password:   hash,
			Name:       req.Name,
			Quota:      req.Quota,
			LocalPart:  localPart,
			Domain:     domain,
			Maildir:    fmt.Sprintf("%s/%s/", domain, localPart),
			Active:     req.Active,
			Phone:      req.Phone,
			EmailOther: req.EmailOther,
			Created:    time.Now(),
			Modified:   time.Now(),
		}

		// Транзакция: создание ящика + алиаса
		tx := db.DB.Begin()

		if err := tx.Create(&box).Error; err != nil {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create mailbox: " + err.Error()})
		}

		// Создаем зеркальный алиас (как делает PostfixAdmin)
		alias := models.Alias{
			Address:  box.Username,
			Goto:     box.Username,
			Domain:   box.Domain,
			Created:  time.Now(),
			Modified: time.Now(),
			Active:   true,
		}

		if err := tx.Where("address = ?", box.Username).FirstOrCreate(&alias).Error; err != nil {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create associated alias: " + err.Error()})
		}

		tx.Commit()

		audit.Log(db.DB, claims.Username, box.Domain, "create mailbox + alias", box.Username)

		return c.JSON(http.StatusCreated, box)
	})

	// Обновление ящика
	mailboxes.PUT("/:username", func(c echo.Context) error {
		username := c.Param("username")
		var existing models.Mailbox
		if err := db.DB.Where("username = ?", username).First(&existing).Error; err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "mailbox not found"})
		}

		claims := c.Get("user").(*auth.Claims)
		if !hasDomainAccess(claims, existing.Domain) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied to this domain"})
		}

		type UpdateRequest struct {
			Password   string `json:"password" validate:"omitempty,min=8"`
			Name       string `json:"name"`
			Quota      int64  `json:"quota" validate:"min=0"`
			Active     bool   `json:"active"`
			Phone      string `json:"phone"`
			EmailOther string `json:"email_other" validate:"omitempty,email"`
		}
		var req UpdateRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}
		if err := c.Validate(&req); err != nil {
			return err
		}

		updates := map[string]interface{}{
			"name":        req.Name,
			"quota":       req.Quota,
			"active":      req.Active,
			"phone":       req.Phone,
			"email_other": req.EmailOther,
			"modified":    time.Now(),
		}

		// Если пароль передан - хешируем безопасной случайной солью
		if req.Password != "" {
			hash, err := auth.GenerateHash(req.Password)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "hashing failed"})
			}
			updates["password"] = hash
		}

		// Используем Updates с map или структурой, чтобы избежать затирания полей
		if err := db.DB.Model(&existing).Updates(updates).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update mailbox"})
		}

		audit.Log(db.DB, claims.Username, existing.Domain, "update mailbox", existing.Username)

		return c.JSON(http.StatusOK, existing)
	})

	// Удаление ящика
	mailboxes.DELETE("/:username", func(c echo.Context) error {
		username := c.Param("username")
		// Получаем домен для лога перед удалением
		var box models.Mailbox
		if err := db.DB.Select("domain").Where("username = ?", username).First(&box).Error; err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "mailbox not found"})
		}

		claims := c.Get("user").(*auth.Claims)
		if !hasDomainAccess(claims, box.Domain) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied to this domain"})
		}

		tx := db.DB.Begin()

		if err := tx.Where("username = ?", username).Delete(&models.Mailbox{}).Error; err != nil {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete mailbox"})
		}

		// Бережно удаляем и связанный алиас
		if err := tx.Where("address = ?", username).Delete(&models.Alias{}).Error; err != nil {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete associated alias"})
		}

		// Удаляем автоответчик, если есть
		if err := tx.Where("email = ?", username).Delete(&models.Vacation{}).Error; err != nil {
			tx.Rollback()
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete associated vacation"})
		}

		tx.Commit()

		audit.Log(db.DB, claims.Username, box.Domain, "delete mailbox + alias + vacation", username)

		return c.NoContent(http.StatusNoContent)
	})
}
