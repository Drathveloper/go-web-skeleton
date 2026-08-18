package helper_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
	"github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/common/http/helper"
)

func TestGetSession(t *testing.T) {
	t.Parallel()

	tests := []struct {
		currentSession any
		name           string
		expectedErrMsg string
		wantErr        bool
	}{
		{
			name: "test get session should succeed",
			currentSession: &domain.Session{
				ID: "someID",
			},
			wantErr: false,
		},
		{
			name:           "test get session should return error when session is not present",
			currentSession: nil,
			wantErr:        true,
			expectedErrMsg: "get session failed: session not found",
		},
		{
			name:           "test get session should return error when session is not valid type",
			currentSession: "someID",
			wantErr:        true,
			expectedErrMsg: "get session failed: session not found",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			context, _ := gin.CreateTestContext(httptest.NewRecorder())
			if tt.currentSession != nil {
				context.Set(constants.SessionGinContextKey, tt.currentSession)
			}

			session, err := helper.GetSession(context)

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.expectedErrMsg, err.Error())
				require.ErrorIs(t, err, helper.ErrSessionNotFound)
				require.Nil(t, session)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.currentSession, session)
			}
		})
	}
}

// MustGetSession is called by every middleware that runs after SessionHandler.
// It panicking is deliberate: reaching it without a session means the router was
// wired wrong, and returning a zero session would hide that behind a page that
// renders as an anonymous user.
func TestMustGetSession(t *testing.T) {
	t.Parallel()

	t.Run("test must get session should return the session in the context", func(t *testing.T) {
		t.Parallel()

		expected := &domain.Session{ID: "someID"}
		context, _ := gin.CreateTestContext(httptest.NewRecorder())
		context.Set(constants.SessionGinContextKey, expected)

		require.Same(t, expected, helper.MustGetSession(context))
	})

	t.Run("test must get session should panic when there is no session", func(t *testing.T) {
		t.Parallel()

		context, _ := gin.CreateTestContext(httptest.NewRecorder())

		require.PanicsWithError(t, "get session failed: session not found", func() {
			helper.MustGetSession(context)
		})
	})
}

// A session ID is the whole authentication token: it has to carry the full 32
// bytes of randomness and never repeat.
func TestGenerateSessionID(t *testing.T) {
	t.Parallel()

	const (
		sessionIDLength = 43 // 32 bytes in base64 without padding
		samples         = 100
	)

	seen := make(map[string]struct{}, samples)
	for range samples {
		sessionID := helper.GenerateSessionID()

		require.Len(t, sessionID, sessionIDLength)
		require.NotContains(t, seen, sessionID, "generated session IDs must not repeat")
		seen[sessionID] = struct{}{}
	}
}
