package constants

const (
	SessionGinContextKey = "session"
	CSRFTokenKey         = "csrf_token"
	SessionIDKey         = "sessionid"
)

// HomePath is where the application sends an authenticated user with no other
// destination: the redirect after a successful login and the fallback of the
// auth flows. It lives here so handlers never hardcode an application path.
const HomePath = "/"
