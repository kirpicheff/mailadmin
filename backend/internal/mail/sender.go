package mail

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

// EmailMessage представляет структуру письма
type EmailMessage struct {
	From     string
	To       []string
	Subject  string
	Body     string
	IsHTML   bool
}

// SendEmail отправляет письмо через локальный или внешний SMTP
func SendEmail(msg *EmailMessage) error {
	// По умолчанию пробуем локальный Postfix без авторизации
	// Для продакшена параметры можно вынести в конфиг
	host := "127.0.0.1"
	port := "25"
	addr := net.JoinHostPort(host, port)

	header := make(map[string]string)
	header["From"] = msg.From
	header["To"] = strings.Join(msg.To, ",")
	header["Subject"] = msg.Subject
	header["MIME-Version"] = "1.0"
	
	contentType := "text/plain; charset=\"utf-8\""
	if msg.IsHTML {
		contentType = "text/html; charset=\"utf-8\""
	}
	header["Content-Type"] = contentType
	header["Content-Transfer-Encoding"] = "base64"

	// Формируем тело в Base64 для корректного отображения кириллицы
	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + msg.Body

	// Для локального Postfix обычно авторизация не нужна
	// Если нужно будет добавить поддержку Auth - расширим структуру
	
	return smtp.SendMail(addr, nil, msg.From, msg.To, []byte(message))
}

// SendWithTLS для случаев, когда нужен защищенный порт (например 465 или 587 с STARTTLS)
func SendWithTLS(host, port, user, password string, msg *EmailMessage) error {
    auth := smtp.PlainAuth("", user, password, host)
    tlsconfig := &tls.Config{
        InsecureSkipVerify: true,
        ServerName: host,
    }

    conn, err := tls.Dial("tcp", host+":"+port, tlsconfig)
    if err != nil {
        return err
    }

    c, err := smtp.NewClient(conn, host)
    if err != nil {
        return err
    }

    if err = c.Auth(auth); err != nil {
        return err
    }

    if err = c.Mail(msg.From); err != nil {
        return err
    }

    for _, addr := range msg.To {
        if err = c.Rcpt(addr); err != nil {
            return err
        }
    }

    w, err := c.Data()
    if err != nil {
        return err
    }

    _, err = w.Write([]byte(msg.Body))
    if err != nil {
        return err
    }

    err = w.Close()
    if err != nil {
        return err
    }

    return c.Quit()
}
