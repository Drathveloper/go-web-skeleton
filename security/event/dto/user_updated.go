package dto

import "log/slog"

// UserUpdatedEventName is the topic the user updated event is published on.
// Subscribers in common/event/routes reference this constant instead of the
// bare string.
const UserUpdatedEventName = "user_updated"

type UserUpdatedEvent struct {
	ActorUsername string
	Username      string
	Roles         []string
	UserID        uint
	IsSuccess     bool
}

func (e UserUpdatedEvent) GetName() string {
	return UserUpdatedEventName
}

func (e UserUpdatedEvent) LogAttrs() []slog.Attr {
	return []slog.Attr{
		slog.String("actor_username", e.ActorUsername),
		slog.String("username", e.Username),
		slog.Any("roles", e.Roles),
		slog.Uint64("user_id", uint64(e.UserID)),
		slog.Bool("is_success", e.IsSuccess),
	}
}
