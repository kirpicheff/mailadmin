package audit

import (
	"github.com/user/mailadmin/internal/models"
	"gorm.io/gorm"
	"time"
)

// Log записывает действие в таблицу логов
func Log(db *gorm.DB, username, domain, action, data string) {
	log := models.Log{
		Timestamp: time.Now(),
		Username:  username,
		Domain:    domain,
		Action:    action,
		Data:      data,
	}
	// Мы игнорируем ошибку записи лога, чтобы не прерывать основную операцию,
	// но в реальной системе здесь стоило бы хотя бы вывести ошибку в системный лог.
	db.Create(&log)
}
