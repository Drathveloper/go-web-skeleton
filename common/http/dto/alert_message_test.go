package dto_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Drathveloper/go-web-skeleton/common/http/dto"
)

// None of these constructors takes an error, and that is the contract: the
// message they carry is displayed verbatim, so it has to be text meant for a
// user. Handing them err.Error() is how driver output, SQL fragments and
// hostnames end up on screen; the error belongs in the log.
func TestAlertMessageConstructors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		actual   dto.AlertMessage
		expected dto.AlertMessage
	}{
		{
			name:   "test error message carries the status code",
			actual: dto.NewErrorMsg(http.StatusForbidden, "someTitle", "someMessage"),
			expected: dto.AlertMessage{
				Code: http.StatusForbidden, Title: "someTitle", Message: "someMessage", Type: "error",
			},
		},
		{
			name:     "test warning message has no status code",
			actual:   dto.NewWarningMsg("someTitle", "someMessage"),
			expected: dto.AlertMessage{Title: "someTitle", Message: "someMessage", Type: "warning"},
		},
		{
			name:     "test info message has no status code",
			actual:   dto.NewInfoMsg("someTitle", "someMessage"),
			expected: dto.AlertMessage{Title: "someTitle", Message: "someMessage", Type: "info"},
		},
		{
			name:     "test success message has no status code",
			actual:   dto.NewSuccessMsg("someTitle", "someMessage"),
			expected: dto.AlertMessage{Title: "someTitle", Message: "someMessage", Type: "success"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.expected, tt.actual)
		})
	}
}
