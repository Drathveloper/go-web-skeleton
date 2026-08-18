package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
	commondomain "github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/security/domain"
	"github.com/Drathveloper/go-web-skeleton/security/event/dto"
	"github.com/Drathveloper/go-web-skeleton/security/http/handler"
)

// The logout tests never reach the login flow, so the stubs below only have to
// satisfy the constructor.
type stubAuthService struct{}

func (stubAuthService) Login(_ context.Context, _ *domain.Login) (*domain.User, error) {
	return &domain.User{}, nil
}

type stubSessionService struct {
	destroyErr error
}

func (s *stubSessionService) CreateUserSession(
	_ context.Context, _, _ string, _ *domain.User) (*commondomain.Session, error) {
	return &commondomain.Session{}, nil
}

func (s *stubSessionService) DestroySession(_ context.Context, _ *commondomain.Session) error {
	return s.destroyErr
}

type stubConfigProvider struct{}

func (stubConfigProvider) GetSessionTTL() time.Duration {
	return time.Hour
}

func (stubConfigProvider) IsSecureCookie() bool {
	return false
}

func (stubConfigProvider) GetCookieDomain() string {
	return ""
}

func TestLogoutProcess_PublishesLogoutEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		destroyErr   error
		name         string
		wantLocation string
		wantStatus   int
		wantSuccess  bool
	}{
		{
			name:         "test logout should publish a successful logout event",
			wantStatus:   http.StatusFound,
			wantLocation: "/auth/login",
			wantSuccess:  true,
		},
		{
			name:         "test logout should publish a failed logout event when destroy session fails",
			destroyErr:   errors.New("destroy session failed"),
			wantStatus:   http.StatusFound,
			wantLocation: "/",
			wantSuccess:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			publisher := &publisherSpy{}
			ctrl := handler.NewAuthentication(
				stubAuthService{}, &stubSessionService{destroyErr: tt.destroyErr}, stubConfigProvider{}, publisher)

			engine := gin.New()
			engine.POST("/auth/logout",
				func(c *gin.Context) {
					c.Set(constants.SessionGinContextKey,
						&commondomain.Session{ID: "session-id", Username: actorUsername, Language: "en", UserID: 1})
				},
				ctrl.LogoutProcess())

			recorder := httptest.NewRecorder()
			request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/auth/logout", nil)
			engine.ServeHTTP(recorder, request)

			require.Equal(t, tt.wantStatus, recorder.Code)
			require.Equal(t, tt.wantLocation, recorder.Header().Get("Location"))
			require.Len(t, publisher.events, 1)
			require.Equal(t, &dto.LogoutEvent{Username: actorUsername, IsSuccess: tt.wantSuccess}, publisher.events[0])
		})
	}
}
