package api

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/user/mailadmin/internal/auth"
	"github.com/user/mailadmin/internal/db"
	"github.com/user/mailadmin/internal/models"
)

func init() {
	// Инициализация валидатора API токенов
	auth.APITokenValidator = func(tokenString string) (*auth.Claims, error) {
		var token models.APIToken
		if err := db.DB.Where("token = ? AND active = ?", tokenString, true).First(&token).Error; err != nil {
			return nil, fmt.Errorf("invalid or inactive api token")
		}

		if token.ExpiresAt != nil && token.ExpiresAt.Before(time.Now()) {
			return nil, fmt.Errorf("api token expired")
		}

		// Если токен найден и активен, возвращаем Claims.
		// Для API токенов SuperAdmin = true, но ограничиваем права с помощью APIScopes.
		return &auth.Claims{
			Username:   "API_TOKEN",
			SuperAdmin: true,
			IsAPIToken: true,
			APIScopes:  []string{token.Scope},
		}, nil
	}
}

// RegisterTokenHandlers регистрирует маршруты для управления API токенами
func RegisterTokenHandlers(g *echo.Group, secret string) {
	tokensGroup := g.Group("/system/api-tokens")
	tokensGroup.Use(auth.JWTMiddleware(secret))
	tokensGroup.Use(auth.SuperAdminMiddleware)

	// Список всех токенов
	tokensGroup.GET("", func(c echo.Context) error {
		var tokens []models.APIToken
		if err := db.DB.Order("id desc").Find(&tokens).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch tokens: " + err.Error()})
		}
		// Никогда не отдаем сами токены целиком при листинге в целях безопасности
		// В идеале мы отдаем только их часть или вообще не отдаем.
		for i := range tokens {
			if len(tokens[i].Token) > 10 {
				tokens[i].Token = tokens[i].Token[:10] + "..."
			}
		}
		return c.JSON(http.StatusOK, tokens)
	})

	// Создание токена
	tokensGroup.POST("", func(c echo.Context) error {
		type CreateRequest struct {
			Description string `json:"description" validate:"required"`
			Scope       string `json:"scope" validate:"required"`
		}

		var req CreateRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}

		// Генерация безопасного токена
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
		}
		tokenString := "ma_tk_" + hex.EncodeToString(b)

		token := models.APIToken{
			Token:       tokenString,
			Description: req.Description,
			Scope:       req.Scope,
			CreatedAt:   time.Now(),
			Active:      true,
		}

		if err := db.DB.Create(&token).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to save token"})
		}

		// Отдаем полный токен только при создании
		return c.JSON(http.StatusOK, token)
	})

	// Удаление токена (или отключение)
	tokensGroup.DELETE("/:id", func(c echo.Context) error {
		id := c.Param("id")
		if err := db.DB.Where("id = ?", id).Delete(&models.APIToken{}).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete token"})
		}
		return c.JSON(http.StatusOK, map[string]string{"message": "token deleted"})
	})
}
