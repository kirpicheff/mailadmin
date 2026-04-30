package db

import (
	"fmt"
	"github.com/user/mailadmin/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
	"strings"
)

var DB *gorm.DB

func InitDB(dsn string) {
	var err error
	if dsn == "" {
		dsn = os.Getenv("DB_DSN")
	}
	if dsn == "" {
		dsn = "user:password@tcp(127.0.0.1:3306)/mailadmin?charset=utf8&parseTime=True&loc=Local"
	}
	
	// Принудительно заменяем utf8mb4 на utf8, чтобы избежать Error 1071 (длина индексов)
	// на MariaDB/MySQL с лимитом 1000 байт. 255*3=765 < 1000.
	if strings.Contains(dsn, "charset=utf8mb4") {
		log.Println("Forcing DSN charset from utf8mb4 to utf8 for index compatibility.")
		dsn = strings.Replace(dsn, "charset=utf8mb4", "charset=utf8", 1)
	}

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// 1. Автоматическое исправление кодировок для баз от старого PostfixAdmin
	// Переходим на utf8 (utf8mb3) для всех таблиц. 
	// Это решает проблему Error 1071 (длина ключа), так как 255*3 = 765 байт, что влезает в лимит 1000 байт.
	// Кириллица при этом поддерживается корректно.
	var charSet string
	DB.Raw("SELECT character_set_name FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'domain' AND column_name = 'description' LIMIT 1").Scan(&charSet)

	if charSet != "" && charSet != "utf8" && charSet != "utf8mb3" {
		log.Printf("Old charset detected (%s). Forcing conversion to utf8 (utf8mb3) to avoid index length issues...", charSet)

		// Отключаем проверку внешних ключей
		DB.Exec("SET FOREIGN_KEY_CHECKS = 0")

		// Сбрасываем ключи, которые мешают конвертации
		DB.Exec("ALTER TABLE vacation_notification DROP FOREIGN KEY vacation_notification_pkey")
		DB.Exec("ALTER TABLE vacation_notification DROP FOREIGN KEY vacation_notification_ibfk_1")

		tablesToConvert := []string{
			"domain", "admin", "domain_admins", "mailbox", "alias",
			"alias_domain", "vacation", "vacation_notification", "log", "quota2",
		}

		for _, table := range tablesToConvert {
			var exists int
			DB.Raw("SELECT 1 FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ? LIMIT 1", table).Scan(&exists)
			if exists == 1 {
				log.Printf("Converting table %s to utf8...", table)
				// Используем ROW_FORMAT=DYNAMIC на всякий случай
				DB.Exec(fmt.Sprintf("ALTER TABLE `%s` ROW_FORMAT=DYNAMIC", table))
				// Конвертируем в utf8 (utf8mb3)
				DB.Exec(fmt.Sprintf("ALTER TABLE `%s` CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci", table))
			}
		}

		// Возвращаем ключи
		DB.Exec("ALTER TABLE vacation_notification ADD CONSTRAINT vacation_notification_pkey FOREIGN KEY (on_vacation) REFERENCES vacation(email) ON DELETE CASCADE")
		DB.Exec("SET FOREIGN_KEY_CHECKS = 1")
		log.Println("Database conversion to utf8 completed.")
	} else {
		log.Println("Database charset is already utf8/utf8mb3.")
	}

	// 2. Автоматическая миграция (создание таблиц, если их нет)
	err = DB.AutoMigrate(
		&models.Domain{},
		&models.Mailbox{},
		&models.Alias{},
		&models.Admin{},
		&models.DomainAdmin{},
		&models.AliasDomain{},
		&models.Log{},
		&models.Quota2{},
		&models.AppConfig{},
		&models.Vacation{},
		&models.SieveRule{},
	)
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	log.Println("Database connection established and initialized.")
}
