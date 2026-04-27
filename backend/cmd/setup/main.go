package main

import (
	"fmt"
	"log"
	"time"

	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/sha512_crypt"
	"github.com/user/mailadmin/internal/config"
	"github.com/user/mailadmin/internal/db"
	"github.com/user/mailadmin/internal/models"
)

func main() {
	cfg := config.LoadConfig()
	db.InitDB(cfg.DBDSN)

	username := "admin@example.local" // Поменяйте на свой
	password := "admin123"

	// Хешируем пароль в SHA512-CRYPT
	cryptService := crypt.New(crypt.SHA512)
	hash, err := cryptService.Generate([]byte(password), []byte("$6$salt"))
	if err != nil {
		log.Fatal(err)
	}

	admin := models.Admin{
		Username:       username,
		Password:       hash,
		Active:         true,
		SuperAdmin:     true,
		PasswordExpiry: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), // Просрочен
	}

	// Проверяем, существует ли уже такой админ
	var count int64
	db.DB.Model(&models.Admin{}).Where("username = ?", username).Count(&count)
	if count > 0 {
		fmt.Printf("Администратор %s уже существует. Сбрасываю пароль и срок действия...\n", username)
		db.DB.Model(&models.Admin{}).Where("username = ?", username).Updates(map[string]interface{}{
			"password":        admin.Password,
			"password_expiry": admin.PasswordExpiry,
		})
	} else {
		if err := db.DB.Create(&admin).Error; err != nil {
			log.Fatalf("Ошибка при создании админа: %v", err)
		}
		fmt.Printf("Создан суперадмин: %s\nПароль: %s\nТребуется смена при входе.\n", username, password)
	}
}
