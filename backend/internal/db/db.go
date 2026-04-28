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

	// Автоматическая миграция (создание таблиц, если их нет)
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
	)
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// Автоматическое исправление кодировок для баз от старого PostfixAdmin
	// Проверяем, нужна ли конвертация (чтобы не дергать DROP/ADD FOREIGN KEY при каждом запуске)
	var charSet string
	DB.Raw("SELECT character_set_name FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'vacation' AND column_name = 'subject' LIMIT 1").Scan(&charSet)
	
	if charSet != "" && charSet != "utf8mb4" {
		log.Println("Old charset detected (" + charSet + "). Forcing database conversion to utf8mb4...")
		
		// 1. Временно убиваем внешний ключ, чтобы MariaDB разрешила конвертировать базу
		DB.Exec("ALTER TABLE vacation_notification DROP FOREIGN KEY vacation_notification_pkey")
		DB.Exec("ALTER TABLE vacation_notification DROP FOREIGN KEY vacation_notification_ibfk_1") // На случай других стандартов
		
		tablesToConvert := []string{
			"domain", "admin", "domain_admins", "mailbox", "alias", 
			"alias_domain", "vacation", "vacation_notification", "log",
		}
		
		// 2. Конвертируем все таблицы
		for _, table := range tablesToConvert {
			DB.Exec(fmt.Sprintf("ALTER TABLE `%s` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", table))
		}
		
		// 3. Возвращаем ключ обратно
		DB.Exec("ALTER TABLE vacation_notification ADD CONSTRAINT vacation_notification_pkey FOREIGN KEY (on_vacation) REFERENCES vacation(email) ON DELETE CASCADE")
		
		log.Println("Database charsets forced to utf8mb4 successfully.")
	} else {
		log.Println("Database charsets are already up-to-date (utf8mb4).")
	}

	log.Println("Database connection established and initialized.")
}
