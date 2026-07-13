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

		var setting models.Setting
		receivePasswords := true
		if err := db.DB.Where("setting_key = ?", "disable_pass_notif:"+username).First(&setting).Error; err == nil {
			if setting.Value == "true" {
				receivePasswords = false
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"admin":             admin,
			"domains":           domains,
			"receive_passwords": receivePasswords,
		})
	})

	// Создание нового админа
	adminGroup.POST("", func(c echo.Context) error {
		type CreateRequest struct {
			models.Admin
			Username         string   `json:"username" validate:"required,email"`
			Password         string   `json:"password" validate:"required,min=8"`
			Domains          []string `json:"domains"`
			ReceivePasswords bool     `json:"receive_passwords"`
		}
		var req CreateRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}
		if err := c.Validate(&req); err != nil {
			return err
		}

		// Хешируем пароль со случайной солью (CRIT-2)
		hash, err := auth.GenerateHash(req.Password)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "hashing failed"})
		}

		req.Admin.Username = req.Username
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

			// Если суперадмин не хочет получать пароли, сохраняем настройку
			if !req.ReceivePasswords {
				setting := models.Setting{
					Key:   "disable_pass_notif:" + req.Admin.Username,
					Value: "true",
				}
				if err := tx.Save(&setting).Error; err != nil {
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
			Password         string   `json:"password" validate:"omitempty,min=8"`
			Active           bool     `json:"active"`
			SuperAdmin       bool     `json:"superadmin"`
			Phone            string   `json:"phone"`
			EmailOther       string   `json:"email_other" validate:"omitempty,email"`
			Domains          []string `json:"domains"`
			ReceivePasswords bool     `json:"receive_passwords"`
		}
		var req UpdateRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}
		if err := c.Validate(&req); err != nil {
			return err
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

			// Обновляем опцию рассылки паролей
			if !req.ReceivePasswords {
				setting := models.Setting{
					Key:   "disable_pass_notif:" + username,
					Value: "true",
				}
				if err := tx.Save(&setting).Error; err != nil {
					return err
				}
			} else {
				if err := tx.Where("setting_key = ?", "disable_pass_notif:"+username).Delete(&models.Setting{}).Error; err != nil {
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
			// Удаляем также настройку уведомлений
			if err := tx.Where("setting_key = ?", "disable_pass_notif:"+username).Delete(&models.Setting{}).Error; err != nil {
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
