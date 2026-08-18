package dto

import "log/slog"

// LogoutEventName is the topic the logout event is published on. Subscribers in
// common/event/routes reference this constant instead of the bare string.
const LogoutEventName = "logout"

type LogoutEvent struct {
	Username  string
	IsSuccess bool
}

func (e LogoutEvent) GetName() string {
	return LogoutEventName
}

func (e LogoutEvent) LogAttrs() []slog.Attr {
	return []slog.Attr{
		slog.String("username", e.Username),
		slog.Bool("is_success", e.IsSuccess),
	}
}
