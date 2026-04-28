package api

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/user/mailadmin/internal/auth"
)

// SystemStats структура для ответа
type SystemStats struct {
	Uptime       string          `json:"uptime"`
	Hostname     string          `json:"hostname"`
	RAMTotal     int             `json:"ram_total"`
	RAMUsed      int             `json:"ram_used"`
	RAMPerc      int             `json:"ram_perc"`
	DiskTotal    string          `json:"disk_total"`
	DiskUsed     string          `json:"disk_used"`
	DiskPerc     int             `json:"disk_perc"`
	LoadAvg      string          `json:"load"`
	MailQueue    int             `json:"queue"`
	IMAPSessions int             `json:"imap_sessions"`
	DBThreads    int             `json:"db_threads"`
	RedisMemory  string          `json:"redis_memory"`
	F2BCount     int             `json:"f2b_count"`
	SSLRemaining int             `json:"ssl_days"`
	Services     []ServiceStatus `json:"services"`
}

type ServiceStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Info   string `json:"info"`
}

// RegisterSystemHandlers регистрирует маршруты системного мониторинга
func RegisterSystemHandlers(g *echo.Group, secret string) {
	system := g.Group("/system")
	system.Use(auth.JWTMiddleware(secret))

	system.GET("/health", func(c echo.Context) error {
		claims := c.Get("user").(*auth.Claims)
		if !claims.SuperAdmin {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied: superadmin only"})
		}

		stats := getSystemStats()
		return c.JSON(http.StatusOK, stats)
	})
}

func runCmd(name string, arg ...string) string {
	cmd := exec.Command(name, arg...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out.String())
}

func getSystemStats() SystemStats {
	var s SystemStats
	s.Hostname, _ = os.Hostname()

	// 1. RAM (free -m)
	ramOut := runCmd("free", "-m")
	if ramOut != "" {
		lines := strings.Split(ramOut, "\n")
		if len(lines) > 1 {
			fields := strings.Fields(lines[1])
			if len(fields) > 2 {
				s.RAMTotal, _ = strconv.Atoi(fields[1])
				s.RAMUsed, _ = strconv.Atoi(fields[2])
				if s.RAMTotal > 0 {
					s.RAMPerc = int(float64(s.RAMUsed) / float64(s.RAMTotal) * 100)
				}
			}
		}
	}

	// 2. Disk (df /data)
	// Пытаемся /data, если нет - корень
	diskPath := "/data"
	if _, err := os.Stat(diskPath); os.IsNotExist(err) {
		diskPath = "/"
	}
	diskOut := runCmd("df", "-h", diskPath)
	if diskOut != "" {
		lines := strings.Split(diskOut, "\n")
		if len(lines) > 1 {
			fields := strings.Fields(lines[1])
			if len(fields) > 4 {
				s.DiskTotal = fields[1]
				s.DiskUsed = fields[2]
				percStr := strings.TrimSuffix(fields[4], "%")
				s.DiskPerc, _ = strconv.Atoi(percStr)
			}
		}
	}

	// 3. Load Avg (uptime)
	uptimeOut := runCmd("uptime")
	if uptimeOut != "" {
		parts := strings.Split(uptimeOut, "load average:")
		if len(parts) > 1 {
			s.LoadAvg = strings.TrimSpace(strings.Split(parts[1], ",")[0])
		}
		// Упрощенный парсинг аптайма
		upParts := strings.Split(uptimeOut, "up")
		if len(upParts) > 1 {
			s.Uptime = strings.TrimSpace(strings.Split(upParts[1], ",")[0])
		}
	}

	// 4. Mail Queue (postqueue -p)
	queueOut := runCmd("postqueue", "-p")
	if queueOut != "" {
		lines := strings.Split(queueOut, "\n")
		lastLine := lines[len(lines)-1]
		if strings.Contains(lastLine, "Mail queue is empty") {
			s.MailQueue = 0
		} else {
			re := regexp.MustCompile(`(\d+) Requests.`)
			match := re.FindStringSubmatch(lastLine)
			if len(match) > 1 {
				s.MailQueue, _ = strconv.Atoi(match[1])
			}
		}
	}

	// 5. IMAP Sessions (doveadm who)
	imapOut := runCmd("doveadm", "who")
	if imapOut != "" {
		lines := strings.Split(imapOut, "\n")
		if len(lines) > 1 {
			s.IMAPSessions = len(lines) - 1 // Пропускаем заголовок
		}
	}

	// 6. DB Threads (mysqladmin status)
	dbOut := runCmd("mysqladmin", "status")
	if dbOut != "" {
		re := regexp.MustCompile(`Threads: (\d+)`)
		match := re.FindStringSubmatch(dbOut)
		if len(match) > 1 {
			s.DBThreads, _ = strconv.Atoi(match[1])
		}
	}

	// 7. Redis Memory
	redisOut := runCmd("redis-cli", "info", "memory")
	if redisOut != "" {
		re := regexp.MustCompile(`used_memory_human:([\d\w.]+)`)
		match := re.FindStringSubmatch(redisOut)
		if len(match) > 1 {
			s.RedisMemory = match[1]
		}
	}

	// 8. Fail2Ban
	f2bOut := runCmd("fail2ban-client", "status", "postfix")
	if f2bOut != "" {
		re := regexp.MustCompile(`Currently banned:\s+(\d+)`)
		match := re.FindStringSubmatch(f2bOut)
		if len(match) > 1 {
			s.F2BCount, _ = strconv.Atoi(match[1])
		}
	}

	// 9. Supervisor Services
	superOut := runCmd("supervisorctl", "status")
	if superOut != "" {
		scanner := bufio.NewScanner(strings.NewReader(superOut))
		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				s.Services = append(s.Services, ServiceStatus{
					Name:   fields[0],
					Status: fields[1],
					Info:   strings.Join(fields[2:], " "),
				})
			}
		}
	}

	// 10. SSL Remaining (Упрощенно смотрим первый попавшийся сертификат)
	// В реальной системе пути могут отличаться, берем из примера
	certPath := fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem", strings.ToLower(s.Hostname))
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		certPath = "/etc/nginx/ssl/mailserver.crt"
	}

	if _, err := os.Stat(certPath); err == nil {
		sslOut := runCmd("openssl", "x509", "-enddate", "-noout", "-in", certPath)
		// notAfter=Oct 15 12:13:44 2026 GMT
		if strings.Contains(sslOut, "notAfter=") {
			dateStr := strings.TrimPrefix(sslOut, "notAfter=")
			expiry, err := time.Parse("Jan _2 15:04:05 2006 MST", dateStr)
			if err == nil {
				days := int(time.Until(expiry).Hours() / 24)
				s.SSLRemaining = days
			}
		}
	}

	return s
}
