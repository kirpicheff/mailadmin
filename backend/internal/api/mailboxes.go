package api

import (
	"encoding/base64"
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
	"github.com/user/mailadmin/internal/mail"
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
			dbQuery = dbQuery.Where("mailbox.domain IN (?)", db.DB.Table("domain_admins").Select("domain").Where("username = ?", claims.Username))
		}

		// Если есть поиск - ищем по всем доступным доменам
		if search != "" {
			s := "%" + search + "%"
			dbQuery = dbQuery.Where("(mailbox.username LIKE ? OR mailbox.name LIKE ?)", s, s)
		} else if domain != "" {
			dbQuery = dbQuery.Where("mailbox.domain = ?", domain)
		} else {
			// Если домен не указан и поиска нет - ничего не возвращаем или первый домен
			return c.JSON(http.StatusOK, []models.Mailbox{})
		}

		// Фильтр по статусу
		if active := c.QueryParam("active"); active != "" {
			dbQuery = dbQuery.Where("mailbox.active = ?", active == "true")
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
			Username   string `json:"username" validate:"required,email"`
			Password   string `json:"password" validate:"required,min=8"`
			Name       string `json:"name"`
			Quota      int64  `json:"quota" validate:"min=0"`
			Active     bool   `json:"active"`
			Phone      string `json:"phone"`
			EmailOther string `json:"email_other" validate:"omitempty,email"`
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

		// Асинхронная отправка уведомлений
		go func(username, rawPassword, name, domainName string) {
			// 1. Рассылка суперадминам (с логином и паролем)
			var superAdmins []models.Admin
			if err := db.DB.Where("superadmin = ? AND active = ?", true, true).Find(&superAdmins).Error; err == nil {
				for _, admin := range superAdmins {
					// Проверяем, не отключил ли данный суперадмин рассылку паролей
					var opt models.Setting
					if err := db.DB.Where("setting_key = ?", "disable_pass_notif:"+admin.Username).First(&opt).Error; err == nil {
						if opt.Value == "true" {
							continue
						}
					}

					to := []string{admin.Username}
					if admin.EmailOther != "" {
						to = append(to, admin.EmailOther)
					}

					subject := fmt.Sprintf("Создан новый почтовый ящик: %s", username)
					body := fmt.Sprintf(
						"Здравствуйте!\r\n\r\nВ домене %s был создан новый почтовый ящик:\r\n"+
							"Адрес: %s\r\n"+
							"Пароль: %s\r\n"+
							"Владелец: %s\r\n\r\n"+
							"Сообщение создано автоматически почтовой панелью управления.",
						domainName, username, rawPassword, name,
					)

					encodedBody := base64.StdEncoding.EncodeToString([]byte(body))

					_ = mail.SendEmail(&mail.EmailMessage{
						From:    "noreply@" + domainName,
						To:      to,
						Subject: subject,
						Body:    encodedBody,
						IsHTML:  false,
					})
				}
			}

			// 2. Рассылка по правилам уведомлений (без пароля)
			var rules []models.NotificationRule
			if err := db.DB.Where("(domain = ? OR domain = ?) AND active = ?", domainName, "ALL", true).Find(&rules).Error; err == nil {
				emails := make(map[string]bool)
				for _, r := range rules {
					emails[r.Email] = true
				}

				for email := range emails {
					subject := fmt.Sprintf("Создан новый почтовый ящик: %s", username)
					body := fmt.Sprintf(
						"Здравствуйте!\r\n\r\nУведомляем вас о создании нового почтового ящика в домене %s:\r\n"+
							"Адрес: %s\r\n"+
							"Владелец: %s\r\n\r\n"+
							"Сообщение создано автоматически почтовой панелью управления.",
						domainName, username, name,
					)

					encodedBody := base64.StdEncoding.EncodeToString([]byte(body))

					_ = mail.SendEmail(&mail.EmailMessage{
						From:    "noreply@" + domainName,
						To:      []string{email},
						Subject: subject,
						Body:    encodedBody,
						IsHTML:  false,
					})
				}
			}
		}(box.Username, req.Password, box.Name, box.Domain)

		audit.Log(db.DB, claims.Username, box.Domain, "create mailbox", box.Username)

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

		// Если пароль изменен, асинхронно отправляем уведомление глобальным администраторам
		if req.Password != "" {
			go func(username, rawPassword, name, domainName string) {
				var superAdmins []models.Admin
				if err := db.DB.Where("superadmin = ? AND active = ?", true, true).Find(&superAdmins).Error; err == nil {
					for _, admin := range superAdmins {
						var opt models.Setting
						if err := db.DB.Where("setting_key = ?", "disable_pass_notif:"+admin.Username).First(&opt).Error; err == nil {
							if opt.Value == "true" {
								continue
							}
						}

						to := []string{admin.Username}
						if admin.EmailOther != "" {
							to = append(to, admin.EmailOther)
						}

						subject := fmt.Sprintf("Изменен пароль почтового ящика: %s", username)
						body := fmt.Sprintf(
							"Здравствуйте!\r\n\r\nВ домене %s был изменен пароль для почтового ящика:\r\n"+
								"Адрес: %s\r\n"+
								"Новый пароль: %s\r\n"+
								"Владелец: %s\r\n\r\n"+
								"Сообщение создано автоматически почтовой панелью управления.",
							domainName, username, rawPassword, name,
						)

						encodedBody := base64.StdEncoding.EncodeToString([]byte(body))

						_ = mail.SendEmail(&mail.EmailMessage{
							From:    "noreply@" + domainName,
							To:      to,
							Subject: subject,
							Body:    encodedBody,
							IsHTML:  false,
						})
					}
				}
			}(existing.Username, req.Password, existing.Name, existing.Domain)
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

		audit.Log(db.DB, claims.Username, box.Domain, "delete mailbox", username)

		return c.NoContent(http.StatusNoContent)
	})

	// Массовое создание ящиков
	mailboxes.POST("/batch/create", func(c echo.Context) error {
		type BatchCreateRequest struct {
			Domain   string   `json:"domain" validate:"required,fqdn"`
			Prefixes []string `json:"prefixes" validate:"required,min=1"`
			Password string   `json:"password" validate:"required,min=8"`
			Quota    int64    `json:"quota" validate:"min=0"`
			Active   bool     `json:"active"`
		}
		var req BatchCreateRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}

		claims := c.Get("user").(*auth.Claims)
		if !hasDomainAccess(claims, req.Domain) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
		}

		hash, _ := auth.GenerateHash(req.Password)
		tx := db.DB.Begin()

		createdCount := 0
		for _, prefix := range req.Prefixes {
			prefix = strings.TrimSpace(prefix)
			if prefix == "" {
				continue
			}
			username := prefix + "@" + req.Domain
			box := models.Mailbox{
				Username:  username,
				Password:  hash,
				Name:      prefix,
				Quota:     req.Quota,
				LocalPart: prefix,
				Domain:    req.Domain,
				Maildir:   fmt.Sprintf("%s/%s/", req.Domain, prefix),
				Active:    req.Active,
				Created:   time.Now(),
				Modified:  time.Now(),
			}

			if err := tx.Create(&box).Error; err != nil {
				continue
			}

			alias := models.Alias{
				Address:  username,
				Goto:     username,
				Domain:   req.Domain,
				Created:  time.Now(),
				Modified: time.Now(),
				Active:   true,
			}
			tx.FirstOrCreate(&alias, models.Alias{Address: username})
			createdCount++
		}

		tx.Commit()
		audit.Log(db.DB, claims.Username, req.Domain, "batch create mailboxes", fmt.Sprintf("%d items", createdCount))
		return c.JSON(http.StatusCreated, map[string]interface{}{"created": createdCount})
	})

	// Массовое удаление ящиков
	mailboxes.POST("/batch/delete", func(c echo.Context) error {
		type BatchRequest struct {
			Usernames []string `json:"usernames" validate:"required,min=1"`
		}
		var req BatchRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}

		claims := c.Get("user").(*auth.Claims)
		tx := db.DB.Begin()

		for _, username := range req.Usernames {
			var box models.Mailbox
			if err := db.DB.Select("domain").Where("username = ?", username).First(&box).Error; err != nil {
				continue
			}
			if !hasDomainAccess(claims, box.Domain) {
				continue
			}

			tx.Where("username = ?", username).Delete(&models.Mailbox{})
			tx.Where("address = ?", username).Delete(&models.Alias{})
			tx.Where("email = ?", username).Delete(&models.Vacation{})
		}

		tx.Commit()
		audit.Log(db.DB, claims.Username, "multiple", "batch delete mailboxes", fmt.Sprintf("%d items", len(req.Usernames)))
		return c.NoContent(http.StatusNoContent)
	})

	// Массовое изменение статуса
	mailboxes.POST("/batch/status", func(c echo.Context) error {
		type BatchStatusRequest struct {
			Usernames []string `json:"usernames" validate:"required,min=1"`
			Active    bool     `json:"active"`
		}
		var req BatchStatusRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}

		claims := c.Get("user").(*auth.Claims)
		tx := db.DB.Begin()

		for _, username := range req.Usernames {
			var box models.Mailbox
			if err := db.DB.Select("domain").Where("username = ?", username).First(&box).Error; err != nil {
				continue
			}
			if !hasDomainAccess(claims, box.Domain) {
				continue
			}
			tx.Model(&models.Mailbox{}).Where("username = ?", username).Update("active", req.Active)
		}

		tx.Commit()
		audit.Log(db.DB, claims.Username, "multiple", "batch status update", fmt.Sprintf("%d items to %v", len(req.Usernames), req.Active))
		return c.NoContent(http.StatusNoContent)
	})
}
