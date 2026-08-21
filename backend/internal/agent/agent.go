package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
)

const SocketPath = "/var/run/mailadmin/agent.sock"
const LogPath = "/var/log/mailadmin-agent.log"

type ActionType string

const (
	ActionFail2banUnban        ActionType = "fail2ban_unban"
	ActionFail2banBan          ActionType = "fail2ban_ban"
	ActionFail2banStatus       ActionType = "fail2ban_status"
	ActionFail2banWhitelistAdd    ActionType = "fail2ban_whitelist_add"
	ActionFail2banWhitelistList   ActionType = "fail2ban_whitelist_list"
	ActionFail2banWhitelistDelete ActionType = "fail2ban_whitelist_delete"
	ActionQueueDelete             ActionType = "queue_delete"
	ActionQueueFlush           ActionType = "queue_flush"
	ActionQueueStatus          ActionType = "queue_status"
	ActionImapStatus           ActionType = "imap_status"
	ActionServiceStatus        ActionType = "service_status"
	ActionPing                 ActionType = "ping"
)

type AgentRequest struct {
	Action  ActionType      `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

// isAllowedJail проверяет, что имя jail состоит только из безопасных символов
func isAllowedJail(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}
	return true
}

func initLogger() *log.Logger {
	file, err := os.OpenFile(LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Printf("Cannot open audit log file %s: %v. Logging to stdout only.", LogPath, err)
		return log.New(os.Stdout, "AGENT: ", log.LstdFlags)
	}
	// Пишем и в файл, и в stdout (чтобы логи попадали в Supervisor)
	multi := io.MultiWriter(file, os.Stdout)
	return log.New(multi, "AGENT: ", log.LstdFlags)
}

func Start() {
	fmt.Println("Agent: Start() called")
	logger := initLogger()
	logger.Println("Starting agent daemon...")
	fmt.Printf("Agent: logging initialized, socket path: %s\n", SocketPath)

	// Создаем директорию для сокета, если она не существует
	socketDir := filepath.Dir(SocketPath)
	if err := os.MkdirAll(socketDir, 0755); err != nil {
		logger.Fatalf("Failed to create socket directory %s: %v", socketDir, err)
	}

	// Удаляем старый сокет, если он существует
	if err := os.RemoveAll(SocketPath); err != nil {
		logger.Fatalf("Failed to remove old socket: %v", err)
	}

	// Создаем слушатель
	listener, err := net.Listen("unix", SocketPath)
	if err != nil {
		logger.Fatalf("Failed to listen on socket %s: %v", SocketPath, err)
	}
	defer listener.Close()

	// Устанавливаем права и владельца (чтобы веб-узел mailadmin мог писать в сокет)
	if err := SetSocketPermissions(SocketPath); err != nil {
		logger.Printf("Warning: failed to set socket permissions: %v", err)
	}

	// Мы не делаем здесь chown, так как пользователя mailadmin может не быть на машине разработчика (Windows).
	// Это обрабатывается установочным скриптом в продакшене.

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req AgentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Printf("Invalid JSON: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		logger.Printf(">>> AGENT REQUEST: Action=%s, Payload=%s, RemoteAddr=%s", req.Action, string(req.Payload), r.RemoteAddr)

		switch req.Action {
		case ActionPing:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("pong"))

		case ActionFail2banUnban:
			var p struct {
				IP   string `json:"ip"`
				Jail string `json:"jail"`
			}
			if err := json.Unmarshal(req.Payload, &p); err != nil {
				http.Error(w, "Invalid payload", http.StatusBadRequest)
				return
			}
			parsedIP := net.ParseIP(p.IP)
			if parsedIP == nil {
				logger.Printf("Invalid IP: %s", p.IP)
				http.Error(w, "Invalid IP", http.StatusBadRequest)
				return
			}
			if !isAllowedJail(p.Jail) {
				logger.Printf("Disallowed jail: %s", p.Jail)
				http.Error(w, "Forbidden jail", http.StatusForbidden)
				return
			}
			out, err := exec.Command("/usr/bin/fail2ban-client", "set", p.Jail, "unbanip", parsedIP.String()).CombinedOutput()
			if err != nil {
				logger.Printf("Exec error: %v, output: %s", err, string(out))
				http.Error(w, "Exec error", http.StatusInternalServerError)
				return
			}
			logger.Printf("Successfully unbanned %s from %s", parsedIP.String(), p.Jail)
			w.WriteHeader(http.StatusOK)

		case ActionFail2banBan:
			var p struct {
				IP   string `json:"ip"`
				Jail string `json:"jail"`
			}
			if err := json.Unmarshal(req.Payload, &p); err != nil {
				http.Error(w, "Invalid payload", http.StatusBadRequest)
				return
			}
			parsedIP := net.ParseIP(p.IP)
			if parsedIP == nil {
				http.Error(w, "Invalid IP", http.StatusBadRequest)
				return
			}
			if !isAllowedJail(p.Jail) {
				http.Error(w, "Forbidden jail", http.StatusForbidden)
				return
			}
			out, err := exec.Command("/usr/bin/fail2ban-client", "set", p.Jail, "banip", parsedIP.String()).CombinedOutput()
			if err != nil {
				logger.Printf("Exec error: %v, output: %s", err, string(out))
				http.Error(w, "Exec error", http.StatusInternalServerError)
				return
			}
			logger.Printf("Successfully banned %s from %s", parsedIP.String(), p.Jail)
			w.WriteHeader(http.StatusOK)

		case ActionFail2banWhitelistAdd:
			var p struct {
				IP string `json:"ip"`
			}
			if err := json.Unmarshal(req.Payload, &p); err != nil {
				http.Error(w, "Invalid payload", http.StatusBadRequest)
				return
			}
			parsedIP := net.ParseIP(p.IP)
			if parsedIP == nil {
				http.Error(w, "Invalid IP", http.StatusBadRequest)
				return
			}

			jailPath, cfg, err := loadJailLocal()
			if err != nil {
				logger.Printf("Failed to load jail.local: %v", err)
				http.Error(w, "Failed to load jail.local", http.StatusInternalServerError)
				return
			}


			section := cfg.Section("DEFAULT")
			key, err := section.GetKey("ignoreip")
			if err != nil {
				_, err = section.NewKey("ignoreip", parsedIP.String())
				if err != nil {
					http.Error(w, "Failed to create ignoreip key", http.StatusInternalServerError)
					return
				}
			} else {
				currentVal := key.Value()
				if !strings.Contains(currentVal, parsedIP.String()) {
					key.SetValue(strings.TrimSpace(currentVal) + " " + parsedIP.String())
				}
			}

			if err := cfg.SaveTo(jailPath); err != nil {
				logger.Printf("Failed to save %s: %v", jailPath, err)
				http.Error(w, "Failed to save jail.local", http.StatusInternalServerError)
				return
			}

			out, err := exec.Command("/usr/bin/fail2ban-client", "reload").CombinedOutput()
			if err != nil {
				logger.Printf("Fail2ban reload error: %v, output: %s", err, string(out))
				http.Error(w, "Fail2ban reload error", http.StatusInternalServerError)
				return
			}

			logger.Printf("Successfully added %s to whitelist in %s", parsedIP.String(), jailPath)
			w.WriteHeader(http.StatusOK)

		case ActionFail2banWhitelistList:
			_, cfg, err := loadJailLocal()
			if err != nil {
				logger.Printf("Failed to load jail.local: %v", err)
				http.Error(w, "Failed to load jail.local", http.StatusInternalServerError)
				return
			}
			
			var ips []string
			section := cfg.Section("DEFAULT")
			if key, err := section.GetKey("ignoreip"); err == nil {
				for _, ip := range strings.Fields(key.Value()) {
					if ip != "" {
						ips = append(ips, ip)
					}
				}
			}
			
			resp, _ := json.Marshal(ips)
			w.Header().Set("Content-Type", "application/json")
			w.Write(resp)

		case ActionFail2banWhitelistDelete:
			var p struct {
				IP string `json:"ip"`
			}
			if err := json.Unmarshal(req.Payload, &p); err != nil {
				http.Error(w, "Invalid payload", http.StatusBadRequest)
				return
			}
			
			jailPath, cfg, err := loadJailLocal()
			if err != nil {
				logger.Printf("Failed to load jail.local: %v", err)
				http.Error(w, "Failed to load jail.local", http.StatusInternalServerError)
				return
			}

			section := cfg.Section("DEFAULT")
			if key, err := section.GetKey("ignoreip"); err == nil {
				currentVal := key.Value()
				var newIPs []string
				for _, ip := range strings.Fields(currentVal) {
					if ip != p.IP {
						newIPs = append(newIPs, ip)
					}
				}
				key.SetValue(strings.Join(newIPs, " "))
				
				if err := cfg.SaveTo(jailPath); err != nil {
					logger.Printf("Failed to save %s: %v", jailPath, err)
					http.Error(w, "Failed to save jail.local", http.StatusInternalServerError)
					return
				}

				out, err := exec.Command("/usr/bin/fail2ban-client", "reload").CombinedOutput()
				if err != nil {
					logger.Printf("Fail2ban reload error: %v, output: %s", err, string(out))
					http.Error(w, "Fail2ban reload error", http.StatusInternalServerError)
					return
				}
			}

			logger.Printf("Successfully removed %s from whitelist in %s", p.IP, jailPath)
			w.WriteHeader(http.StatusOK)

		case ActionQueueDelete:
			var p struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(req.Payload, &p); err != nil {
				http.Error(w, "Invalid payload", http.StatusBadRequest)
				return
			}
			// базовая валидация
			if p.ID == "" || len(p.ID) > 20 {
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}
			out, err := exec.Command("/usr/sbin/postsuper", "-d", p.ID).CombinedOutput()
			if err != nil {
				logger.Printf("Exec error: %v, output: %s", err, string(out))
				http.Error(w, "Exec error", http.StatusInternalServerError)
				return
			}
			logger.Printf("Successfully deleted message %s from queue", p.ID)
			w.WriteHeader(http.StatusOK)

		case ActionQueueFlush:
			out, err := exec.Command("/usr/sbin/postqueue", "-f").CombinedOutput()
			if err != nil {
				logger.Printf("Exec error: %v, output: %s", err, string(out))
				http.Error(w, "Exec error", http.StatusInternalServerError)
				return
			}
			logger.Printf("Successfully flushed queue")
			w.WriteHeader(http.StatusOK)

		case ActionImapStatus:
			out, err := exec.Command("/usr/bin/doveadm", "who").CombinedOutput()
			if err != nil {
				logger.Printf("Doveadm error: %v, output: %s", err, string(out))
				http.Error(w, "Doveadm error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			w.Write(out)

		case ActionFail2banStatus:
			jail := r.URL.Query().Get("jail")
			var out []byte
			var err error
			if jail != "" {
				if !isAllowedJail(jail) {
					logger.Printf("DIAGNOSTIC: Jail '%s' is NOT allowed", jail)
					http.Error(w, "Forbidden jail", http.StatusForbidden)
					return
				}
				out, err = exec.Command("/usr/bin/fail2ban-client", "status", jail).CombinedOutput()
			} else {
				out, err = exec.Command("/usr/bin/fail2ban-client", "status").CombinedOutput()
			}
			if err != nil {
				logger.Printf("Fail2ban error: %v, output: %s", err, string(out))
				http.Error(w, "Fail2ban error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			w.Write(out)

		case ActionServiceStatus:
			out, err := exec.Command("/usr/bin/supervisorctl", "status").CombinedOutput()
			if err != nil {
				logger.Printf("Supervisor error: %v, output: %s", err, string(out))
				http.Error(w, "Supervisor error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			w.Write(out)

		case ActionQueueStatus:
			out, err := exec.Command("/usr/sbin/postqueue", "-p").CombinedOutput()
			if err != nil {
				logger.Printf("Postqueue error: %v, output: %s", err, string(out))
				http.Error(w, "Postqueue error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			w.Write(out)

		default:
			logger.Printf("Unknown action: %s", req.Action)
			http.Error(w, "Unknown action", http.StatusBadRequest)
		}
	})

	server := &http.Server{Handler: mux}
	// В рамках этого рефакторинга права на файл (0660 root:mailadmin) обеспечивают основную защиту.

	logger.Printf("Agent listening on %s", SocketPath)
	if err := server.Serve(listener); err != nil {
		logger.Fatalf("Server error: %v", err)
	}
}

func loadJailLocal() (string, *ini.File, error) {
	jailPath := os.Getenv("F2B_JAIL_LOCAL_PATH")
	if jailPath == "" {
		jailPath = "/etc/fail2ban/jail.local"
	}

	realPath, err := filepath.EvalSymlinks(jailPath)
	if err == nil {
		jailPath = realPath
	}

	cfg, err := ini.LoadSources(ini.LoadOptions{
		PreserveSurroundedQuote: true,
		IgnoreInlineComment:     true,
	}, jailPath)

	if err != nil {
		// Файл не существует или не читается - создаем пустой
		cfg = ini.Empty()
		cfg.NewSection("DEFAULT")
		err = nil
	} else {
		// Проверка на кривой файл (ключи без секции [DEFAULT])
		emptySection := cfg.Section("")
		if len(emptySection.Keys()) > 0 {
			var rescuedIPs string
			if key, errGet := emptySection.GetKey("ignoreip"); errGet == nil {
				rescuedIPs = key.Value()
			}
			
			// Пересоздаем правильный конфиг
			cfg = ini.Empty()
			defSec, _ := cfg.NewSection("DEFAULT")
			if rescuedIPs != "" {
				defSec.NewKey("ignoreip", rescuedIPs)
			}
		}
	}

	return jailPath, cfg, err
}
