package parser

import (
	"bufio"
	"os"
	"testing"
)

func TestRealLogFile(t *testing.T) {
	file, err := os.Open("c:\\Users\\user\\Documents\\MailAdmin\\mail.log")
	if err != nil {
		t.Skip("mail.log не найден, пропускаем тест реального лога")
		return
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() && len(lines) < 10000 {
		lines = append(lines, scanner.Text())
	}

	result := ParsePostfixLogs(lines)
	t.Logf("Анализ первых %d строк mail.log:", len(lines))
	t.Logf("Всего транзакций: %d", result.TotalTransactions)
	t.Logf("Успешно отправлено (sent): %d", result.SentCount)
	t.Logf("Отложено (deferred): %d", result.DeferredCount)
	t.Logf("Ошибка доставки (bounced): %d", result.BouncedCount)
	t.Logf("Отклонено (NOQUEUE reject): %d", result.RejectCount)
	t.Logf("Топ отправителей: %+v", result.TopSenders)
	t.Logf("Топ получателей: %+v", result.TopRecipients)
	t.Logf("Топ клиентов: %+v", result.TopClients)
}

func TestParsePostfixLogs(t *testing.T) {
	// Имитация логов Postfix
	testLogs := []string{
		// 1. Успешная отправка (sent)
		"Jul 14 12:00:01 mailserver postfix/smtpd[1001]: 4B87F40CCF: client=sender-host[1.2.3.4]",
		"Jul 14 12:00:02 mailserver postfix/cleanup[1002]: 4B87F40CCF: message-id=<test-msg-1@sender.com>",
		"Jul 14 12:00:03 mailserver postfix/qmgr[1003]: 4B87F40CCF: from=<sender@sender.com>, size=1024, nrcpt=1",
		"Jul 14 12:00:04 mailserver postfix/smtp[1004]: 4B87F40CCF: to=<recipient@dest.com>, relay=mail.dest.com[5.6.7.8]:25, delay=3, delays=0.1/0/2.8/0.1, dsn=2.0.0, status=sent (250 2.0.0 OK)",

		// 2. Отказ на входе (NOQUEUE reject)
		"Jul 14 12:05:00 mailserver postfix/smtpd[1001]: NOQUEUE: reject: RCPT from bad-host[9.9.9.9]: 554 5.7.1 <bad-user@dest.com>: Recipient address rejected: Access denied; from=<spammer@bad.com> to=<bad-user@dest.com>",

		// 3. Ошибка доставки (deferred)
		"Jul 14 12:10:01 mailserver postfix/qmgr[1003]: 5C99E10DD3: from=<sender2@sender.com>, size=2048, nrcpt=1",
		"Jul 14 12:10:02 mailserver postfix/smtp[1004]: 5C99E10DD3: to=<slow-recipient@dest.com>, relay=none, delay=10, delays=0.1/9.9/0/0, dsn=4.4.1, status=deferred (connect to mail.slow.com[1.1.1.1]: Connection timed out)",
	}

	result := ParsePostfixLogs(testLogs)

	// Проверяем количество транзакций и отказов
	if result.TotalTransactions != 2 {
		t.Errorf("Ожидалось 2 транзакции, получено: %d", result.TotalTransactions)
	}

	if result.RejectCount != 1 {
		t.Errorf("Ожидался 1 отказ NOQUEUE, получено: %d", result.RejectCount)
	}

	// Проверяем транзакцию 4B87F40CCF (успешную)
	var tx1 *Transaction
	for _, tx := range result.Transactions {
		if tx.QueueID == "4B87F40CCF" {
			tx1 = &tx
			break
		}
	}

	if tx1 == nil {
		t.Fatal("Транзакция 4B87F40CCF не найдена")
	}

	if tx1.From != "sender@sender.com" {
		t.Errorf("Неверный отправитель: %s", tx1.From)
	}

	if tx1.ClientIP != "1.2.3.4" || tx1.ClientHost != "sender-host" {
		t.Errorf("Неверные данные клиента: host=%s, ip=%s", tx1.ClientHost, tx1.ClientIP)
	}

	if tx1.MessageID != "test-msg-1@sender.com" {
		t.Errorf("Неверный Message-ID: %s", tx1.MessageID)
	}

	if len(tx1.Deliveries) != 1 {
		t.Errorf("Ожидалась 1 попытка доставки, получено: %d", len(tx1.Deliveries))
	} else {
		del := tx1.Deliveries[0]
		if del.To != "recipient@dest.com" || del.Status != "sent" || del.DSN != "2.0.0" {
			t.Errorf("Неверные данные доставки: %+v", del)
		}
	}

	// Проверяем отказ NOQUEUE
	if len(result.Rejects) != 1 {
		t.Fatal("Отказ NOQUEUE не найден в списке Rejects")
	}

	rej := result.Rejects[0]
	if rej.From != "spammer@bad.com" || rej.To != "bad-user@dest.com" || rej.Client != "bad-host[9.9.9.9]" {
		t.Errorf("Неверные данные отказа: %+v", rej)
	}
}
