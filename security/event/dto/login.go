package dto

// LoginEventName is the topic the login event is published on. Subscribers in
// common/event/routes reference this constant instead of the bare string.
const LoginEventName = "login"

type LoginEvent struct {
	Username  string
	IsSuccess bool
}

func (e LoginEvent) GetName() string {
	return LoginEventName
}
