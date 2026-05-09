package agent

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

const SocketPath = "/var/run/mailadmin/agent.sock"
const LogPath = "/var/log/mailadmin-agent.log"

type ActionType string

const (
	ActionFail2banUnban  ActionType = "fail2ban_unban"
	ActionFail2banBan    ActionType = "fail2ban_ban"
	ActionFail2banStatus ActionType = "fail2ban_status"
	ActionQueueDelete    ActionType = "queue_delete"
	ActionQueueFlush     ActionType = "queue_flush"
	ActionQueueStatus    ActionType = "queue_status"
	ActionImapStatus     ActionType = "imap_status"
	ActionServiceStatus  ActionType = "service_status"
	ActionPing           ActionType = "ping"
)

type AgentRequest struct {
	Action  ActionType      `json:"action"`
	Payload json.RawMessage `json:"payload"`
}

var allowedJails = map[string]bool{
	"sshd": true, "postfix": true, "dovecot": true, "sieve": true, "roundcube": true,
}

func initLogger() *log.Logger {
	file, err := os.OpenFile(LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Printf("Cannot open audit log file %s: %v. Logging to stdout only.", LogPath, err)
		return log.New(os.Stdout, "AGENT: ", log.LstdFlags)
	}
	return log.New(file, "AGENT: ", log.LstdFlags)
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

		logger.Printf("Received action: %s", req.Action)

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
			if !allowedJails[p.Jail] {
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
			if !allowedJails[p.Jail] {
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
				if !allowedJails[jail] {
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
	// В идеале мы должны проверять SO_PEERCRED при Accept, но net/http не предоставляет к этому легкого доступа.
	// Для продакшена можно использовать кастомную обертку над слушателем.
	// В рамках этого рефакторинга права на файл (0660 root:mailadmin) обеспечивают основную защиту.

	logger.Printf("Agent listening on %s", SocketPath)
	if err := server.Serve(listener); err != nil {
		logger.Fatalf("Server error: %v", err)
	}
}
