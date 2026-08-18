package dto_test

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Drathveloper/go-web-skeleton/pkg/event"
	"github.com/Drathveloper/go-web-skeleton/security/event/dto"
)

// Every security event must satisfy the bus contract and expose its fields as
// structured log attributes; a DTO that stops compiling here would otherwise
// only fail at runtime, as a silent hole in the audit trail.
var (
	_ event.Event = dto.LoginEvent{}
	_ event.Event = dto.LogoutEvent{}
	_ event.Event = dto.UserCreatedEvent{}
	_ event.Event = dto.UserUpdatedEvent{}

	_ interface{ LogAttrs() []slog.Attr } = dto.LoginEvent{}
	_ interface{ LogAttrs() []slog.Attr } = dto.LogoutEvent{}
	_ interface{ LogAttrs() []slog.Attr } = dto.UserCreatedEvent{}
	_ interface{ LogAttrs() []slog.Attr } = dto.UserUpdatedEvent{}
)

func TestEventDTOs_GetName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		evt  event.Event
		name string
		want string
	}{
		{
			name: "test login event should be named login",
			evt:  dto.LoginEvent{},
			want: "login",
		},
		{
			name: "test logout event should be named logout",
			evt:  dto.LogoutEvent{},
			want: "logout",
		},
		{
			name: "test user created event should be named user_created",
			evt:  dto.UserCreatedEvent{},
			want: "user_created",
		},
		{
			name: "test user updated event should be named user_updated",
			evt:  dto.UserUpdatedEvent{},
			want: "user_updated",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, tt.evt.GetName())
		})
	}
}

// AllEventNames feeds the audit logger subscriptions: it must contain the name
// of every DTO in this package, no more and no less.
func TestAllEventNames_ContainsEveryEventName(t *testing.T) {
	t.Parallel()

	want := []string{
		dto.LoginEvent{}.GetName(),
		dto.LogoutEvent{}.GetName(),
		dto.UserCreatedEvent{}.GetName(),
		dto.UserUpdatedEvent{}.GetName(),
	}

	require.Equal(t, want, dto.AllEventNames())
}
