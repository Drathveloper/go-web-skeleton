package validation

import (
	"slices"

	"github.com/go-playground/validator/v10"
)

func roleValidator(allowedRoles []string) func(validator.FieldLevel) bool {
	return func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		if value == "" {
			return true
		}
		return slices.Contains(allowedRoles, value)
	}
}
