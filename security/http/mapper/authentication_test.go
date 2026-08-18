package mapper_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	commondomain "github.com/Drathveloper/go-web-skeleton/common/domain"
	commondto "github.com/Drathveloper/go-web-skeleton/common/http/dto"
	"github.com/Drathveloper/go-web-skeleton/security/http/mapper"
)

func TestDomainSessionToLoginViewResponse(t *testing.T) {
	tests := []struct {
		session  *commondomain.Session
		expected *commondto.ViewResponse[any]
		name     string
	}{
		{
			name: "logged in user sets IsLogged true",
			session: &commondomain.Session{
				CSRFToken: "csrf-abc-123",
				Username:  "alice",
				UserID:    42,
			},
			expected: &commondto.ViewResponse[any]{
				CSRFToken:   "csrf-abc-123",
				User:        "alice",
				Msgs:        []commondto.AlertMessage{},
				Breadcrumbs: []string{"Security", "Login"},
				IsLogged:    true,
			},
		},
		{
			name: "guest user (UserID 0) sets IsLogged false",
			session: &commondomain.Session{
				CSRFToken: "csrf-xyz",
				UserID:    0,
			},
			expected: &commondto.ViewResponse[any]{
				CSRFToken:   "csrf-xyz",
				Msgs:        []commondto.AlertMessage{},
				Breadcrumbs: []string{"Security", "Login"},
				IsLogged:    false,
			},
		},
		{
			name: "alert messages are mapped correctly",
			session: &commondomain.Session{
				CSRFToken: "token",
				UserID:    1,
				AlertMessages: []commondomain.AlertMessage{
					{Type: "error", Message: "invalid credentials"},
					{Type: "info", Message: "session expired"},
				},
			},
			expected: &commondto.ViewResponse[any]{
				CSRFToken: "token",
				Msgs: []commondto.AlertMessage{
					{Type: "error", Message: "invalid credentials"},
					{Type: "info", Message: "session expired"},
				},
				Breadcrumbs: []string{"Security", "Login"},
				IsLogged:    true,
			},
		},
		{
			name: "empty alert messages slice produces zero msgs",
			session: &commondomain.Session{
				CSRFToken:     "token",
				UserID:        5,
				AlertMessages: []commondomain.AlertMessage{},
			},
			expected: &commondto.ViewResponse[any]{
				CSRFToken:   "token",
				Msgs:        []commondto.AlertMessage{},
				Breadcrumbs: []string{"Security", "Login"},
				IsLogged:    true,
			},
		},
		{
			name: "nil alert messages produces zero msgs",
			session: &commondomain.Session{
				CSRFToken:     "token",
				UserID:        3,
				AlertMessages: nil,
			},
			expected: &commondto.ViewResponse[any]{
				CSRFToken:   "token",
				Msgs:        []commondto.AlertMessage{},
				Breadcrumbs: []string{"Security", "Login"},
				IsLogged:    true,
			},
		},
		// The login screen is rendered before the user has authenticated, so the
		// language negotiated for the session is the only thing that localizes it.
		{
			name: "the session language reaches the view",
			session: &commondomain.Session{
				CSRFToken: "token",
				Language:  "es",
				UserID:    0,
			},
			expected: &commondto.ViewResponse[any]{
				Language:    "es",
				CSRFToken:   "token",
				Msgs:        []commondto.AlertMessage{},
				Breadcrumbs: []string{"Security", "Login"},
				IsLogged:    false,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := mapper.DomainSessionToLoginViewResponse(tc.session)

			require.NotNil(t, result)
			// The login view carries no data of its own: everything it renders comes
			// from the session, so a non-nil Data would be a payload nobody asked for.
			require.Nil(t, result.Data)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestFormLoginToDomainLogin(t *testing.T) {
	tests := []struct {
		name           string
		values         url.Values
		wantUsername   string
		wantPassword   string
		wantRememberMe bool
	}{
		{
			name: "all fields present",
			values: url.Values{
				"username":    {"alice"},
				"password":    {"s3cr3t"},
				"remember_me": {"on"},
			},
			wantUsername:   "alice",
			wantPassword:   "s3cr3t",
			wantRememberMe: true,
		},
		{
			name: "remember_me absent sets it false",
			values: url.Values{
				"username": {"bob"},
				"password": {"pass123"},
			},
			wantUsername:   "bob",
			wantPassword:   "pass123",
			wantRememberMe: false,
		},
		{
			name:           "empty form returns zero-value login without error",
			values:         url.Values{},
			wantUsername:   "",
			wantPassword:   "",
			wantRememberMe: false,
		},
		{
			name: "remember_me with empty string value sets it false",
			values: url.Values{
				"username":    {"carol"},
				"password":    {"pwd"},
				"remember_me": {""},
			},
			wantUsername:   "carol",
			wantPassword:   "pwd",
			wantRememberMe: false,
		},
		{
			name: "username and password with special characters",
			values: url.Values{
				"username": {"user@domain.com"},
				"password": {"p@$$w0rd!"},
			},
			wantUsername:   "user@domain.com",
			wantPassword:   "p@$$w0rd!",
			wantRememberMe: false,
		},
		{
			name: "only the first value of a repeated field is taken",
			values: url.Values{
				"username": {"first", "second"},
				"password": {"pwd", "other"},
			},
			wantUsername:   "first",
			wantPassword:   "pwd",
			wantRememberMe: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			login, err := mapper.FormLoginToDomainLogin(tc.values)

			require.NoError(t, err)
			require.NotNil(t, login)
			require.Equal(t, tc.wantUsername, login.Username)
			require.Equal(t, tc.wantPassword, login.Password)
			require.Equal(t, tc.wantRememberMe, login.RememberMe)
		})
	}
}
