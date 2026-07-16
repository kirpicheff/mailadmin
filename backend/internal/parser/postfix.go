package parser

import (
	"regexp"
	"strconv"
	"strings"
)

// DeliveryAttempt представляет одну попытку доставки письма получателю
type DeliveryAttempt struct {
	To        string `json:"to"`
	RelayHost string `json:"relay_host"`
	RelayIP   string `json:"relay_ip"`
	Status    string `json:"status"` // sent, deferred, bounced
	DSN       string `json:"dsn"`    // код доставки, например 2.0.0 или 5.1.1
	StatusMsg string `json:"status_msg"`
	Delay     string `json:"delay"`
	Delays    string `json:"delays"`
	IsTLS     bool   `json:"is_tls"`
}

// Transaction представляет жизненный цикл одного письма по его QueueID
type Transaction struct {
	QueueID     string            `json:"queue_id"`
	Timestamp   string            `json:"timestamp"`
	ClientHost  string            `json:"client_host"`
	ClientIP    string            `json:"client_ip"`
	From        string            `json:"from"`
	MessageID   string            `json:"message_id"`
	Size        int               `json:"size"`
	Deliveries  []DeliveryAttempt `json:"deliveries"`
	IsForward   bool              `json:"is_forward"`
	OrigQueueID string            `json:"orig_queue_id"`
}

// RejectInfo представляет информацию о заблокированном подключении (NOQUEUE)
type RejectInfo struct {
	Timestamp string `json:"timestamp"`
	Client    string `json:"client"`
	From      string `json:"from"`
	To        string `json:"to"`
	Reason    string `json:"reason"`
}

// AnalysisResult содержит агрегированные результаты анализа логов
type AnalysisResult struct {
	TotalTransactions int            `json:"total_transactions"`
	SentCount         int            `json:"sent_count"`
	DeferredCount     int            `json:"deferred_count"`
	BouncedCount      int            `json:"bounced_count"`
	RejectCount       int            `json:"reject_count"`
	AverageDelay      float64        `json:"average_delay"`
	Transactions      []Transaction  `json:"transactions"`
	Rejects           []RejectInfo   `json:"rejects"`
	TopSenders        []KeyValue     `json:"top_senders"`
	TopRecipients     []KeyValue     `json:"top_recipients"`
	TopClients        []KeyValue     `json:"top_clients"`
	TopErrors         []KeyValue     `json:"top_errors"`
	TopSASLFailures   []KeyValue     `json:"top_sasl_failures"`
}

// KeyValue вспомогательная структура для топов
type KeyValue struct {
	Key   string `json:"key"`
	Value int    `json:"value"`
}

// Регулярные выражения для разбора логов
var (
	// Базовый разбор строки лога Postfix с захватом PID
	// Пример: Jul 14 21:17:20 hostname postfix/smtpd[12345]: 4B87F40CCF: client=unknown[1.2.3.4]
	lineRegex = regexp.MustCompile(`^([A-Z][a-z]{2}\s+\d+\s+\d{2}:\d{2}:\d{2}|\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:[+-]\d{2}:\d{2}|Z))\s+(\S+)\s+postfix/([^\[:]+)(?:\[(\d+)\])?:(?:\s+([0-9a-zA-Z]+):)?\s*(.*)$`)

	// Разбор информации о клиенте в smtpd
	clientRegex = regexp.MustCompile(`client=([^\[]+)\[([^\]]+)\]`)

	// Разбор отправителя и размера в qmgr
	qmgrRegex = regexp.MustCompile(`from=<([^>]*)>(?:,\s+size=(\d+))?`)

	// Разбор cleanup (Message-ID)
	messageIDRegex = regexp.MustCompile(`message-id=<([^>]*)>`)

	// Разбор доставки smtp/local/virtual/lmtp
	deliveryRegex = regexp.MustCompile(`to=<([^>]*)>(?:,\s+orig_to=<([^>]*)>)?,\s+relay=([^,\[]*)(?:\[([^\]]*)\](?::\d+)?)?,\s+.*?delay=([^,]+),\s+delays=([^,]+),\s+dsn=([^,\s]*),\s+status=(\w+)\s+\((.*)\)`)

	// Разбор отказов (NOQUEUE reject в smtpd и milter-reject в cleanup)
	rejectRegex = regexp.MustCompile(`(?:reject|milter-reject):\s+.*?from\s+([^:]+):\s+(.*?);\s+from=<?([^>\s]*)>?\s+to=<?([^>\s]*)>?`)

	// Разбор SASL Authentication failed
	saslFailRegex = regexp.MustCompile(`warning:.*?(?:unknown)?\[([0-9a-fA-F\.:]+)\].*?SASL [A-Z0-9]+ authentication failed`)

	// Разбор установления TLS
	tlsRegex = regexp.MustCompile(`(?:Anonymous|Trusted) TLS connection established`)
)

// ParsePostfixLogs анализирует срез строк лога Postfix
func ParsePostfixLogs(lines []string) *AnalysisResult {
	transactionsMap := make(map[string]*Transaction)
	var rejects []RejectInfo
	saslFailuresMap := make(map[string]int)
	tlsSessions := make(map[string]bool)

	// Вспомогательная мапа для хранения информации о клиенте по PID процесса smtpd
	clientSessions := make(map[string]struct{ Host, IP string })

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := lineRegex.FindStringSubmatch(line)
		if len(matches) < 7 {
			continue
		}

		timestamp := matches[1]
		// hostname := matches[2]
		component := matches[3]
		pid := matches[4]
		queueID := matches[5]
		payload := matches[6]

		sessionKey := component + ":" + pid

		if tlsRegex.MatchString(payload) {
			tlsSessions[sessionKey] = true
			continue
		}

		if saslMatches := saslFailRegex.FindStringSubmatch(payload); len(saslMatches) >= 2 {
			saslFailuresMap[saslMatches[1]]++
			continue
		}

		// Если это smtpd и есть информация о подключении клиента
		if component == "smtpd" && strings.HasPrefix(payload, "connect from ") {
			clientInfo := strings.TrimPrefix(payload, "connect from ")
			if cMatches := clientRegex.FindStringSubmatch("client=" + clientInfo); len(cMatches) >= 3 {
				// Запоминаем клиента
			}
		}

		// Проверка на reject (как NOQUEUE, так и milter-reject)
		if rMatches := rejectRegex.FindStringSubmatch(payload); len(rMatches) >= 5 {
			rejects = append(rejects, RejectInfo{
				Timestamp: timestamp,
				Client:    rMatches[1],
				Reason:    rMatches[2],
				From:      rMatches[3],
				To:        rMatches[4],
			})
			if queueID == "NOQUEUE" || strings.Contains(payload, "milter-reject:") {
				// Удаляем из транзакций, если он туда уже попал (чтобы не дублировать спам в основном списке)
				if queueID != "NOQUEUE" {
					delete(transactionsMap, queueID)
				}
				continue
			}
		}

		// Если есть QueueID, обрабатываем транзакцию
		if queueID != "" && queueID != "NOQUEUE" {
			tx, exists := transactionsMap[queueID]
			if !exists {
				tx = &Transaction{
					QueueID:   queueID,
					Timestamp: timestamp,
				}
				transactionsMap[queueID] = tx
			}

			switch component {
			case "smtpd":
				if cMatches := clientRegex.FindStringSubmatch(payload); len(cMatches) >= 3 {
					tx.ClientHost = cMatches[1]
					tx.ClientIP = cMatches[2]
				}
			case "cleanup":
				if mMatches := messageIDRegex.FindStringSubmatch(payload); len(mMatches) >= 2 {
					tx.MessageID = mMatches[1]
				}
			case "qmgr":
				if qMatches := qmgrRegex.FindStringSubmatch(payload); len(qMatches) >= 2 {
					tx.From = qMatches[1]
					if len(qMatches) >= 3 && qMatches[2] != "" {
						if sz, err := strconv.Atoi(qMatches[2]); err == nil {
							tx.Size = sz
						}
					}
				}
			case "smtp", "local", "virtual", "pipe", "lmtp":
				if dMatches := deliveryRegex.FindStringSubmatch(payload); len(dMatches) >= 10 {
					to := dMatches[1]
					relayHost := dMatches[3]
					relayIP := dMatches[4]
					delay := dMatches[5]
					delays := dMatches[6]
					dsn := dMatches[7]
					status := dMatches[8]
					statusMsg := dMatches[9]

					// Проверяем наличие перенаправления (forwarded as)
					if strings.Contains(statusMsg, "forwarded as ") {
						tx.IsForward = true
					}

					attempt := DeliveryAttempt{
						To:        to,
						RelayHost: relayHost,
						RelayIP:   relayIP,
						Status:    status,
						DSN:       dsn,
						StatusMsg: statusMsg,
						Delay:     delay,
						Delays:    delays,
						IsTLS:     tlsSessions[sessionKey],
					}
					tx.Deliveries = append(tx.Deliveries, attempt)
				}
			}
		}
	}

	// Агрегация результатов
	result := &AnalysisResult{
		Transactions: make([]Transaction, 0, len(transactionsMap)),
		Rejects:      rejects,
		RejectCount:  len(rejects),
	}

	sendersMap := make(map[string]int)
	recipientsMap := make(map[string]int)
	clientsMap := make(map[string]int)
	errorsMap := make(map[string]int)

	var totalDelay float64
	var delayCount int

	for _, tx := range transactionsMap {
		// Добавляем транзакцию в общий список
		result.Transactions = append(result.Transactions, *tx)
		result.TotalTransactions++

		// Подсчет статистики по отправителям, получателям и клиентам
		if tx.From != "" {
			sendersMap[tx.From]++
		}
		if tx.ClientIP != "" {
			clientsMap[tx.ClientIP]++
		}

		// Подсчет статусов доставки
		for _, del := range tx.Deliveries {
			if del.To != "" {
				recipientsMap[del.To]++
			}
			if del.Delay != "" {
				if d, err := strconv.ParseFloat(del.Delay, 64); err == nil {
					totalDelay += d
					delayCount++
				}
			}
			switch del.Status {
			case "sent":
				result.SentCount++
			case "deferred":
				result.DeferredCount++
				errorsMap[del.DSN]++
			case "bounced":
				result.BouncedCount++
				errorsMap[del.DSN]++
			}
		}
	}

	if delayCount > 0 {
		result.AverageDelay = totalDelay / float64(delayCount)
	}

	// Сортировка и выборка ТОПов
	result.TopSenders = sortMapToKeyValues(sendersMap, 10)
	result.TopRecipients = sortMapToKeyValues(recipientsMap, 10)
	result.TopClients = sortMapToKeyValues(clientsMap, 10)
	result.TopErrors = sortMapToKeyValues(errorsMap, 10)
	result.TopSASLFailures = sortMapToKeyValues(saslFailuresMap, 10)

	// Очищаем сессионные данные
	_ = clientSessions

	return result
}

// Вспомогательная функция для преобразования мапы в сортированный слайс
func sortMapToKeyValues(m map[string]int, limit int) []KeyValue {
	kvs := make([]KeyValue, 0, len(m))
	for k, v := range m {
		kvs = append(kvs, KeyValue{Key: k, Value: v})
	}

	// Простая сортировка пузырьком/выбором (или быстрая), так как списки обычно небольшие
	for i := 0; i < len(kvs); i++ {
		for j := i + 1; j < len(kvs); j++ {
			if kvs[i].Value < kvs[j].Value {
				kvs[i], kvs[j] = kvs[j], kvs[i]
			}
		}
	}

	if len(kvs) > limit {
		return kvs[:limit]
	}
	return kvs
}
