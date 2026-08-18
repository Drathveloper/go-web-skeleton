package mapper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/security/repository/redis/entity"
	"github.com/Drathveloper/go-web-skeleton/security/repository/redis/mapper"
)

func TestSessionEntityToSessionDomain(t *testing.T) {
	tests := []struct {
		input    *entity.Session
		expected *domain.Session
		name     string
		id       string
	}{
		{
			name: "complete session with roles and alert messages",
			id:   "session-123",
			input: &entity.Session{
				UserID:    123,
				Username:  "john.doe",
				Roles:     []string{"admin", "editor"},
				CSRFToken: "csrf-token-abc",
				AlertMessages: []entity.AlertMessages{
					{Code: 123, Title: "Error", Message: "Something went wrong", Type: "error"},
					{Code: 456, Title: "Info", Message: "Welcome back", Type: "info"},
				},
			},
			expected: &domain.Session{
				ID:        "session-123",
				UserID:    123,
				Username:  "john.doe",
				Roles:     []domain.Role{"admin", "editor"},
				CSRFToken: "csrf-token-abc",
				AlertMessages: domain.AlertMessages{
					{Code: 123, Title: "Error", Message: "Something went wrong", Type: "error"},
					{Code: 456, Title: "Info", Message: "Welcome back", Type: "info"},
				},
			},
		},
		{
			name: "session with empty roles and no alert messages",
			id:   "session-456",
			input: &entity.Session{
				UserID:        333,
				Username:      "jane.doe",
				Roles:         []string{},
				CSRFToken:     "csrf-token-xyz",
				AlertMessages: []entity.AlertMessages{},
			},
			expected: &domain.Session{
				ID:            "session-456",
				UserID:        333,
				Username:      "jane.doe",
				Roles:         []domain.Role{},
				CSRFToken:     "csrf-token-xyz",
				AlertMessages: domain.AlertMessages{},
			},
		},
		{
			name: "session with empty id",
			id:   "",
			input: &entity.Session{
				UserID:        1,
				Username:      "anonymous",
				Roles:         []string{"viewer"},
				CSRFToken:     "csrf-token-000",
				AlertMessages: nil,
			},
			expected: &domain.Session{
				ID:            "",
				UserID:        1,
				Username:      "anonymous",
				Roles:         []domain.Role{"viewer"},
				CSRFToken:     "csrf-token-000",
				AlertMessages: domain.AlertMessages{},
			},
		},
		{
			name: "session with nil roles and nil alert messages",
			id:   "session-nil",
			input: &entity.Session{
				UserID:        0,
				Username:      "nil.user",
				Roles:         nil,
				CSRFToken:     "",
				AlertMessages: nil,
			},
			expected: &domain.Session{
				ID:            "session-nil",
				UserID:        0,
				Username:      "nil.user",
				Roles:         []domain.Role{},
				CSRFToken:     "",
				AlertMessages: domain.AlertMessages{},
			},
		},
		// The language pair used to be write-only: persisted by
		// SessionDomainToSessionEntity and dropped on the way back, so the
		// language the user picked was lost on the next request.
		{
			name: "language chosen by the user is read back",
			id:   "session-lang",
			input: &entity.Session{
				UserID:               9,
				Username:             "lang.user",
				Roles:                []string{"admin"},
				CSRFToken:            "csrf-lang",
				Language:             "es",
				IsLanguageOverridden: true,
			},
			expected: &domain.Session{
				ID:                   "session-lang",
				UserID:               9,
				Username:             "lang.user",
				Roles:                []domain.Role{"admin"},
				CSRFToken:            "csrf-lang",
				Language:             "es",
				IsLanguageOverridden: true,
				AlertMessages:        domain.AlertMessages{},
			},
		},
		{
			name: "language negotiated from the request is read back as not overridden",
			id:   "session-lang-auto",
			input: &entity.Session{
				UserID:               10,
				Username:             "auto.user",
				Roles:                []string{},
				CSRFToken:            "csrf-auto",
				Language:             "en",
				IsLanguageOverridden: false,
			},
			expected: &domain.Session{
				ID:                   "session-lang-auto",
				UserID:               10,
				Username:             "auto.user",
				Roles:                []domain.Role{},
				CSRFToken:            "csrf-auto",
				Language:             "en",
				IsLanguageOverridden: false,
				AlertMessages:        domain.AlertMessages{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.SessionEntityToSessionDomain(tt.id, tt.input)

			require.NotNil(t, result)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestSessionDomainToSessionEntity(t *testing.T) {
	tests := []struct {
		input    *domain.Session
		expected *entity.Session
		name     string
	}{
		{
			name: "complete domain session with roles and alert messages",
			input: &domain.Session{
				ID:        "session-123",
				UserID:    343,
				Username:  "john.doe",
				Roles:     []domain.Role{"admin", "editor"},
				CSRFToken: "csrf-token-abc",
				AlertMessages: domain.AlertMessages{
					{Code: 12, Title: "Error", Message: "Something went wrong", Type: "error"},
					{Code: 13, Title: "Warning", Message: "Be careful", Type: "warning"},
				},
			},
			expected: &entity.Session{
				UserID:    343,
				Username:  "john.doe",
				Roles:     []string{"admin", "editor"},
				CSRFToken: "csrf-token-abc",
				AlertMessages: []entity.AlertMessages{
					{Code: 12, Title: "Error", Message: "Something went wrong", Type: "error"},
					{Code: 13, Title: "Warning", Message: "Be careful", Type: "warning"},
				},
			},
		},
		{
			name: "domain session with empty roles and no alert messages",
			input: &domain.Session{
				ID:            "session-789",
				UserID:        333,
				Username:      "empty.roles",
				Roles:         []domain.Role{},
				CSRFToken:     "csrf-token-789",
				AlertMessages: domain.AlertMessages{},
			},
			expected: &entity.Session{
				UserID:        333,
				Username:      "empty.roles",
				Roles:         []string{},
				CSRFToken:     "csrf-token-789",
				AlertMessages: []entity.AlertMessages{},
			},
		},
		{
			name: "domain session with nil roles and nil alert messages",
			input: &domain.Session{
				ID:            "session-nil",
				UserID:        0,
				Username:      "nil.fields",
				Roles:         nil,
				CSRFToken:     "",
				AlertMessages: nil,
			},
			expected: &entity.Session{
				UserID:        0,
				Username:      "nil.fields",
				Roles:         []string{},
				CSRFToken:     "",
				AlertMessages: []entity.AlertMessages{},
			},
		},
		{
			name: "domain session ID is not mapped to entity",
			input: &domain.Session{
				ID:        "should-not-appear",
				UserID:    11,
				Username:  "check.user",
				Roles:     []domain.Role{"viewer"},
				CSRFToken: "csrf-check",
			},
			expected: &entity.Session{
				UserID:        11,
				Username:      "check.user",
				Roles:         []string{"viewer"},
				CSRFToken:     "csrf-check",
				AlertMessages: []entity.AlertMessages{},
			},
		},
		// IsLanguageOverridden is persisted as given, not derived from a
		// non-empty Language: deciding whether the language is a user override
		// belongs to the middleware.
		{
			name: "a language that is not an override is persisted as not overridden",
			input: &domain.Session{
				ID:                   "session-lang",
				UserID:               2,
				Username:             "lang.user",
				Roles:                []domain.Role{"admin"},
				CSRFToken:            "csrf-lang",
				Language:             "es",
				IsLanguageOverridden: false,
			},
			expected: &entity.Session{
				UserID:               2,
				Username:             "lang.user",
				Roles:                []string{"admin"},
				CSRFToken:            "csrf-lang",
				Language:             "es",
				IsLanguageOverridden: false,
				AlertMessages:        []entity.AlertMessages{},
			},
		},
		{
			name: "an explicit language override is persisted as overridden",
			input: &domain.Session{
				ID:                   "session-lang-override",
				UserID:               3,
				Username:             "override.user",
				Roles:                []domain.Role{},
				CSRFToken:            "csrf-override",
				Language:             "en",
				IsLanguageOverridden: true,
			},
			expected: &entity.Session{
				UserID:               3,
				Username:             "override.user",
				Roles:                []string{},
				CSRFToken:            "csrf-override",
				Language:             "en",
				IsLanguageOverridden: true,
				AlertMessages:        []entity.AlertMessages{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.SessionDomainToSessionEntity(tt.input)

			require.NotNil(t, result)
			require.Equal(t, tt.expected, result)
		})
	}
}

// TestSessionRoundTrip pins what the store must preserve: everything the domain
// carries except the ID, which is the key and never travels in the value.
func TestSessionRoundTrip(t *testing.T) {
	original := &domain.Session{
		ID:                   "round-trip",
		UserID:               77,
		Username:             "round.trip",
		Roles:                []domain.Role{"admin", "user"},
		CSRFToken:            "csrf-round",
		Language:             "es",
		IsLanguageOverridden: true,
		AlertMessages: domain.AlertMessages{
			{Code: 1, Title: "Title", Message: "Message", Type: "info"},
		},
	}

	restored := mapper.SessionEntityToSessionDomain(original.ID, mapper.SessionDomainToSessionEntity(original))

	require.Equal(t, original, restored)
}
