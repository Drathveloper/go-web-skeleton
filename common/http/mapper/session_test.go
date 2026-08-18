package mapper_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/common/http/dto"
	"github.com/Drathveloper/go-web-skeleton/common/http/mapper"
	"github.com/Drathveloper/go-web-skeleton/common/i18n"
)

// MapDomainErrorToViewResponse takes message identifiers, not text. That is the
// whole point of the signature: there is no parameter a caller could hand an
// err.Error() to, so driver messages, SQL and hostnames cannot reach the error
// page. What it must produce is the localized catalog text.
func TestMapDomainErrorToViewResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		expectedResp *dto.ViewResponse[any]
		name         string
		lang         string
		titleKey     string
		messageKey   string
		code         int
	}{
		{
			name:       "test map domain error should localize title and message in english",
			lang:       "en",
			code:       http.StatusForbidden,
			titleKey:   "errors.title",
			messageKey: "errors.forbidden",
			expectedResp: &dto.ViewResponse[any]{
				Language: "en",
				Msgs: []dto.AlertMessage{
					{
						Title:   "Error",
						Message: "You do not have permission to perform this action.",
						Type:    "error",
						Code:    http.StatusForbidden,
					},
				},
			},
		},
		{
			name:       "test map domain error should localize title and message in spanish",
			lang:       "es",
			code:       http.StatusInternalServerError,
			titleKey:   "errors.title",
			messageKey: "errors.internal",
			expectedResp: &dto.ViewResponse[any]{
				Language: "es",
				Msgs: []dto.AlertMessage{
					{
						Title:   "Error",
						Message: "Error interno del servidor.",
						Type:    "error",
						Code:    http.StatusInternalServerError,
					},
				},
			},
		},
		{
			name:       "test map domain error should fall back to the default language when none is given",
			lang:       "",
			code:       http.StatusNotFound,
			titleKey:   "errors.title",
			messageKey: "errors.not_found",
			expectedResp: &dto.ViewResponse[any]{
				Language: i18n.DefaultLanguage,
				Msgs: []dto.AlertMessage{
					{
						Title:   "Error",
						Message: "The requested resource does not exist.",
						Type:    "error",
						Code:    http.StatusNotFound,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual := mapper.MapDomainErrorToViewResponse(tt.lang, tt.code, tt.titleKey, tt.messageKey)

			require.Equal(t, tt.expectedResp, actual)
			require.NotEqual(t, tt.messageKey, actual.Msgs[0].Message,
				"the message identifier itself must never be what the user reads")
		})
	}
}

func TestMapDataToViewResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		session  *domain.Session
		expected *dto.ViewResponse[string]
		name     string
	}{
		{
			name: "test map data should carry the session state of a logged user",
			session: &domain.Session{
				Username:  "someUser",
				CSRFToken: "someToken",
				Language:  "es",
				UserID:    7,
				AlertMessages: domain.AlertMessages{
					domain.NewErrorAlertMessage(http.StatusForbidden, "someTitle", "someMessage"),
				},
			},
			expected: &dto.ViewResponse[string]{
				Language:  "es",
				CSRFToken: "someToken",
				User:      "someUser",
				Msgs: []dto.AlertMessage{
					{
						Code:    http.StatusForbidden,
						Title:   "someTitle",
						Message: "someMessage",
						Type:    "error",
					},
				},
				Breadcrumbs: []string{"someBreadcrumb"},
				IsLogged:    true,
			},
		},
		{
			name: "test map data should report an anonymous session as not logged",
			session: &domain.Session{
				Language: "en",
				UserID:   0,
			},
			expected: &dto.ViewResponse[string]{
				Language:    "en",
				Msgs:        []dto.AlertMessage{},
				Breadcrumbs: []string{"someBreadcrumb"},
				IsLogged:    false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data := "someData"
			tt.expected.Data = &data

			actual := mapper.MapDataToViewResponse(&data, []string{"someBreadcrumb"}, tt.session)

			require.Equal(t, tt.expected, actual)
		})
	}
}
