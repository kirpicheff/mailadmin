package auth

import (
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/md5_crypt"    // Импортируем для регистрации алгоритма $1$
	_ "github.com/GehirnInc/crypt/sha512_crypt" // Импортируем для регистрации алгоритма $6$
)

// CheckPassword проверяет соответствие пароля хешу.
// Поддерживает форматы:
// $1$ - MD5-Crypt (старый PostfixAdmin)
// $6$ - SHA512-Crypt (современный стандарт)
func CheckPassword(password, hash string) (bool, error) {
	// Если хеш пустой - доступа нет
	if hash == "" {
		return false, nil
	}

	var cryptService crypt.Crypter

	if strings.HasPrefix(hash, "$1$") {
		cryptService = crypt.New(crypt.MD5)
	} else if strings.HasPrefix(hash, "$6$") {
		cryptService = crypt.New(crypt.SHA512)
	} else {
		return false, fmt.Errorf("unsupported hash format: %s", hash[:min(len(hash), 4)])
	}

	err := cryptService.Verify(hash, []byte(password))
	if err != nil {
		return false, err
	}

	return true, nil
}

// GenerateHash генерирует SHA512-Crypt хеш со случайной солью
func GenerateHash(password string) (string, error) {
	saltChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789./"
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// Конвертируем случайные байты в валидные символы соли
	for i := range salt {
		salt[i] = saltChars[int(salt[i])%len(saltChars)]
	}

	saltStr := fmt.Sprintf("$6$%s", string(salt))

	cryptService := crypt.New(crypt.SHA512)
	hash, err := cryptService.Generate([]byte(password), []byte(saltStr))
	if err != nil {
		return "", err
	}

	return hash, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
