package db

import (
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

	log.Println("Database connection established and migrated.")
}
