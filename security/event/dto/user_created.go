package dto

import "log/slog"

// UserCreatedEventName is the topic the user created event is published on.
// Subscribers in common/event/routes reference this constant instead of the
// bare string.
const UserCreatedEventName = "user_created"

type UserCreatedEvent struct {
	ActorUsername string
	Username      string
	Roles         []string
	IsSuccess     bool
}

func (e UserCreatedEvent) GetName() string {
	return UserCreatedEventName
}

func (e UserCreatedEvent) LogAttrs() []slog.Attr {
	return []slog.Attr{
		slog.String("actor_username", e.ActorUsername),
		slog.String("username", e.Username),
		slog.Any("roles", e.Roles),
		slog.Bool("is_success", e.IsSuccess),
	}
}
