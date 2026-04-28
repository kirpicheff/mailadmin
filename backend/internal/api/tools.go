package api

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/user/mailadmin/internal/audit"
	"github.com/user/mailadmin/internal/auth"
	"github.com/user/mailadmin/internal/db"
	"github.com/user/mailadmin/internal/mail"
	"github.com/user/mailadmin/internal/models"
)

// RegisterToolsHandlers регистрирует вспомогательные инструменты
func RegisterToolsHandlers(g *echo.Group, secret string) {
	tools := g.Group("/tools")
	tools.Use(auth.JWTMiddleware(secret))

	// Проверка MX записи для домена
	tools.GET("/check-mx/:domain", func(c echo.Context) error {
		domain := c.Param("domain")
		mxs, err := net.LookupMX(domain)
		if err != nil {
			return c.JSON(http.StatusOK, map[string]interface{}{"valid": false, "error": err.Error()})
		}

		var records []string
		for _, mx := range mxs {
			records = append(records, fmt.Sprintf("%s (%d)", mx.Host, mx.Pref))
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"valid":   len(mxs) > 0,
			"records": records,
		})
	})

	// Отправка одиночного письма
	tools.POST("/send-email", func(c echo.Context) error {
		type Request struct {
			From    string `json:"from" validate:"required,email"`
			To      string `json:"to" validate:"required,email"`
			Subject string `json:"subject" validate:"required"`
			Body    string `json:"body" validate:"required"`
		}
		var req Request
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}
		if err := c.Validate(&req); err != nil {
			return err
		}

		claims := c.Get("user").(*auth.Claims)
		var fromDomain string
		for i := len(req.From) - 1; i >= 0; i-- {
			if req.From[i] == '@' {
				fromDomain = req.From[i+1:]
				break
			}
		}
		if fromDomain == "" || !hasDomainAccess(claims, fromDomain) {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied to send from this domain"})
		}

		encodedBody := base64.StdEncoding.EncodeToString([]byte(req.Body))

		msg := &mail.EmailMessage{
			From:    req.From,
			To:      []string{req.To},
			Subject: req.Subject,
			Body:    encodedBody,
		}

		if err := mail.SendEmail(msg); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to send email: " + err.Error()})
		}

		audit.Log(db.DB, claims.Username, "*", "send email", fmt.Sprintf("To: %s, Subj: %s", req.To, req.Subject))

		return c.JSON(http.StatusOK, map[string]string{"message": "Email sent successfully"})
	})

	// Широковещательная рассылка
	tools.POST("/broadcast", func(c echo.Context) error {
		type Request struct {
			From          string   `json:"from" validate:"required,email"`
			Domains       []string `json:"domains"`
			Subject       string   `json:"subject" validate:"required"`
			Body          string   `json:"body" validate:"required"`
			OnlyMailboxes bool     `json:"only_mailboxes"`
		}
		var req Request
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}
		if err := c.Validate(&req); err != nil {
			return err
		}

		claims := c.Get("user").(*auth.Claims)

		if !claims.SuperAdmin {
			if len(req.Domains) == 0 {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "you must specify domains to broadcast to"})
			}
			for _, d := range req.Domains {
				if !hasDomainAccess(claims, d) {
					return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied to one or more domains"})
				}
			}
		}

		// Собираем список получателей асинхронно
		go func() {
			var recipients []string

			query := db.DB.Model(&models.Mailbox{})
			if len(req.Domains) > 0 {
				query = query.Where("domain IN (?)", req.Domains)
			}
			query.Pluck("username", &recipients)

			// Если нужны и алиасы
			if !req.OnlyMailboxes {
				var aliasRecipients []string
				aliasQuery := db.DB.Model(&models.Alias{})
				if len(req.Domains) > 0 {
					aliasQuery = aliasQuery.Where("domain IN (?)", req.Domains)
				}
				aliasQuery.Pluck("address", &aliasRecipients)
				recipients = append(recipients, aliasRecipients...)
			}

			// Уникализация
			uniqueRecipients := make(map[string]bool)
			for _, r := range recipients {
				uniqueRecipients[r] = true
			}

			encodedBody := base64.StdEncoding.EncodeToString([]byte(req.Body))

			// Рассылка в параллельных потоках (пул из 5 воркеров, чтобы не спамить слишком быстро)
			jobs := make(chan string, len(uniqueRecipients))
			var wg sync.WaitGroup

			for w := 1; w <= 5; w++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for addr := range jobs {
						msg := &mail.EmailMessage{
							From:    req.From,
							To:      []string{addr},
							Subject: req.Subject,
							Body:    encodedBody,
						}
						_ = mail.SendEmail(msg)
						time.Sleep(100 * time.Millisecond) // Небольшая пауза
					}
				}()
			}

			for addr := range uniqueRecipients {
				jobs <- addr
			}
			close(jobs)
			wg.Wait()

			audit.Log(db.DB, claims.Username, "*", "broadcast", fmt.Sprintf("Domains: %v, Recipients: %d", req.Domains, len(uniqueRecipients)))
		}()

		return c.JSON(http.StatusAccepted, map[string]string{"message": "Broadcast started in background"})
	})
}
