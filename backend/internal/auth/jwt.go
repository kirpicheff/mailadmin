package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Username           string   `json:"username"`
	SuperAdmin         bool     `json:"superadmin"`
	MustChangePassword bool     `json:"must_change_password"`
	IsAPIToken         bool     `json:"is_api_token"`
	APIScopes          []string `json:"api_scopes"`
	jwt.RegisteredClaims
}

// GenerateAccessToken создает токен доступа (60 минут)
func GenerateAccessToken(username string, isSuper bool, mustChange bool, secret string) (string, error) {
	claims := &Claims{
		Username:           username,
		SuperAdmin:         isSuper,
		MustChangePassword: mustChange,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(60 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateRefreshToken создает долгоживущий токен обновления (7 дней)
func GenerateRefreshToken(username string, secret string) (string, error) {
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken проверяет валидность токена и возвращает Claims
func ValidateToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// HIGH-6: явная проверка алгоритма подписи — защита от atack alg:none
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
