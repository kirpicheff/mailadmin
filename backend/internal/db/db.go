package db

import (
	"fmt"
	"github.com/user/mailadmin/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

var DB *gorm.DB

func InitDB(dsn string) {
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// 1. Автоматическое исправление кодировок для баз от старого PostfixAdmin
	// Делаем это ДО AutoMigrate, иначе MariaDB выдаст ошибку при попытке изменить колонки с "неправильными" данными
	var charSet string
	DB.Raw("SELECT character_set_name FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'domain' AND column_name = 'description' LIMIT 1").Scan(&charSet)

	// Если кодировка не utf8mb4 (включая utf8mb3/utf8), пытаемся обновиться
	if charSet != "" && charSet != "utf8mb4" {
		log.Printf("Database charset is '%s'. Forcing conversion to utf8mb4/utf8mb3...", charSet)

		// Отключаем проверку внешних ключей
		DB.Exec("SET FOREIGN_KEY_CHECKS = 0")

		// Специально для MariaDB/PostfixAdmin: дропаем ключи, которые мешают конвертации даже при FOREIGN_KEY_CHECKS=0
		DB.Exec("ALTER TABLE vacation_notification DROP FOREIGN KEY vacation_notification_pkey")
		DB.Exec("ALTER TABLE vacation_notification DROP FOREIGN KEY vacation_notification_ibfk_1")

		tablesToConvert := []string{
			"domain", "admin", "domain_admins", "mailbox", "alias",
			"alias_domain", "vacation", "vacation_notification", "log", "quota2",
		}

		// Конвертируем существующие таблицы
		for _, table := range tablesToConvert {
			var exists int
			DB.Raw("SELECT 1 FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ? LIMIT 1", table).Scan(&exists)
			if exists == 1 {
				log.Printf("Processing table %s...", table)
				
				// Пробуем сделать ROW_FORMAT=DYNAMIC для поддержки длинных индексов (Error 1071)
				DB.Exec(fmt.Sprintf("ALTER TABLE `%s` ROW_FORMAT=DYNAMIC", table))

				// Пытаемся конвертировать в utf8mb4
				err := DB.Exec(fmt.Sprintf("ALTER TABLE `%s` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", table)).Error
				if err != nil {
					log.Printf("Table %s: utf8mb4 failed (%v), falling back to utf8mb3...", table, err)
					// Если utf8mb4 не пролез (старая MariaDB или слишком длинные ключи), используем обычный utf8
					DB.Exec(fmt.Sprintf("ALTER TABLE `%s` CONVERT TO CHARACTER SET utf8 COLLATE utf8_general_ci", table))
				}
			}
		}

		// Возвращаем ключи обратно
		DB.Exec("ALTER TABLE vacation_notification ADD CONSTRAINT vacation_notification_pkey FOREIGN KEY (on_vacation) REFERENCES vacation(email) ON DELETE CASCADE")
		DB.Exec("SET FOREIGN_KEY_CHECKS = 1")
		log.Println("Database conversion step completed.")
	} else if charSet == "utf8mb4" {
		log.Println("Database charsets are already utf8mb4.")
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
