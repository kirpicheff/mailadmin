package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/GehirnInc/crypt"
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
		var box models.Mailbox
		if err := c.Bind(&box); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}

		// Валидация username (должен быть email)
		parts := strings.Split(box.Username, "@")
		if len(parts) != 2 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "username must be a full email address"})
		}
		box.LocalPart = parts[0]
		box.Domain = parts[1]

		// Установка maildir: domain/user/
		box.Maildir = fmt.Sprintf("%s/%s/", box.Domain, box.LocalPart)

		// Хеширование пароля
		if box.Password != "" {
			cryptService := crypt.New(crypt.SHA512)
			hash, err := cryptService.Generate([]byte(box.Password), []byte("$6$salt"))
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "hashing failed"})
			}
			box.Password = hash
		}

		box.Created = time.Now()
		box.Modified = time.Now()

		if err := db.DB.Create(&box).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create mailbox: " + err.Error()})
		}

		claims := c.Get("user").(*auth.Claims)

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

		var update models.Mailbox
		if err := c.Bind(&update); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}

		// Если пароль передан - хешируем
		if update.Password != "" && !strings.HasPrefix(update.Password, "$6$") {
			cryptService := crypt.New(crypt.SHA512)
			hash, err := cryptService.Generate([]byte(update.Password), []byte("$6$salt"))
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "hashing failed"})
			}
			update.Password = hash
		} else {
			update.Password = existing.Password
		}

		update.Modified = time.Now()

		// Используем Updates с map или структурой, чтобы избежать затирания полей
		if err := db.DB.Model(&existing).Updates(map[string]interface{}{
			"name":            update.Name,
			"password":        update.Password,
			"quota":           update.Quota,
			"active":          update.Active,
			"phone":           update.Phone,
			"email_other":     update.EmailOther,
			"password_expiry": update.PasswordExpiry,
			"modified":        update.Modified,
		}).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update mailbox"})
		}

		claims := c.Get("user").(*auth.Claims)
		audit.Log(db.DB, claims.Username, existing.Domain, "update mailbox", existing.Username)

		return c.JSON(http.StatusOK, existing)
	})

	// Удаление ящика
	mailboxes.DELETE("/:username", func(c echo.Context) error {
		username := c.Param("username")
		// Получаем домен для лога перед удалением
		var box models.Mailbox
		db.DB.Select("domain").Where("username = ?", username).First(&box)

		if err := db.DB.Where("username = ?", username).Delete(&models.Mailbox{}).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete mailbox"})
		}

		claims := c.Get("user").(*auth.Claims)
		audit.Log(db.DB, claims.Username, box.Domain, "delete mailbox", username)

		return c.NoContent(http.StatusNoContent)
	})
}
