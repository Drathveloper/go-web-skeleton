package domain

type Role string

// The block below is the single source of truth for the roles of the project:
// `scaffold new --roles a,b,c` rewrites this file wholesale. Declare roles here
// and nowhere else — the middleware.Authorize arguments and the "isvalidrole"
// validator, fed by GetAllowedRoles, both derive from it.
const (
	AdminRole Role = "admin"
	UserRole  Role = "user"
)

var allowedRoles = []Role{AdminRole, UserRole} //nolint:gochecknoglobals

func GetAllowedRoles() []string {
	allowedRolesStr := make([]string, 0, len(allowedRoles))
	for _, r := range allowedRoles {
		allowedRolesStr = append(allowedRolesStr, string(r))
	}
	return allowedRolesStr
}
