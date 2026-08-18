package domain_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Drathveloper/go-web-skeleton/common/domain"
)

func TestSession_AddAlertMessages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		baseAlertMessages     domain.AlertMessages
		appendAlertMessages   domain.AlertMessages
		expectedAlertMessages domain.AlertMessages
	}{
		{
			name:              "test session should add alert messages when alert messages are nil",
			baseAlertMessages: nil,
			appendAlertMessages: domain.AlertMessages{
				{
					Title:   "someTitle",
					Message: "someMessage",
					Type:    "error",
					Code:    http.StatusForbidden,
				},
			},
			expectedAlertMessages: domain.AlertMessages{
				{
					Title:   "someTitle",
					Message: "someMessage",
					Type:    "error",
					Code:    http.StatusForbidden,
				},
			},
		},
		{
			name: "test session should add alert messages when alert messages are already present",
			baseAlertMessages: domain.AlertMessages{
				{
					Title:   "someTitle",
					Message: "someMessage",
					Type:    "error",
					Code:    http.StatusForbidden,
				},
			},
			appendAlertMessages: domain.AlertMessages{
				{
					Title:   "someTitle2",
					Message: "someMessage2",
					Type:    "error",
					Code:    http.StatusInternalServerError,
				},
			},
			expectedAlertMessages: domain.AlertMessages{
				{
					Title:   "someTitle",
					Message: "someMessage",
					Type:    "error",
					Code:    http.StatusForbidden,
				},
				{
					Title:   "someTitle2",
					Message: "someMessage2",
					Type:    "error",
					Code:    http.StatusInternalServerError,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			session := domain.Session{
				AlertMessages: tt.baseAlertMessages,
			}

			session.AddAlertMessages(tt.appendAlertMessages...)

			require.Equal(t, tt.expectedAlertMessages, session.AlertMessages)
		})
	}
}

// FlushAlertMessages has to leave the session with nothing to render: the flash
// messages are shown once and the middleware relies on this to not repeat them
// on the next page.
func TestSession_FlushAlertMessages(t *testing.T) {
	t.Parallel()

	session := domain.Session{
		AlertMessages: domain.AlertMessages{
			domain.NewErrorAlertMessage(http.StatusForbidden, "someTitle", "someMessage"),
		},
	}

	session.FlushAlertMessages()

	require.Empty(t, session.AlertMessages)
}

func TestSession_AlertMessageConstructors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		actual   domain.AlertMessage
		expected domain.AlertMessage
	}{
		{
			name:   "test error alert message carries the status code and the error type",
			actual: domain.NewErrorAlertMessage(http.StatusForbidden, "someTitle", "someMessage"),
			expected: domain.AlertMessage{
				Code:    http.StatusForbidden,
				Title:   "someTitle",
				Message: "someMessage",
				Type:    "error",
			},
		},
		{
			name:   "test success alert message has no status code",
			actual: domain.NewSuccessAlertMessage("someTitle", "someMessage"),
			expected: domain.AlertMessage{
				Code:    0,
				Title:   "someTitle",
				Message: "someMessage",
				Type:    "success",
			},
		},
		{
			name:   "test warning alert message has no status code",
			actual: domain.NewWarningAlertMessage("someTitle", "someMessage"),
			expected: domain.AlertMessage{
				Code:    0,
				Title:   "someTitle",
				Message: "someMessage",
				Type:    "warning",
			},
		},
		{
			name:   "test info alert message has no status code",
			actual: domain.NewInfoAlertMessage("someTitle", "someMessage"),
			expected: domain.AlertMessage{
				Code:    0,
				Title:   "someTitle",
				Message: "someMessage",
				Type:    "info",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.expected, tt.actual)
		})
	}
}

func TestGetAllowedRoles(t *testing.T) {
	t.Parallel()

	require.Equal(t, []string{"admin", "user"}, domain.GetAllowedRoles())
}
