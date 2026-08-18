package dto

type User struct {
	Username string
	Roles    []string
	ID       uint
}

type Users struct {
	Users []User
}

type UpdateUserViewResponse struct {
	AllowedRoles []string
	User         User
}

type CreateUserViewResponse struct {
	AllowedRoles []string
}

// CreateUserProcessRequest carries the new user form. The rules live in `binding`
// tags and nothing else: gin's binder reads that tag only, so a rule written under
// `validate` would never run. `isvalidrole` is registered from
// common/domain.GetAllowedRoles, and `eqfield` is what actually makes the confirm
// password field mean something.
type CreateUserProcessRequest struct {
	Username        string   `binding:"required"                  form:"username"`
	Password        string   `binding:"required"                  form:"password"`
	ConfirmPassword string   `binding:"required,eqfield=Password" form:"confirm_password"`
	Roles           []string `binding:"dive,isvalidrole"          form:"roles"`
}

// UpdateUserProcessRequest carries the edit user form. Password is optional here:
// an empty pair leaves the stored credential untouched, and a non-empty one still
// has to match its confirmation.
type UpdateUserProcessRequest struct {
	Username        string   `binding:"required"         form:"username"`
	Password        string   `form:"password"`
	ConfirmPassword string   `binding:"eqfield=Password" form:"confirm_password"`
	Roles           []string `binding:"dive,isvalidrole" form:"roles"`
}
