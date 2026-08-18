package validation

import (
	"log/slog"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	commondomain "github.com/Drathveloper/go-web-skeleton/common/domain"
)

func RegisterValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := v.RegisterValidation("isvalidrole", roleValidator(commondomain.GetAllowedRoles())); err != nil {
			slog.Error("register isvalidrole validator failed", slog.Any("error", err))
			return
		}
		slog.Debug("registered isvalidrole validator")
	}
}
