package mapper_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	commondomain "github.com/Drathveloper/go-web-skeleton/common/domain"
	commondto "github.com/Drathveloper/go-web-skeleton/common/http/dto"
	"github.com/Drathveloper/go-web-skeleton/security/domain"
	"github.com/Drathveloper/go-web-skeleton/security/http/dto"
	"github.com/Drathveloper/go-web-skeleton/security/http/mapper"
)

func TestDomainUsersToUsersViewResponse(t *testing.T) {
	tests := []struct {
		session  *commondomain.Session
		expected *commondto.ViewResponse[dto.Users]
		name     string
		users    []domain.User
	}{
		{
			name:  "empty user list with logged session",
			users: []domain.User{},
			session: &commondomain.Session{
				CSRFToken:     "csrf-token-123",
				Username:      "admin",
				UserID:        1,
				AlertMessages: nil,
			},
			expected: &commondto.ViewResponse[dto.Users]{
				Data:        &dto.Users{Users: []dto.User{}},
				CSRFToken:   "csrf-token-123",
				User:        "admin",
				Msgs:        []commondto.AlertMessage{},
				Breadcrumbs: []string{"Security", "Users"},
				IsLogged:    true,
			},
		},
		{
			name:  "empty user list with anonymous session",
			users: []domain.User{},
			session: &commondomain.Session{
				CSRFToken:     "",
				Username:      "",
				UserID:        0,
				AlertMessages: nil,
			},
			expected: &commondto.ViewResponse[dto.Users]{
				Data:        &dto.Users{Users: []dto.User{}},
				CSRFToken:   "",
				User:        "",
				Msgs:        []commondto.AlertMessage{},
				Breadcrumbs: []string{"Security", "Users"},
				IsLogged:    false,
			},
		},
		{
			name: "single user without roles",
			users: []domain.User{
				{
					ID:       1,
					Username: "john_doe",
					Roles:    []commondomain.Role{},
				},
			},
			session: &commondomain.Session{
				CSRFToken:     "csrf-abc",
				Username:      "admin",
				UserID:        99,
				AlertMessages: nil,
			},
			expected: &commondto.ViewResponse[dto.Users]{
				Data: &dto.Users{
					Users: []dto.User{
						{ID: 1, Username: "john_doe", Roles: []string{}},
					},
				},
				CSRFToken:   "csrf-abc",
				User:        "admin",
				Msgs:        []commondto.AlertMessage{},
				Breadcrumbs: []string{"Security", "Users"},
				IsLogged:    true,
			},
		},
		{
			name: "single user with multiple roles",
			users: []domain.User{
				{
					ID:       2,
					Username: "jane_doe",
					Roles:    []commondomain.Role{commondomain.AdminRole, commondomain.UserRole},
				},
			},
			session: &commondomain.Session{
				CSRFToken:     "csrf-xyz",
				Username:      "superadmin",
				UserID:        1,
				AlertMessages: nil,
			},
			expected: &commondto.ViewResponse[dto.Users]{
				Data: &dto.Users{
					Users: []dto.User{
						{ID: 2, Username: "jane_doe", Roles: []string{"admin", "user"}},
					},
				},
				CSRFToken:   "csrf-xyz",
				User:        "superadmin",
				Msgs:        []commondto.AlertMessage{},
				Breadcrumbs: []string{"Security", "Users"},
				IsLogged:    true,
			},
		},
		{
			name: "multiple users with roles",
			users: []domain.User{
				{
					ID:       1,
					Username: "alice",
					Roles:    []commondomain.Role{commondomain.AdminRole},
				},
				{
					ID:       2,
					Username: "bob",
					Roles:    []commondomain.Role{commondomain.UserRole, commondomain.AdminRole},
				},
				{
					ID:       3,
					Username: "charlie",
					Roles:    []commondomain.Role{},
				},
			},
			session: &commondomain.Session{
				CSRFToken:     "csrf-multi",
				Username:      "admin",
				UserID:        10,
				AlertMessages: nil,
			},
			expected: &commondto.ViewResponse[dto.Users]{
				Data: &dto.Users{
					Users: []dto.User{
						{ID: 1, Username: "alice", Roles: []string{"admin"}},
						{ID: 2, Username: "bob", Roles: []string{"user", "admin"}},
						{ID: 3, Username: "charlie", Roles: []string{}},
					},
				},
				CSRFToken:   "csrf-multi",
				User:        "admin",
				Msgs:        []commondto.AlertMessage{},
				Breadcrumbs: []string{"Security", "Users"},
				IsLogged:    true,
			},
		},
		{
			name:  "session with alert messages",
			users: []domain.User{},
			session: &commondomain.Session{
				CSRFToken: "csrf-msgs",
				Username:  "admin",
				UserID:    5,
				AlertMessages: []commondomain.AlertMessage{
					{Type: "success", Message: "User created successfully"},
					{Type: "error", Message: "Something went wrong"},
				},
			},
			expected: &commondto.ViewResponse[dto.Users]{
				Data:      &dto.Users{Users: []dto.User{}},
				CSRFToken: "csrf-msgs",
				User:      "admin",
				Msgs: []commondto.AlertMessage{
					{Type: "success", Message: "User created successfully"},
					{Type: "error", Message: "Something went wrong"},
				},
				Breadcrumbs: []string{"Security", "Users"},
				IsLogged:    true,
			},
		},
		{
			name: "user with nil roles is mapped as empty slice",
			users: []domain.User{
				{
					ID:       7,
					Username: "nilroles",
					Roles:    nil,
				},
			},
			session: &commondomain.Session{
				CSRFToken: "csrf-nil",
				Username:  "admin",
				UserID:    1,
			},
			expected: &commondto.ViewResponse[dto.Users]{
				Data: &dto.Users{
					Users: []dto.User{
						{ID: 7, Username: "nilroles", Roles: []string{}},
					},
				},
				CSRFToken:   "csrf-nil",
				User:        "admin",
				Msgs:        []commondto.AlertMessage{},
				Breadcrumbs: []string{"Security", "Users"},
				IsLogged:    true,
			},
		},
		{
			name:  "the session language reaches the view",
			users: []domain.User{},
			session: &commondomain.Session{
				CSRFToken: "csrf-lang",
				Username:  "admin",
				Language:  "es",
				UserID:    1,
			},
			expected: &commondto.ViewResponse[dto.Users]{
				Data:        &dto.Users{Users: []dto.User{}},
				Language:    "es",
				CSRFToken:   "csrf-lang",
				User:        "admin",
				Msgs:        []commondto.AlertMessage{},
				Breadcrumbs: []string{"Security", "Users"},
				IsLogged:    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.DomainUsersToUsersViewResponse(tt.users, tt.session)

			require.NotNil(t, result)
			require.NotNil(t, result.Data)
			require.Equal(t, tt.expected, result)
			// The list view must never carry password hashes: the DTO has no field
			// for one, and this is what keeps it that way.
			requireNoPasswordField(t)
		})
	}
}

func TestSessionDomainToCreateUserViewResponse(t *testing.T) {
	session := &commondomain.Session{
		Username:  "someUsername",
		CSRFToken: "someCsrfToken",
	}
	expected := &commondto.ViewResponse[dto.CreateUserViewResponse]{
		Data: &dto.CreateUserViewResponse{
			// The form offers exactly the roles the project declares, so a role
			// removed from common/domain disappears from the form with it.
			AllowedRoles: commondomain.GetAllowedRoles(),
		},
		CSRFToken:   "someCsrfToken",
		User:        "someUsername",
		Msgs:        []commondto.AlertMessage{},
		Breadcrumbs: []string{"Security", "Users"},
		IsLogged:    false,
	}

	actual := mapper.SessionDomainToCreateUserViewResponse(session)

	require.Equal(t, expected, actual)
	require.Equal(t, []string{"admin", "user"}, actual.Data.AllowedRoles)
}

func TestDTOCreateUserProcessRequestToDomainUser(t *testing.T) {
	request := &dto.CreateUserProcessRequest{
		Username:        "someUsername",
		Password:        "1234",
		ConfirmPassword: "1234",
		Roles:           []string{"admin"},
	}
	expected := &domain.User{
		Username: "someUsername",
		Password: "1234",
		Roles:    []commondomain.Role{commondomain.AdminRole},
	}

	actual := mapper.DTOCreateUserProcessRequestToDomainUser(request)

	require.Equal(t, expected, actual)
}

func TestDTOCreateUserProcessRequestToDomainUser_WithoutRoles(t *testing.T) {
	request := &dto.CreateUserProcessRequest{
		Username:        "someUsername",
		Password:        "1234",
		ConfirmPassword: "1234",
		Roles:           nil,
	}
	expected := &domain.User{
		Username: "someUsername",
		Password: "1234",
		Roles:    []commondomain.Role{},
	}

	actual := mapper.DTOCreateUserProcessRequestToDomainUser(request)

	require.Equal(t, expected, actual)
}

func TestDTOUpdateUserProcessRequestToDomainUser(t *testing.T) {
	userID := uint(1)
	request := &dto.UpdateUserProcessRequest{
		Username:        "someUsername",
		Password:        "1234",
		ConfirmPassword: "1234",
		Roles:           []string{"admin"},
	}
	expected := &domain.User{
		Username: "someUsername",
		Password: "1234",
		Roles:    []commondomain.Role{commondomain.AdminRole},
		ID:       1,
	}

	actual := mapper.DTOUpdateUserProcessRequestToDomainUser(userID, request)

	require.Equal(t, expected, actual)
}

// TestDTOUpdateUserProcessRequestToDomainUser_BlankPassword: the edit form leaves the
// password blank when it is not being changed, and the mapper must pass that through
// untouched for the service to recognise it.
func TestDTOUpdateUserProcessRequestToDomainUser_BlankPassword(t *testing.T) {
	request := &dto.UpdateUserProcessRequest{
		Username:        "someUsername",
		Password:        "",
		ConfirmPassword: "",
		Roles:           []string{"user"},
	}
	expected := &domain.User{
		Username: "someUsername",
		Password: "",
		Roles:    []commondomain.Role{commondomain.UserRole},
		ID:       9,
	}

	actual := mapper.DTOUpdateUserProcessRequestToDomainUser(9, request)

	require.Equal(t, expected, actual)
}

func TestDomainUserToUpdateUserViewResponse(t *testing.T) {
	user := &domain.User{
		Username: "otherUser",
		Password: "1234",
		Roles:    []commondomain.Role{commondomain.AdminRole},
		ID:       1,
	}
	session := &commondomain.Session{
		ID:        "123",
		Username:  "someUsername",
		CSRFToken: "someCsrfToken",
		Roles:     []commondomain.Role{commondomain.AdminRole},
		AlertMessages: []commondomain.AlertMessage{
			{
				Title:   "someTitle",
				Message: "someMessage",
				Type:    "error",
				Code:    1,
			},
		},
		UserID: 1,
	}
	expected := &commondto.ViewResponse[dto.UpdateUserViewResponse]{
		Data: &dto.UpdateUserViewResponse{
			User: dto.User{
				ID:       1,
				Username: "otherUser",
				Roles:    []string{"admin"},
			},
			AllowedRoles: commondomain.GetAllowedRoles(),
		},
		CSRFToken: "someCsrfToken",
		User:      "someUsername",
		Msgs: []commondto.AlertMessage{
			{
				Title:   "someTitle",
				Message: "someMessage",
				Type:    "error",
				Code:    1,
			},
		},
		Breadcrumbs: []string{"Security", "Users"},
		IsLogged:    true,
	}

	actual := mapper.DomainUserToUpdateUserViewResponse(user, session)

	require.Equal(t, expected, actual)
	// The edit form must not be pre-filled with the stored hash: dto.User has no
	// field to put it in, and that is the point.
	requireNoPasswordField(t)
}

func TestDomainUserToDTOUser(t *testing.T) {
	tests := []struct {
		user     *domain.User
		expected *dto.User
		name     string
	}{
		{
			name: "user with roles",
			user: &domain.User{
				ID:       3,
				Username: "someUser",
				Password: "$argon2id$v=19$m=65536,t=1,p=1$c2FsdA$aGFzaA",
				Roles:    []commondomain.Role{commondomain.AdminRole, commondomain.UserRole},
			},
			expected: &dto.User{
				ID:       3,
				Username: "someUser",
				Roles:    []string{"admin", "user"},
			},
		},
		{
			name: "user with nil roles",
			user: &domain.User{
				ID:       4,
				Username: "noRoles",
				Roles:    nil,
			},
			expected: &dto.User{
				ID:       4,
				Username: "noRoles",
				Roles:    []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := mapper.DomainUserToDTOUser(tt.user)

			require.Equal(t, tt.expected, actual)
		})
	}
}

// requireNoPasswordField asserts the shape of the DTO rather than its values: the user
// rows are rendered straight into HTML, so a password field added here would be published.
func requireNoPasswordField(t *testing.T) {
	t.Helper()

	_, found := reflect.TypeFor[dto.User]().FieldByName("Password")
	require.False(t, found, "dto.User must not expose a password field")
}
