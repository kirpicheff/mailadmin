package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/user/mailadmin/internal/auth"
	"github.com/user/mailadmin/internal/config"
	"github.com/user/mailadmin/internal/db"
	"github.com/user/mailadmin/internal/models"
)

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	User        struct {
		Username   string `json:"username"`
		SuperAdmin bool   `json:"superadmin"`
	} `json:"user"`
}

// RegisterAuthHandlers регистрирует маршруты аутентификации
func RegisterAuthHandlers(e *echo.Group, cfg *config.Config) {
	e.POST("/login", func(c echo.Context) error {
		req := new(LoginRequest)
		if err := c.Bind(req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		var admin models.Admin
		username := strings.TrimSpace(req.Username)
		if err := db.DB.Where("username = ? AND active = ?", username, true).First(&admin).Error; err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		}

		// Проверка пароля
		valid, err := auth.CheckPassword(req.Password, admin.Password)
		if err != nil || !valid {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		}

		// Проверка срока пароля
		mustChange := false
		if !admin.PasswordExpiry.IsZero() && admin.PasswordExpiry.Before(time.Now()) {
			mustChange = true
		}

		// Генерация токенов
		accessToken, err := auth.GenerateAccessToken(admin.Username, admin.SuperAdmin, mustChange, cfg.JWTSecret)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not generate access token"})
		}

		refreshToken, err := auth.GenerateRefreshToken(admin.Username, cfg.JWTSecret)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not generate refresh token"})
		}

		// Сохраняем refresh token в отдельную таблицу сессий
		session := models.Session{
			Username:     admin.Username,
			RefreshToken: refreshToken,
			UserAgent:    c.Request().UserAgent(),
			IP:           c.RealIP(),
			ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
		}
		db.DB.Create(&session)

		// Установка Refresh Token в HttpOnly cookie
		cookie := new(http.Cookie)
		cookie.Name = "refreshToken"
		cookie.Value = refreshToken
		cookie.Expires = session.ExpiresAt
		cookie.HttpOnly = true
		cookie.Secure = true                      // Только по HTTPS
		cookie.SameSite = http.SameSiteStrictMode // Защита от CSRF
		cookie.Path = "/api/auth/refresh"         // Только для эндпоинта обновления
		c.SetCookie(cookie)

		resp := TokenResponse{
			AccessToken: accessToken,
		}
		resp.User.Username = admin.Username
		resp.User.SuperAdmin = admin.SuperAdmin

		return c.JSON(http.StatusOK, resp)
	})

	e.POST("/refresh", func(c echo.Context) error {
		cookie, err := c.Cookie("refreshToken")
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "no refresh token"})
		}

		_, err = auth.ValidateToken(cookie.Value, cfg.JWTSecret)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
		}

		// Проверка в таблице сессий
		var session models.Session
		if err := db.DB.Where("refresh_token = ?", cookie.Value).First(&session).Error; err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "session expired or revoked"})
		}

		// Ротация: удаляем старую сессию
		db.DB.Delete(&session)

		// Проверяем, существует ли еще пользователь и активен ли он
		var admin models.Admin
		if err := db.DB.Where("username = ? AND active = ?", session.Username, true).First(&admin).Error; err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "user deactivated"})
		}

		// Проверка срока пароля
		mustChange := false
		if !admin.PasswordExpiry.IsZero() && admin.PasswordExpiry.Before(time.Now()) {
			mustChange = true
		}

		// Генерация нового Access Token и нового Refresh Token
		newAccessToken, err := auth.GenerateAccessToken(admin.Username, admin.SuperAdmin, mustChange, cfg.JWTSecret)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not generate access token"})
		}

		newRefreshToken, err := auth.GenerateRefreshToken(admin.Username, cfg.JWTSecret)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not generate refresh token"})
		}

		// Сохраняем новую сессию
		newSession := models.Session{
			Username:     admin.Username,
			RefreshToken: newRefreshToken,
			UserAgent:    c.Request().UserAgent(),
			IP:           c.RealIP(),
			ExpiresAt:    time.Now().Add(7 * 24 * time.Hour),
		}
		db.DB.Create(&newSession)

		// Обновляем куку
		newCookie := new(http.Cookie)
		newCookie.Name = "refreshToken"
		newCookie.Value = newRefreshToken
		newCookie.Expires = newSession.ExpiresAt
		newCookie.HttpOnly = true
		newCookie.Secure = true
		newCookie.SameSite = http.SameSiteStrictMode
		newCookie.Path = "/api/auth/refresh"
		c.SetCookie(newCookie)

		return c.JSON(http.StatusOK, map[string]string{"access_token": newAccessToken})
	})

	e.POST("/logout", func(c echo.Context) error {
		// Удаляем сессию из базы
		cookieToken, err := c.Cookie("refreshToken")
		if err == nil && cookieToken.Value != "" {
			db.DB.Where("refresh_token = ?", cookieToken.Value).Delete(&models.Session{})
		}

		// Очистка куки (те же флаги, что при установке)
		cookie := new(http.Cookie)
		cookie.Name = "refreshToken"
		cookie.Value = ""
		cookie.Expires = time.Now().Add(-1 * time.Hour)
		cookie.HttpOnly = true
		cookie.Secure = true
		cookie.SameSite = http.SameSiteStrictMode
		cookie.Path = "/api/auth/refresh"
		c.SetCookie(cookie)

		return c.NoContent(http.StatusOK)
	})

	e.GET("/me", func(c echo.Context) error {
		// Middleware заполнит контекст данными пользователя
		user := c.Get("user").(*auth.Claims)
		return c.JSON(http.StatusOK, user)
	}, auth.JWTMiddleware(cfg.JWTSecret))

	e.POST("/change-password", func(c echo.Context) error {
		type ChangeRequest struct {
			NewPassword string `json:"new_password" validate:"required,min=8"`
		}
		req := new(ChangeRequest)
		if err := c.Bind(req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}
		if err := c.Validate(req); err != nil {
			return err
		}

		user := c.Get("user").(*auth.Claims)

		// Хешируем новый пароль со случайной солью (CRIT-2)
		hash, err := auth.GenerateHash(req.NewPassword)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not hash password"})
		}

		// Обновляем в базе и ставим срок годности на 1 год вперёд
		err = db.DB.Model(&models.Admin{}).Where("username = ?", user.Username).Updates(map[string]interface{}{
			"password":        hash,
			"password_expiry": time.Now().AddDate(1, 0, 0), // +1 год
		}).Error

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update database"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "password changed successfully"})
	}, auth.JWTMiddleware(cfg.JWTSecret))
}
