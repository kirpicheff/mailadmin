package api

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
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
	return &CustomValidator{validator: validator.New()}
}
