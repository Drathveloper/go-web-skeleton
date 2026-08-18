package wire

import (
	"fmt"
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/non-standard/validators"
)

const injectValidatorsBaseErrMsg = "inject validators failed"

type RequiredValidators struct {
	Validate *validator.Validate
}

var decimalRegex = regexp.MustCompile(`^-?\d+(\.\d{1,2})?$`)

func is2Decimal(fl validator.FieldLevel) bool {
	return decimalRegex.MatchString(fl.Field().String())
}

func injectValidators(container *Container) error {
	validate := validator.New()

	if err := validate.RegisterValidation("notblank", validators.NotBlank); err != nil {
		return fmt.Errorf("%s: %w", injectValidatorsBaseErrMsg, err)
	}
	if err := validate.RegisterValidation("decimal2", is2Decimal); err != nil {
		return fmt.Errorf("%s: %w", injectValidatorsBaseErrMsg, err)
	}

	if ginValidate, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := ginValidate.RegisterValidation("notblank", validators.NotBlank); err != nil {
			return fmt.Errorf("%s: %w", injectValidatorsBaseErrMsg, err)
		}
		if err := ginValidate.RegisterValidation("decimal2", is2Decimal); err != nil {
			return fmt.Errorf("%s: %w", injectValidatorsBaseErrMsg, err)
		}
	}

	container.Validate = validate
	return nil
}
