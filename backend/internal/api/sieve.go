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

type Condition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"` // contains, not_contains, is, not_is, matches
	Value    string `json:"value"`
}

type Action struct {
	Type   string `json:"type"`   // fileinto, redirect, discard, reject, setflag, vacation
	Target string `json:"target"` // Folder name, email, or flag
}

type Filter struct {
	Name       string      `json:"name"`
	MatchAll   bool        `json:"match_all"` // true = allof, false = anyof
	Conditions []Condition `json:"conditions"`
	Actions    []Action    `json:"actions"`
	Active     bool        `json:"active"`
}


// isValidSieveUsername проверяет, что username является валидным email (MED-2)
func isValidSieveUsername(username string) bool {
	if strings.Contains(username, "/") || strings.Contains(username, "\\") || strings.Contains(username, "..") {
		return false
	}
	parts := strings.Split(username, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return false
	}
	// Имя домена должно содержать точку (example.com)
	return strings.Contains(parts[1], ".")
}

func resolveSievePath(username string, cfg *config.Config) (string, error) {
	var sieveSetting, sieveBeforeSetting string

	if cfg.DovecotConfigDir != "" {
		_ = filepath.Walk(cfg.DovecotConfigDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				return nil
			}
			if matched, _ := filepath.Match("*sieve.conf", info.Name()); matched {
				data, err := os.ReadFile(path)
				if err == nil {
					lines := strings.Split(string(data), "\n")
					for _, line := range lines {
						line = strings.TrimSpace(line)
						if strings.HasPrefix(line, "#") {
							continue
						}
						if strings.Contains(line, "=") {
							parts := strings.SplitN(line, "=", 2)
							key := strings.TrimSpace(parts[0])
							value := strings.TrimSpace(parts[1])
							
							if idx := strings.Index(value, "#"); idx != -1 {
								value = strings.TrimSpace(value[:idx])
							}
							value = strings.Trim(value, `"'`)
							
							if key == "sieve" {
								sieveSetting = value
							} else if key == "sieve_before" {
								sieveBeforeSetting = value
							}
						}
					}
				}
			}
			return nil
		})
	}

	if username == "GLOBAL" {
		if sieveBeforeSetting != "" {
			if strings.HasSuffix(sieveBeforeSetting, "/") {
				return filepath.Join(sieveBeforeSetting, "before.sieve"), nil
			}
			if info, err := os.Stat(sieveBeforeSetting); err == nil && info.IsDir() {
				return filepath.Join(sieveBeforeSetting, "before.sieve"), nil
			}
			return sieveBeforeSetting, nil
		}
		return filepath.Join(cfg.SieveRoot, "before.sieve"), nil
	}

	if sieveSetting != "" {
		domain := ""
		name := username
		parts := strings.Split(username, "@")
		if len(parts) == 2 {
			name = parts[0]
			domain = parts[1]
		}

		path := sieveSetting
		path = strings.ReplaceAll(path, "%d", domain)
		path = strings.ReplaceAll(path, "%n", name)
		path = strings.ReplaceAll(path, "%u", username)
		return path, nil
	}

	return filepath.Join(cfg.SieveRoot, username+".sieve"), nil
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
			// Безопасная сборка пути
			safePath, err := resolveSievePath(username, cfg)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to resolve sieve path"})
			}
			
			// Проверка на выход за пределы папки (защита от path traversal)
			cleanPath := filepath.Clean(safePath)
			if !strings.HasPrefix(cleanPath, filepath.Clean(cfg.SieveRoot)) && !strings.HasPrefix(cleanPath, filepath.Clean(cfg.MailRoot)) {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "path outside of allowed roots"})
			}

			if err := os.WriteFile(safePath, []byte(sieveCode), 0644); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to write file: " + err.Error()})
			}
		}

		return c.JSON(http.StatusOK, rule)
	})

	// Импорт правил Sieve с сервера
	g.POST("/:username/import", func(c echo.Context) error {
		username := c.Param("username")
		claims := c.Get("user").(*auth.Claims)

		// Проверка доступа
		if username != "GLOBAL" {
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

		safePath, err := resolveSievePath(username, cfg)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to resolve sieve path"})
		}

		cleanPath := filepath.Clean(safePath)
		if !strings.HasPrefix(cleanPath, filepath.Clean(cfg.SieveRoot)) && !strings.HasPrefix(cleanPath, filepath.Clean(cfg.MailRoot)) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "path outside of allowed roots"})
		}

		data, err := os.ReadFile(safePath)
		if err != nil {
			if os.IsNotExist(err) {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "file not found on server"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to read file"})
		}

		filters := parseSieveCode(string(data))
		rulesJSON, err := json.Marshal(filters)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to marshal rules"})
		}

		var rule models.SieveRule
		res := db.DB.Where("username = ?", username).First(&rule)
		if res.Error != nil {
			rule = models.SieveRule{
				Username: username,
			}
		}
		rule.RulesJSON = string(rulesJSON)
		rule.Content = string(data)
		rule.Active = true

		if err := db.DB.Save(&rule).Error; err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to save to database"})
		}

		return c.JSON(http.StatusOK, rule)
	})

}

// generateSieveCode превращает сложную JSON-структуру в работающий скрипт Sieve
func generateSieveCode(rulesJSON string) string {
	var filters []Filter
	if err := json.Unmarshal([]byte(rulesJSON), &filters); err != nil {
		return "# Error parsing rules: " + err.Error()
	}

	if len(filters) == 0 {
		return "keep;"
	}

	code := "require [\"fileinto\", \"copy\", \"envelope\", \"reject\", \"imap4flags\", \"regex\", \"vacation\", \"body\"];\n\n"


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

func parseSieveCode(code string) []Filter {
	var filters []Filter
	lines := strings.Split(code, "\n")

	var currentFilter *Filter
	var lastComment string

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#") {
			comment := strings.TrimSpace(strings.TrimPrefix(line, "#"))
			if strings.HasPrefix(comment, "Filter:") {
				comment = strings.TrimSpace(strings.TrimPrefix(comment, "Filter:"))
			} else if strings.HasPrefix(comment, "rule:[") && strings.HasSuffix(comment, "]") {
				comment = comment[6 : len(comment)-1]
			}
			lastComment = comment
			continue
		}

		lowerLine := strings.ToLower(line)
		if strings.HasPrefix(lowerLine, "if ") {
			if currentFilter != nil {
				filters = append(filters, *currentFilter)
			}

			name := lastComment
			if name == "" {
				name = fmt.Sprintf("Rule %d", len(filters)+1)
			}
			lastComment = ""

			currentFilter = &Filter{
				Name:     name,
				Active:   true,
				MatchAll: true,
			}

			if strings.Contains(lowerLine, "anyof") {
				currentFilter.MatchAll = false
			}

			var condsStr string
			startIdx := strings.Index(line, "(")
			endIdx := strings.LastIndex(line, ")")
			if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
				condsStr = line[startIdx+1 : endIdx]
			} else {
				idxIf := strings.Index(lowerLine, "if ")
				idxBrace := strings.Index(line, "{")
				if idxBrace != -1 {
					condsStr = line[idxIf+3 : idxBrace]
				} else {
					condsStr = line[idxIf+3:]
				}
			}

			condsStr = strings.TrimSpace(condsStr)
			condsStr = strings.TrimPrefix(condsStr, "allof")
			condsStr = strings.TrimPrefix(condsStr, "anyof")
			condsStr = strings.TrimSpace(condsStr)

			var condParts []string
			if strings.Contains(condsStr, ",") {
				condParts = strings.Split(condsStr, ",")
			} else {
				condParts = []string{condsStr}
			}

			for _, cp := range condParts {
				cp = strings.TrimSpace(cp)
				if cp == "" {
					continue
				}

				cond := Condition{}
				isNot := false
				if strings.HasPrefix(strings.ToLower(cp), "not ") {
					isNot = true
					cp = cp[4:]
					cp = strings.TrimSpace(cp)
				}

				lowerCp := strings.ToLower(cp)
				if strings.HasPrefix(lowerCp, "header ") {
					op := "contains"
					if strings.Contains(lowerCp, ":is") {
						op = "is"
					} else if strings.Contains(lowerCp, ":matches") {
						op = "matches"
					} else if strings.Contains(lowerCp, ":regex") {
						op = "regex"
					} else if strings.Contains(lowerCp, ":value") {
						// Для :value "ge" и т.д. часто используется i;ascii-numeric
						op = "matches" 
					}

					quotes := extractQuotes(cp)
					if len(quotes) >= 2 {
						// В Sieve параметры (:value, :comparator) идут в начале, 
						// а имя заголовка и значение — в конце.
						cond.Field = quotes[len(quotes)-2]
						cond.Value = quotes[len(quotes)-1]
						cond.Field = normalizeField(cond.Field)
						cond.Operator = op
						if isNot {
							if op == "contains" {
								cond.Operator = "not_contains"
							} else if op == "is" {
								cond.Operator = "not_is"
							}
						}
						currentFilter.Conditions = append(currentFilter.Conditions, cond)
					}
				} else if strings.HasPrefix(lowerCp, "body ") {
					quotes := extractQuotes(cp)
					if len(quotes) >= 1 {
						cond.Field = "Body"
						cond.Operator = "contains"
						cond.Value = quotes[0]
						currentFilter.Conditions = append(currentFilter.Conditions, cond)
					}
				}
			}
		} else {
			if currentFilter != nil {
				if line == "}" {
					filters = append(filters, *currentFilter)
					currentFilter = nil
					continue
				}

				cleanLine := strings.Trim(line, " {};\t")
				lowerClean := strings.ToLower(cleanLine)

				if strings.HasPrefix(lowerClean, "fileinto ") {
					quotes := extractQuotes(cleanLine)
					if len(quotes) >= 1 {
						currentFilter.Actions = append(currentFilter.Actions, Action{Type: "fileinto", Target: quotes[0]})
					}
				} else if strings.HasPrefix(lowerClean, "redirect ") {
					quotes := extractQuotes(cleanLine)
					if len(quotes) >= 1 {
						currentFilter.Actions = append(currentFilter.Actions, Action{Type: "redirect", Target: quotes[0]})
					}
				} else if lowerClean == "discard" {
					currentFilter.Actions = append(currentFilter.Actions, Action{Type: "discard"})
				} else if strings.HasPrefix(lowerClean, "reject ") {
					quotes := extractQuotes(cleanLine)
					if len(quotes) >= 1 {
						currentFilter.Actions = append(currentFilter.Actions, Action{Type: "reject", Target: quotes[0]})
					}
				} else if strings.HasPrefix(lowerClean, "setflag ") {
					quotes := extractQuotes(cleanLine)
					if len(quotes) >= 1 {
						currentFilter.Actions = append(currentFilter.Actions, Action{Type: "setflag", Target: quotes[0]})
					}
				} else if strings.HasPrefix(lowerClean, "vacation ") {
					quotes := extractQuotes(cleanLine)
					if len(quotes) >= 1 {
						currentFilter.Actions = append(currentFilter.Actions, Action{Type: "vacation", Target: quotes[0]})
					}
				}
			}
		}
	}

	if currentFilter != nil {
		filters = append(filters, *currentFilter)
	}

	return filters
}

func extractQuotes(s string) []string {
	var res []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' {
			if inQuotes {
				res = append(res, current.String())
				current.Reset()
				inQuotes = false
			} else {
				inQuotes = true
			}
		} else if inQuotes {
			if c == '\\' && i+1 < len(s) {
				// Обработка экранированных кавычек
				if s[i+1] == '"' || s[i+1] == '\\' {
					current.WriteByte(s[i+1])
					i++
					continue
				}
			}
			current.WriteByte(c)
		}
	}

	// Если кавычек не нашли совсем, пробуем взять просто слова (fallback)
	if len(res) == 0 {
		parts := strings.Fields(s)
		for _, p := range parts {
			p = strings.Trim(p, "\";,[]")
			if p != "" && !strings.HasPrefix(p, ":") && p != "header" && p != "body" && p != "if" && p != "anyof" && p != "allof" {
				res = append(res, p)
			}
		}
	}

	return res
}

func normalizeField(f string) string {
	fLower := strings.ToLower(f)
	switch fLower {
	case "subject":
		return "Subject"
	case "from":
		return "From"
	case "to":
		return "To"
	case "x-spam-flag", "x-spam-status":
		return "X-Spam-Flag"
	}
	return f
}


