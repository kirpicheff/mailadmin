package api

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

var (
	domainRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	emailRegex  = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

// CustomValidator — реализация интерфейса echo.Validator
type CustomValidator struct {
	validator *validator.Validate
}

// Validate выполняет валидацию структуры
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// В реальности здесь можно распарсить ошибку и вернуть более красивый JSON
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

// NewValidator создает новый экземпляр валидатора
func NewValidator() *CustomValidator {
	v := validator.New()

	_ = v.RegisterValidation("email_or_catchall", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		if strings.HasPrefix(val, "@") {
			domain := val[1:]
			return domainRegex.MatchString(domain) && strings.Contains(domain, ".")
		}
		return emailRegex.MatchString(val)
	})

	return &CustomValidator{validator: v}
}
