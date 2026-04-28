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

// QueueItem структура для почтовой очереди
type QueueItem struct {
	ID        string   `json:"id"`
	Size      int      `json:"size"`
	Arrival   string   `json:"arrival"`
	Sender    string   `json:"sender"`
	Recipient []string `json:"recipients"`
	Reason    string   `json:"reason"`
}

func getFullQueue() []QueueItem {
	out := runCmd("postqueue", "-p")
	if out == "" || strings.Contains(out, "Mail queue is empty") {
		return []QueueItem{}
	}

	var items []QueueItem
	lines := strings.Split(out, "\n")

	// Регулярка для заголовка письма в очереди
	// Пример: 4B87F40CCF*     493 Mon Oct 14 10:14:44  sender@example.com
	reHeader := regexp.MustCompile(`^([0-9A-F]+)([\*!]?)\s+(\d+)\s+(\w{3}\s+\w{3}\s+\d+\s+\d+:\d+:\d+)\s+(.+)$`)

	var current *QueueItem

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "-") || strings.HasPrefix(line, "(") {
			// Если есть причина ошибки (обычно в скобках на следующей строке)
			if strings.HasPrefix(line, "(") && current != nil {
				current.Reason = strings.Trim(line, "()")
			}
			continue
		}

		match := reHeader.FindStringSubmatch(line)
		if len(match) > 0 {
			if current != nil {
				items = append(items, *current)
			}
			size, _ := strconv.Atoi(match[3])
			current = &QueueItem{
				ID:      match[1],
				Size:    size,
				Arrival: match[4],
				Sender:  match[5],
			}
		} else if current != nil {
			// Это получатель (обычно идет после заголовка)
			current.Recipient = append(current.Recipient, line)
		}
	}

	if current != nil {
		items = append(items, *current)
	}

	return items
}

// RegisterSystemHandlers регистрирует маршруты системного мониторинга
func RegisterSystemHandlers(g *echo.Group, secret string) {
	system := g.Group("/system")
	system.Use(auth.JWTMiddleware(secret))

	// Проверка на суперадмина
	system.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims := c.Get("user").(*auth.Claims)
			if !claims.SuperAdmin {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "superadmin access required"})
			}
			return next(c)
		}
	})

	system.GET("/health", func(c echo.Context) error {
		stats := getSystemStats()
		return c.JSON(http.StatusOK, stats)
	})

	// Почтовая очередь
	system.GET("/queue", func(c echo.Context) error {
		queue := getFullQueue()
		return c.JSON(http.StatusOK, queue)
	})

	system.POST("/queue/flush", func(c echo.Context) error {
		runCmd("postqueue", "-f")
		return c.NoContent(http.StatusNoContent)
	})

	system.DELETE("/queue/:id", func(c echo.Context) error {
		id := c.Param("id")
		if id == "all" {
			runCmd("postsuper", "-d", "ALL")
		} else {
			runCmd("postsuper", "-d", id)
		}
		return c.NoContent(http.StatusNoContent)
	})

	// Логи сервера
	system.GET("/logs", func(c echo.Context) error {
		lines, _ := strconv.Atoi(c.QueryParam("lines"))
		if lines <= 0 || lines > 5000 {
			lines = 200
		}
		search := c.QueryParam("search")

		// Пытаемся найти лог почты
		logPaths := []string{"/var/log/mail.log", "/var/log/maillog"}
		var content string
		var logFile string
		for _, p := range logPaths {
			if _, err := os.Stat(p); err == nil {
				logFile = p
				break
			}
		}

		if logFile != "" {
			if search != "" {
				// Используем grep для поиска (последние N совпадений)
				// sh -c "grep -i 'pattern' file | tail -n lines"
				shCmd := fmt.Sprintf("grep -i %s %s | tail -n %d", strconv.Quote(search), logFile, lines)
				content = runCmd("sh", "-c", shCmd)
			} else {
				content = runCmd("tail", "-n", strconv.Itoa(lines), logFile)
			}
		}

		// Если файлов нет или пусто (и нет поиска), пробуем journalctl
		if content == "" && logFile == "" {
			if search != "" {
				shCmd := fmt.Sprintf("journalctl -u postfix --no-pager | grep -i %s | tail -n %d", strconv.Quote(search), lines)
				content = runCmd("sh", "-c", shCmd)
			} else {
				content = runCmd("journalctl", "-u", "postfix", "-n", strconv.Itoa(lines), "--no-pager")
			}
		}

		return c.JSON(http.StatusOK, map[string]string{"logs": content})
	})

	// Fail2Ban управление
	system.GET("/fail2ban", func(c echo.Context) error {
		out := runCmd("fail2ban-client", "status")
		reJails := regexp.MustCompile(`Jail list:\s+(.+)`)
		match := reJails.FindStringSubmatch(out)
		if len(match) < 2 {
			return c.JSON(http.StatusOK, []interface{}{})
		}

		jails := strings.Split(match[1], ",")
		type BannedIP struct {
			IP   string `json:"ip"`
			Jail string `json:"jail"`
		}
		var result []BannedIP

		for _, jail := range jails {
			jail = strings.Trim(strings.TrimSpace(jail), ",")
			if jail == "" {
				continue
			}
			jOut := runCmd("fail2ban-client", "status", jail)
			reIPs := regexp.MustCompile(`Banned IP list:\s+(.+)`)
			iMatch := reIPs.FindStringSubmatch(jOut)
			if len(iMatch) >= 2 {
				ips := strings.Fields(iMatch[1])
				for _, ip := range ips {
					result = append(result, BannedIP{IP: ip, Jail: jail})
				}
			}
		}
		return c.JSON(http.StatusOK, result)
	})

	system.DELETE("/fail2ban/unban", func(c echo.Context) error {
		ip := c.QueryParam("ip")
		jail := c.QueryParam("jail")
		if ip == "" || jail == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "ip and jail required"})
		}
		runCmd("fail2ban-client", "set", jail, "unbanip", ip)
		return c.NoContent(http.StatusNoContent)
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

	// 8. Fail2Ban (Все тюрьмы)
	f2bStatus := runCmd("fail2ban-client", "status")
	reJails := regexp.MustCompile(`Jail list:\s+(.+)`)
	mJails := reJails.FindStringSubmatch(f2bStatus)
	if len(mJails) >= 2 {
		jails := strings.Split(mJails[1], ",")
		totalBanned := 0
		for _, j := range jails {
			j = strings.Trim(strings.TrimSpace(j), ",")
			if j == "" { continue }
			jOut := runCmd("fail2ban-client", "status", j)
			reBanned := regexp.MustCompile(`Currently banned:\s+(\d+)`)
			mBanned := reBanned.FindStringSubmatch(jOut)
			if len(mBanned) >= 2 {
				count, _ := strconv.Atoi(mBanned[1])
				totalBanned += count
			}
		}
		s.F2BCount = totalBanned
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
