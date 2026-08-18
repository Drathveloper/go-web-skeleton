package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
	commondomain "github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/common/http/helper"
	"github.com/Drathveloper/go-web-skeleton/common/http/templates"
	"github.com/Drathveloper/go-web-skeleton/security/domain"
	eventdto "github.com/Drathveloper/go-web-skeleton/security/event/dto"
	"github.com/Drathveloper/go-web-skeleton/security/http/handler"
)

const actorUsername = "admin"

type stubUserManagementService struct {
	createErr error
	updateErr error
}

func (s *stubUserManagementService) ListUsers(
	_ context.Context, _ *commondomain.Pagination) ([]domain.User, error) {
	return nil, nil
}

func (s *stubUserManagementService) GetUserByID(_ context.Context, _ uint) (*domain.User, error) {
	return &domain.User{}, nil
}

func (s *stubUserManagementService) CreateUser(_ context.Context, _ *domain.User) error {
	return s.createErr
}

func (s *stubUserManagementService) UpdateUser(_ context.Context, _ *domain.User) error {
	return s.updateErr
}

func newUserManagementEngine(t *testing.T, ctrl *handler.UserManagement) *gin.Engine {
	t.Helper()

	engine := gin.New()
	require.NoError(t, templates.InitializeTemplateRenderer(engine))
	engine.Use(func(c *gin.Context) {
		c.Set(constants.SessionGinContextKey, &commondomain.Session{
			ID: "session-id", Username: actorUsername, Language: "en", CSRFToken: "csrf", UserID: 1,
		})
	})
	engine.POST("/auth/user/new", ctrl.CreateUserProcess())
	engine.POST("/auth/user/:id", ctrl.UpdateUserProcess())

	return engine
}

func postForm(t *testing.T, engine *gin.Engine, path string, form url.Values, htmx bool) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequestWithContext(
		t.Context(), http.MethodPost, path, strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if htmx {
		request.Header.Set(helper.HXRequestHeader, "true")
	}
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	return recorder
}

func createUserForm() url.Values {
	return url.Values{
		"username":         {"newuser"},
		"password":         {"secret"},
		"confirm_password": {"secret"},
		"roles":            {"user"},
	}
}

func TestCreateUserProcess_PublishesUserCreatedEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		createErr    error
		name         string
		wantLocation string
		wantStatus   int
		htmx         bool
		wantSuccess  bool
	}{
		{
			name:         "test create user should publish a successful event",
			wantStatus:   http.StatusFound,
			wantLocation: "/auth/user",
			wantSuccess:  true,
		},
		{
			name:        "test create user should publish a successful event on the htmx flow",
			htmx:        true,
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:         "test create user should publish a failed event when the service fails",
			createErr:    errors.New("create user failed"),
			wantStatus:   http.StatusFound,
			wantLocation: "/auth/user/new",
			wantSuccess:  false,
		},
		{
			name:        "test create user should publish a failed event on the htmx flow when the service fails",
			createErr:   errors.New("create user failed"),
			htmx:        true,
			wantStatus:  http.StatusUnprocessableEntity,
			wantSuccess: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			publisher := &publisherSpy{}
			ctrl := handler.NewUserManagement(&stubUserManagementService{createErr: tt.createErr}, publisher)
			engine := newUserManagementEngine(t, ctrl)

			recorder := postForm(t, engine, "/auth/user/new", createUserForm(), tt.htmx)

			require.Equal(t, tt.wantStatus, recorder.Code)
			if tt.wantLocation != "" {
				require.Equal(t, tt.wantLocation, recorder.Header().Get("Location"))
			}
			require.Len(t, publisher.events, 1)
			require.Equal(t, &eventdto.UserCreatedEvent{
				ActorUsername: actorUsername,
				Username:      "newuser",
				Roles:         []string{"user"},
				IsSuccess:     tt.wantSuccess,
			}, publisher.events[0])
		})
	}
}

// A form that never reaches the service is not a user creation attempt: binding
// failures must stay out of the audit trail.
func TestCreateUserProcess_DoesNotPublishOnBindingError(t *testing.T) {
	t.Parallel()

	publisher := &publisherSpy{}
	ctrl := handler.NewUserManagement(&stubUserManagementService{}, publisher)
	engine := newUserManagementEngine(t, ctrl)

	form := createUserForm()
	form.Del("password")
	recorder := postForm(t, engine, "/auth/user/new", form, false)

	require.Equal(t, http.StatusFound, recorder.Code)
	require.Empty(t, publisher.events)
}

func updateUserForm() url.Values {
	return url.Values{
		"username": {"edited"},
		"roles":    {"admin"},
	}
}

func TestUpdateUserProcess_PublishesUserUpdatedEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		updateErr    error
		name         string
		wantLocation string
		wantStatus   int
		htmx         bool
		wantSuccess  bool
	}{
		{
			name:         "test update user should publish a successful event",
			wantStatus:   http.StatusFound,
			wantLocation: "/auth/user",
			wantSuccess:  true,
		},
		{
			name:        "test update user should publish a successful event on the htmx flow",
			htmx:        true,
			wantStatus:  http.StatusOK,
			wantSuccess: true,
		},
		{
			name:         "test update user should publish a failed event when the service fails",
			updateErr:    errors.New("update user failed"),
			wantStatus:   http.StatusFound,
			wantLocation: "/auth/user/7/edit",
			wantSuccess:  false,
		},
		{
			name:        "test update user should publish a failed event on the htmx flow when the service fails",
			updateErr:   errors.New("update user failed"),
			htmx:        true,
			wantStatus:  http.StatusUnprocessableEntity,
			wantSuccess: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			publisher := &publisherSpy{}
			ctrl := handler.NewUserManagement(&stubUserManagementService{updateErr: tt.updateErr}, publisher)
			engine := newUserManagementEngine(t, ctrl)

			recorder := postForm(t, engine, "/auth/user/7", updateUserForm(), tt.htmx)

			require.Equal(t, tt.wantStatus, recorder.Code)
			if tt.wantLocation != "" {
				require.Equal(t, tt.wantLocation, recorder.Header().Get("Location"))
			}
			require.Len(t, publisher.events, 1)
			require.Equal(t, &eventdto.UserUpdatedEvent{
				ActorUsername: actorUsername,
				Username:      "edited",
				Roles:         []string{"admin"},
				UserID:        7,
				IsSuccess:     tt.wantSuccess,
			}, publisher.events[0])
		})
	}
}

// Neither a malformed user id nor a form that fails binding reaches the
// service, so neither may leave a trace in the audit trail.
func TestUpdateUserProcess_DoesNotPublishBeforeTheServiceCall(t *testing.T) {
	t.Parallel()

	tests := []struct {
		form url.Values
		name string
		path string
	}{
		{
			name: "test update user should not publish on a malformed user id",
			path: "/auth/user/not-a-number",
			form: updateUserForm(),
		},
		{
			name: "test update user should not publish on a binding error",
			path: "/auth/user/7",
			form: url.Values{"roles": {"admin"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			publisher := &publisherSpy{}
			ctrl := handler.NewUserManagement(&stubUserManagementService{}, publisher)
			engine := newUserManagementEngine(t, ctrl)

			recorder := postForm(t, engine, tt.path, tt.form, false)

			require.Equal(t, http.StatusFound, recorder.Code)
			require.Empty(t, publisher.events)
		})
	}
}
