package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/user/mailadmin/internal/auth"
	"github.com/user/mailadmin/internal/config"
	"github.com/user/mailadmin/internal/db"
	"github.com/user/mailadmin/internal/models"
)

// maxSieveSize — максимальный размер rules_json в Байтах (MED-5)
const maxSieveSize = 65536

// isValidSieveUsername проверяет, что username является валидным email (MED-2)
func isValidSieveUsername(username string) bool {
	parts := strings.Split(username, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return false
	}
	// Имя домена должно содержать точку (example.com)
	return strings.Contains(parts[1], ".")
}

func RegisterSieveHandlers(g *echo.Group, secret string, cfg *config.Config) {
	g.Use(auth.JWTMiddleware(secret))

	// Получить правила Sieve для ящика или GLOBAL
	g.GET("/:username", func(c echo.Context) error {
		username := c.Param("username")
		claims := c.Get("user").(*auth.Claims)

		// Проверка доступа
		if username != "GLOBAL" {
			// MED-2: валидируем username как email до проверки домена
			if !isValidSieveUsername(username) {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid username"})
			}
			parts := strings.Split(username, "@")
			if !hasDomainAccess(claims, parts[1]) {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
			}
		} else {
			if !claims.SuperAdmin {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "only superadmins can manage global sieve"})
			}
		}

		var rule models.SieveRule
		if err := db.DB.Where("username = ?", username).First(&rule).Error; err != nil {
			// Если в базе нет, возвращаем пустую структуру
			return c.JSON(http.StatusOK, map[string]interface{}{
				"username":   username,
				"rules_json": "[]",
				"active":     true,
			})
		}

		return c.JSON(http.StatusOK, rule)
	})

	// Сохранить правила Sieve
	g.POST("/:username", func(c echo.Context) error {
		username := c.Param("username")
		claims := c.Get("user").(*auth.Claims)

		type SaveRequest struct {
			RulesJSON string `json:"rules_json"`
			Active    bool   `json:"active"`
		}
		var req SaveRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}

		// MED-5: ограничение размера rules_json для предотвращения записи мусора на диск
		if len(req.RulesJSON) > maxSieveSize {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "rules_json too large"})
		}

		// Проверка доступа (аналогично GET)
		if username != "GLOBAL" {
			// MED-2: валидируем username как email до формирования пути к файлу
			if !isValidSieveUsername(username) {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid username"})
			}
			parts := strings.Split(username, "@")
			if !hasDomainAccess(claims, parts[1]) {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
			}
		} else {
			if !claims.SuperAdmin {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "only superadmins can manage global sieve"})
			}
		}

		// Генерация кода Sieve из JSON
		sieveCode := generateSieveCode(req.RulesJSON)

		// Сохранение в БД
		var rule models.SieveRule
		res := db.DB.Where("username = ?", username).First(&rule)
		if res.Error != nil {
			rule = models.SieveRule{
				Username: username,
			}
		}
		rule.RulesJSON = req.RulesJSON
		rule.Content = sieveCode
		rule.Active = req.Active

		if err := db.DB.Save(&rule).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to save to database"})
		}

		// Запись на диск (если активно)
		if req.Active {
			filename := ""
			if username == "GLOBAL" {
				filename = "before.sieve"
			} else {
				filename = username + ".sieve"
			}

			// Безопасная сборка пути
			safePath := filepath.Join(cfg.SieveRoot, filename)
			
			// Проверка на выход за пределы папки (защита от path traversal)
			if !strings.HasPrefix(safePath, filepath.Clean(cfg.SieveRoot)) {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid path"})
			}

			if err := os.WriteFile(safePath, []byte(sieveCode), 0644); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to write file: " + err.Error()})
			}
		}

		return c.JSON(http.StatusOK, rule)
	})
}

// generateSieveCode превращает сложную JSON-структуру в работающий скрипт Sieve
func generateSieveCode(rulesJSON string) string {
	type Condition struct {
		Field    string `json:"field"`
		Operator string `json:"operator"` // contains, not_contains, is, not_is, matches
		Value    string `json:"value"`
	}
	type Action struct {
		Type   string `json:"type"`   // fileinto, redirect, discard, reject, setflag
		Target string `json:"target"` // Folder name, email, or flag
	}
	type Filter struct {
		Name       string      `json:"name"`
		MatchAll   bool        `json:"match_all"` // true = allof, false = anyof
		Conditions []Condition `json:"conditions"`
		Actions    []Action    `json:"actions"`
		Active     bool        `json:"active"`
	}

	var filters []Filter
	if err := json.Unmarshal([]byte(rulesJSON), &filters); err != nil {
		return "# Error parsing rules: " + err.Error()
	}

	if len(filters) == 0 {
		return "keep;"
	}

	code := "require [\"fileinto\", \"copy\", \"envelope\", \"reject\", \"imap4flags\", \"regex\", \"vacation\"];\n\n"

	for _, f := range filters {
		if !f.Active || len(f.Conditions) == 0 {
			continue
		}

		code += fmt.Sprintf("# Filter: %s\n", f.Name)
		
		logic := "anyof"
		if f.MatchAll {
			logic = "allof"
		}

		var condStrings []string
		for _, c := range f.Conditions {
			s := ""
			switch c.Field {
			case "Subject", "From", "To":
				op := ":contains"
				isNot := false
				switch c.Operator {
				case "not_contains": op = ":contains"; isNot = true
				case "is": op = ":is"
				case "not_is": op = ":is"; isNot = true
				case "matches": op = ":matches"
				case "regex": op = ":regex"
				}
				s = fmt.Sprintf("header %s \"%s\" \"%s\"", op, c.Field, c.Value)
				if isNot {
					s = "not " + s
				}
			case "X-Spam-Flag":
				s = "header :is \"X-Spam-Flag\" \"YES\""
			case "Body":
				s = fmt.Sprintf("body :contains \"%s\"", c.Value)
			}
			if s != "" {
				condStrings = append(condStrings, s)
			}
		}

		if len(condStrings) == 0 {
			continue
		}

		code += fmt.Sprintf("if %s (%s) {\n", logic, strings.Join(condStrings, ", "))
		
		for _, a := range f.Actions {
			switch a.Type {
			case "fileinto":
				code += fmt.Sprintf("    fileinto \"%s\";\n", a.Target)
			case "redirect":
				code += fmt.Sprintf("    redirect \"%s\";\n", a.Target)
			case "discard":
				code += "    discard;\n"
			case "reject":
				code += fmt.Sprintf("    reject \"%s\";\n", a.Target)
			case "setflag":
				code += fmt.Sprintf("    setflag \"%s\";\n", a.Target)
			case "vacation":
				code += fmt.Sprintf("    vacation :days 1 \"%s\";\n", a.Target)
			}
		}
		code += "    stop;\n}\n\n"
	}

	code += "keep;"
	return code
}
