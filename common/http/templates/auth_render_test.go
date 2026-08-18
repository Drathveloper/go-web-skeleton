package templates_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	commondomain "github.com/Drathveloper/go-web-skeleton/common/domain"
	commondto "github.com/Drathveloper/go-web-skeleton/common/http/dto"
	securitydto "github.com/Drathveloper/go-web-skeleton/security/http/dto"
)

// The auth handlers name six templates. Each of them is executed here with the
// exact payload the handler builds, because a missing or broken one only shows
// up as a panic at request time.

func userFormData(user *securitydto.User, roles []string, isEdit bool, errMsg string) gin.H {
	var userRoles []string
	if user != nil {
		userRoles = user.Roles
	}
	return gin.H{
		"Language":     "en",
		"CSRFToken":    "csrf-token-value",
		"User":         user,
		"UserRoles":    userRoles,
		"AllowedRoles": roles,
		"IsEdit":       isEdit,
		"Error":        errMsg,
	}
}

func userRowData(user *securitydto.User, oob string) gin.H {
	return gin.H{
		"Language":  "en",
		"CSRFToken": "csrf-token-value",
		"User":      user,
		"OOB":       oob,
	}
}

func TestLoginPageRendersWithoutChrome(t *testing.T) {
	t.Parallel()
	body := render(t, newEngine(t), http.StatusOK, "auth/login", &commondto.ViewResponse[any]{
		Language:    "en",
		CSRFToken:   "csrf-token-value",
		Breadcrumbs: []string{"Security", "Login"},
		IsLogged:    false,
	})
	require.Contains(t, body, "<title>Login</title>")
	require.Contains(t, body, `<form method="post" action="/auth/login">`)
	require.Contains(t, body, `id="username"`)
	require.Contains(t, body, `type="password"`)
	require.Contains(t, body, `name="remember_me"`)
	require.Contains(t, body, `name="csrf_token" value="csrf-token-value"`)
	// An anonymous visitor gets no application chrome at all.
	require.NotContains(t, body, `class="sidebar"`)
	require.NotContains(t, body, `class="topbar"`)
	require.NotContains(t, body, "user-menu")
	require.NotContains(t, body, `id="modal"`)
	// A login form is a plain POST, not an HTMX swap.
	require.NotContains(t, body, "hx-post")
}

func TestListUsersPageRenders(t *testing.T) {
	t.Parallel()
	body := render(t, newEngine(t), http.StatusOK, "auth/list_users", &commondto.ViewResponse[securitydto.Users]{
		Language:    "en",
		CSRFToken:   "csrf-token-value",
		User:        "admin",
		IsLogged:    true,
		Breadcrumbs: []string{"Security", "Users"},
		Data: &securitydto.Users{Users: []securitydto.User{
			{ID: 1, Username: "alice", Roles: []string{"admin", "user"}},
			{ID: 2, Username: "bob", Roles: []string{"user"}},
		}},
	})
	require.Contains(t, body, `id="users-table"`)
	require.Contains(t, body, `id="users-tbody"`)
	require.Contains(t, body, `id="users-table-search"`)
	require.Contains(t, body, `<tr id="user-row-1">`)
	require.Contains(t, body, `<tr id="user-row-2">`)
	require.Contains(t, body, `hx-get="/auth/user/1/edit"`)
	require.Contains(t, body, `hx-get="/auth/user/new"`)
	require.Contains(t, body, `<span class="badge badge-brand">admin, user</span>`)
	require.Contains(t, body, "<tr data-empty-row hidden>")
	require.Contains(t, body, `class="sidebar"`)
	// There is no delete route for users, so there must be no delete button.
	require.NotContains(t, body, "hx-post=\"/auth/user/1/delete\"")
}

func TestListUsersPageRendersEmptyState(t *testing.T) {
	t.Parallel()
	body := render(t, newEngine(t), http.StatusOK, "auth/list_users", &commondto.ViewResponse[securitydto.Users]{
		Language: "en", IsLogged: true, Data: &securitydto.Users{},
	})
	require.Contains(t, body, "<tr data-empty-row>")
	require.Contains(t, body, "No users available")
}

func TestCreateUserPageRendersAsAPlainForm(t *testing.T) {
	t.Parallel()
	body := render(t, newEngine(t), http.StatusOK, "auth/create_user",
		&commondto.ViewResponse[securitydto.CreateUserViewResponse]{
			Language: "en", CSRFToken: "csrf-token-value", IsLogged: true,
			Data: &securitydto.CreateUserViewResponse{AllowedRoles: commondomain.GetAllowedRoles()},
		})
	require.Contains(t, body, `<form method="post" action="/auth/user/new">`)
	require.NotContains(t, body, "hx-post")
	require.NotContains(t, body, "close-modal")
	for _, role := range commondomain.GetAllowedRoles() {
		require.Contains(t, body, fmt.Sprintf(`name="roles" value="%s"`, role))
	}
	require.NotContains(t, body, "checked")
}

func TestUpdateUserPageRendersAsAPlainForm(t *testing.T) {
	t.Parallel()
	body := render(t, newEngine(t), http.StatusOK, "auth/update_user",
		&commondto.ViewResponse[securitydto.UpdateUserViewResponse]{
			Language: "en", CSRFToken: "csrf-token-value", IsLogged: true,
			Data: &securitydto.UpdateUserViewResponse{
				AllowedRoles: commondomain.GetAllowedRoles(),
				User:         securitydto.User{ID: 7, Username: "alice", Roles: []string{"admin"}},
			},
		})
	require.Contains(t, body, `<form method="post" action="/auth/user/7/edit">`)
	require.Contains(t, body, `value="alice"`)
	require.Contains(t, body, `name="roles" value="admin" class="check" checked`)
	require.Contains(t, body, "Leave blank to keep the current password.")
}

func TestUserFormFragmentRendersForCreate(t *testing.T) {
	t.Parallel()
	body := render(t, newEngine(t), http.StatusOK, "fragments/auth/form",
		userFormData(nil, commondomain.GetAllowedRoles(), false, ""))
	require.Contains(t, body, `hx-post="/auth/user/new"`)
	require.Contains(t, body, `hx-target="#modal"`)
	require.Contains(t, body, "Create user")
	require.NotContains(t, body, "checked")
	require.NotContains(t, body, `class="alert alert-danger"`)
}

func TestUserFormFragmentRendersValidationErrorInlineWith422(t *testing.T) {
	t.Parallel()
	user := &securitydto.User{ID: 7, Username: "alice", Roles: []string{"admin"}}
	body := render(t, newEngine(t), http.StatusUnprocessableEntity, "fragments/auth/form",
		userFormData(user, commondomain.GetAllowedRoles(), true, "The submitted data is not valid."))
	require.Contains(t, body, `hx-post="/auth/user/7/edit"`)
	require.Contains(t, body, `class="alert alert-danger"`)
	require.Contains(t, body, "The submitted data is not valid.")
	require.Contains(t, body, `value="alice"`, "a rejected submit must not empty the form")
	require.Contains(t, body, `name="roles" value="admin" class="check" checked`)
	require.Contains(t, body, `name="roles" value="user" class="check">`, "an unselected role stays unchecked")
	// A password is never echoed back into the form, not even after a 422.
	require.NotContains(t, body, `type="password" value=`)
}

func TestUserRowFragmentSpeaksTheHTMXContract(t *testing.T) {
	t.Parallel()
	user := &securitydto.User{ID: 7, Username: "alice", Roles: []string{"admin"}}

	created := render(t, newEngine(t), http.StatusOK, "fragments/auth/row",
		userRowData(user, "afterbegin:#users-tbody"))
	require.Contains(t, created, `<tbody hx-swap-oob="afterbegin:#users-tbody">`)
	require.Contains(t, created, `<tr id="user-row-7">`)
	require.Contains(t, created, `hx-get="/auth/user/7/edit"`)

	updated := render(t, newEngine(t), http.StatusOK, "fragments/auth/row", userRowData(user, "true"))
	require.Contains(t, updated, `<tr id="user-row-7" hx-swap-oob="true">`)
	require.NotContains(t, updated, "<tbody")
}

// The roles control must come entirely from AllowedRoles: `scaffold new --roles`
// rewrites common/domain/roles.go, and a role named in a template would survive
// that and quietly grant something that no longer exists.
func TestRolesControlIsDrivenOnlyByAllowedRoles(t *testing.T) {
	t.Parallel()
	body := render(t, newEngine(t), http.StatusOK, "fragments/auth/form",
		userFormData(&securitydto.User{ID: 1, Username: "alice", Roles: []string{"auditor"}},
			[]string{"auditor", "operator"}, true, ""))
	require.Contains(t, body, `name="roles" value="auditor" class="check" checked`)
	require.Contains(t, body, `name="roles" value="operator" class="check">`)
	for _, role := range commondomain.GetAllowedRoles() {
		require.NotContains(t, body, fmt.Sprintf(`name="roles" value="%s"`, role),
			"a role that is not in AllowedRoles must not appear")
	}
}
