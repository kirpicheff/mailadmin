package api

import (
	"gorm.io/gorm"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/user/mailadmin/internal/audit"
	"github.com/user/mailadmin/internal/auth"
	"github.com/user/mailadmin/internal/db"
	"github.com/user/mailadmin/internal/models"
)

// RegisterAdminHandlers регистрирует маршруты управления администраторами
func RegisterAdminHandlers(api *echo.Group, secret string) {
	adminGroup := api.Group("/admins")
	adminGroup.Use(auth.JWTMiddleware(secret))
	adminGroup.Use(auth.SuperAdminMiddleware)

	// Список всех админов
	adminGroup.GET("", func(c echo.Context) error {
		var admins []models.Admin
		if err := db.DB.Order("username").Find(&admins).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch admins"})
		}
		return c.JSON(http.StatusOK, admins)
	})

	// Детали одного админа + его домены
	adminGroup.GET("/:username", func(c echo.Context) error {
		username := c.Param("username")
		var admin models.Admin
		if err := db.DB.Where("username = ?", username).First(&admin).Error; err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "admin not found"})
		}

		var domains []string
		db.DB.Model(&models.DomainAdmin{}).Where("username = ?", username).Pluck("domain", &domains)

		return c.JSON(http.StatusOK, map[string]interface{}{
			"admin":   admin,
			"domains": domains,
		})
	})

	// Создание нового админа
	adminGroup.POST("", func(c echo.Context) error {
		type CreateRequest struct {
			models.Admin
			Password string   `json:"password"`
			Domains  []string `json:"domains"`
		}
		var req CreateRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}

		// Хешируем пароль со случайной солью (CRIT-2)
		if req.Password == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "password is required"})
		}
		hash, err := auth.GenerateHash(req.Password)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "hashing failed"})
		}

		req.Admin.Password = hash
		req.Admin.PasswordExpiry = time.Now().AddDate(-1, 0, 0) // Принудительная смена

		err = db.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&req.Admin).Error; err != nil {
				return err
			}
			for _, d := range req.Domains {
				if err := tx.Create(&models.DomainAdmin{Username: req.Admin.Username, Domain: d}).Error; err != nil {
					return err
				}
			}

			claims := c.Get("user").(*auth.Claims)
			audit.Log(tx, claims.Username, "system", "create admin", req.Admin.Username)

			return nil
		})

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create admin: " + err.Error()})
		}
		return c.JSON(http.StatusCreated, req.Admin)
	})

	// Редактирование админа
	adminGroup.PUT("/:username", func(c echo.Context) error {
		username := c.Param("username")
		type UpdateRequest struct {
			Password   string   `json:"password"`
			Active     bool     `json:"active"`
			SuperAdmin bool     `json:"superadmin"`
			Phone      string   `json:"phone"`
			EmailOther string   `json:"email_other"`
			Domains    []string `json:"domains"`
		}
		var req UpdateRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}

		updates := map[string]interface{}{
			"active":      req.Active,
			"superadmin":  req.SuperAdmin,
			"phone":       req.Phone,
			"email_other": req.EmailOther,
		}

		if req.Password != "" {
			// Хешируем с рандомной солью (CRIT-2)
			hash, err := auth.GenerateHash(req.Password)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "hashing failed"})
			}
			updates["password"] = hash
			updates["password_expiry"] = time.Now().AddDate(-1, 0, 0)
		}

		err := db.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&models.Admin{}).Where("username = ?", username).Updates(updates).Error; err != nil {
				return err
			}
			// Обновляем домены: удаляем старые, пишем новые
			if err := tx.Where("username = ?", username).Delete(&models.DomainAdmin{}).Error; err != nil {
				return err
			}
			for _, d := range req.Domains {
				if err := tx.Create(&models.DomainAdmin{Username: username, Domain: d}).Error; err != nil {
					return err
				}
			}

			claims := c.Get("user").(*auth.Claims)
			audit.Log(tx, claims.Username, "system", "update admin", username)

			return nil
		})

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update admin"})
		}
		return c.NoContent(http.StatusOK)
	})

	// Удаление админа
	adminGroup.DELETE("/:username", func(c echo.Context) error {
		username := c.Param("username")
		err := db.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("username = ?", username).Delete(&models.DomainAdmin{}).Error; err != nil {
				return err
			}
			if err := tx.Where("username = ?", username).Delete(&models.Admin{}).Error; err != nil {
				return err
			}

			claims := c.Get("user").(*auth.Claims)
			audit.Log(tx, claims.Username, "system", "delete admin", username)

			return nil
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete admin"})
		}
		return c.NoContent(http.StatusNoContent)
	})
}
