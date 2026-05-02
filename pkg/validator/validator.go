package validator

import (
	"github.com/go-playground/validator/v10"
)

type CustomValidator struct {
	Validator *validator.Validate
}

func NewCustomValidator() *CustomValidator {
	validate := validator.New()

	return &CustomValidator{
		Validator: validate,
	}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.Validator.Struct(i)
}
