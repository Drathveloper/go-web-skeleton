package validation_test

import (
	"testing"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"

	"github.com/Drathveloper/go-web-skeleton/common/http/validation"
)

// The tag is `binding`, not `validate`: gin's validator engine is configured
// with SetTagName("binding"), so a `validate` tag would simply never run — which
// is why every DTO in the project declares its rules under `binding`.
type roleForm struct {
	Role  string   `binding:"isvalidrole"`
	Roles []string `binding:"dive,isvalidrole"`
}

// The `isvalidrole` tag is the only thing standing between a form field and the
// role column: a role the application does not know about would be stored and
// then silently match nothing in Authorize.
func TestRegisterValidators(t *testing.T) {
	validation.RegisterValidators()

	engine, ok := binding.Validator.Engine().(*validator.Validate)
	require.True(t, ok)

	tests := []struct {
		name    string
		role    string
		wantErr bool
	}{
		{
			name: "test admin is a valid role",
			role: "admin",
		},
		{
			name: "test user is a valid role",
			role: "user",
		},
		{
			name: "test an empty role is left to the required tag to reject",
			role: "",
		},
		{
			name:    "test an unknown role is rejected",
			role:    "superadmin",
			wantErr: true,
		},
		{
			name:    "test the role check is case sensitive",
			role:    "ADMIN",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.Struct(roleForm{Role: tt.role, Roles: []string{tt.role}})

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
