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
	Transactions      []Transaction  `json:"transactions"`
	Rejects           []RejectInfo   `json:"rejects"`
	TopSenders        []KeyValue     `json:"top_senders"`
	TopRecipients     []KeyValue     `json:"top_recipients"`
	TopClients        []KeyValue     `json:"top_clients"`
}

// KeyValue вспомогательная структура для топов
type KeyValue struct {
	Key   string `json:"key"`
	Value int    `json:"value"`
}

// Регулярные выражения для разбора логов
var (
	// Базовый разбор строки лога Postfix
	// Пример: Jul 14 21:17:20 hostname postfix/smtpd[12345]: 4B87F40CCF: client=unknown[1.2.3.4]
	// Также поддерживает ISO 8601: 2026-07-01T10:15:32.818512+03:00
	lineRegex = regexp.MustCompile(`^([A-Z][a-z]{2}\s+\d+\s+\d{2}:\d{2}:\d{2}|\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:[+-]\d{2}:\d{2}|Z))\s+(\S+)\s+postfix/([^\[:]+)(?:\[\d+\])?:(?:\s+([0-9a-zA-Z]+):)?\s*(.*)$`)

	// Разбор информации о клиенте в smtpd
	// Пример: client=unknown[1.2.3.4] или client=localhost[127.0.0.1]
	clientRegex = regexp.MustCompile(`client=([^\[]+)\[([^\]]+)\]`)

	// Разбор отправителя и размера в qmgr
	// Пример: from=<sender@example.com>, size=1432, nrcpt=1
	qmgrRegex = regexp.MustCompile(`from=<([^>]*)>(?:,\s+size=(\d+))?`)

	// Разбор cleanup (Message-ID)
	// Пример: message-id=<unique@message.id>
	messageIDRegex = regexp.MustCompile(`message-id=<([^>]*)>`)

	// Разбор доставки smtp/local/virtual
	// Пример: to=<recipient@dest.com>, relay=mail.dest.com[5.6.7.8]:25, delay=0.5, delays=0.1/0/0.3/0.1, dsn=2.0.0, status=sent (250 2.0.0 OK)
	deliveryRegex = regexp.MustCompile(`to=<([^>]*)>(?:,\s+orig_to=<([^>]*)>)?,\s+relay=([^,\[]*)(?:\[([^\]]*)\](?::\d+)?)?,\s+.*?dsn=([^,\s]*),\s+status=(\w+)\s+\((.*)\)`)

	// Разбор NOQUEUE reject в smtpd
	// Пример: reject: RCPT from unknown[1.2.3.4]: 554 5.7.1 <user@dest.com>: Recipient address rejected: Access denied; from=<bad@sender.com> to=<user@dest.com> proto=ESMTP helo=<bad>
	rejectRegex = regexp.MustCompile(`reject:\s+.*?from\s+([^:]+):\s+(\d{3}\s+[\d\.]+.*?);\s+from=<([^>]*)>\s+to=<([^>]*)>`)
)

// ParsePostfixLogs анализирует срез строк лога Postfix
func ParsePostfixLogs(lines []string) *AnalysisResult {
	transactionsMap := make(map[string]*Transaction)
	var rejects []RejectInfo

	// Вспомогательная мапа для хранения информации о клиенте по PID процесса smtpd
	// так как smtpd логирует коннект и транзакцию в разных строках
	clientSessions := make(map[string]struct{ Host, IP string })

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := lineRegex.FindStringSubmatch(line)
		if len(matches) < 6 {
			continue
		}

		timestamp := matches[1]
		// hostname := matches[2]
		component := matches[3]
		queueID := matches[4]
		payload := matches[5]

		// Если это smtpd и есть информация о подключении клиента
		if component == "smtpd" && strings.HasPrefix(payload, "connect from ") {
			clientInfo := strings.TrimPrefix(payload, "connect from ")
			if cMatches := clientRegex.FindStringSubmatch("client=" + clientInfo); len(cMatches) >= 3 {
				// Запоминаем клиента (в реальном логе мы не знаем PID здесь напрямую без парсинга процесса, 
				// но можем временно привязать сессию, если это важно. В данном случае мы можем использовать
				// QueueID, когда он появится, так как postfix пишет client= при установлении QueueID)
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
			case "smtp", "local", "virtual", "pipe":
				if dMatches := deliveryRegex.FindStringSubmatch(payload); len(dMatches) >= 8 {
					to := dMatches[1]
					relayHost := dMatches[3]
					relayIP := dMatches[4]
					dsn := dMatches[5]
					status := dMatches[6]
					statusMsg := dMatches[7]

					// Проверяем наличие перенаправления (forwarded as)
					if strings.Contains(statusMsg, "forwarded as ") {
						tx.IsForward = true
						// Здесь можно выцепить ID нового письма при необходимости
					}

					attempt := DeliveryAttempt{
						To:        to,
						RelayHost: relayHost,
						RelayIP:   relayIP,
						Status:    status,
						DSN:       dsn,
						StatusMsg: statusMsg,
					}
					tx.Deliveries = append(tx.Deliveries, attempt)
				}
			}
		} else if queueID == "NOQUEUE" {
			// Обработка отказов без помещения в очередь
			if rMatches := rejectRegex.FindStringSubmatch(payload); len(rMatches) >= 5 {
				rejects = append(rejects, RejectInfo{
					Timestamp: timestamp,
					Client:    rMatches[1],
					Reason:    rMatches[2],
					From:      rMatches[3],
					To:        rMatches[4],
				})
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
			switch del.Status {
			case "sent":
				result.SentCount++
			case "deferred":
				result.DeferredCount++
			case "bounced":
				result.BouncedCount++
			}
		}
	}

	// Сортировка и выборка ТОПов
	result.TopSenders = sortMapToKeyValues(sendersMap, 10)
	result.TopRecipients = sortMapToKeyValues(recipientsMap, 10)
	result.TopClients = sortMapToKeyValues(clientsMap, 10)

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
