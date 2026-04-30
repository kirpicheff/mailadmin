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

	if charSet != "" && charSet != "utf8mb4" {
		log.Println("Old charset detected (" + charSet + "). Forcing database conversion to utf8mb4 before migration...")

		// Отключаем проверку внешних ключей, чтобы MariaDB разрешила конвертировать базу
		DB.Exec("SET FOREIGN_KEY_CHECKS = 0")

		tablesToConvert := []string{
			"domain", "admin", "domain_admins", "mailbox", "alias",
			"alias_domain", "vacation", "vacation_notification", "log", "quota2",
		}

		// Конвертируем существующие таблицы
		for _, table := range tablesToConvert {
			var exists int
			DB.Raw("SELECT 1 FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ? LIMIT 1", table).Scan(&exists)
			if exists == 1 {
				log.Printf("Converting table %s to utf8mb4...", table)
				DB.Exec(fmt.Sprintf("ALTER TABLE `%s` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", table))
			}
		}

		DB.Exec("SET FOREIGN_KEY_CHECKS = 1")
		log.Println("Database charsets forced to utf8mb4 successfully.")
	} else if charSet == "utf8mb4" {
		log.Println("Database charsets are already up-to-date (utf8mb4).")
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
