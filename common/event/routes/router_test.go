package routes_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Drathveloper/go-web-skeleton/common/event/routes"
	"github.com/Drathveloper/go-web-skeleton/common/wire"
	"github.com/Drathveloper/go-web-skeleton/pkg/event"
	"github.com/Drathveloper/go-web-skeleton/security/event/dto"
	securityeventhandler "github.com/Drathveloper/go-web-skeleton/security/event/handler"
)

// Deliberately not parallel: it swaps the default slog logger, which is the
// only logger the bus workers can reach, to capture the audit lines.
func TestInitializeEventHandlers(t *testing.T) {
	bus, err := event.NewBus(event.NewDefaultOptions())
	require.NoError(t, err)

	container := &wire.Container{
		RequiredEventClients:  wire.RequiredEventClients{EventBus: bus},
		RequiredEventHandlers: wire.RequiredEventHandlers{AuditLogger: securityeventhandler.NewAuditLogger()},
	}

	require.NoError(t, routes.InitializeEventHandlers(container))

	buf := &bytes.Buffer{}
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(buf, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	events := []event.Event{
		dto.LoginEvent{Username: "someone", IsSuccess: true},
		dto.LogoutEvent{Username: "someone", IsSuccess: true},
		dto.UserCreatedEvent{ActorUsername: "someone", Username: "created", IsSuccess: true},
		dto.UserUpdatedEvent{ActorUsername: "someone", Username: "updated", UserID: 1, IsSuccess: false},
	}
	for _, evt := range events {
		bus.Publish(evt)
	}

	// Shutdown waits for the in-flight handlers, so after it returns every
	// published event has been through the audit logger.
	require.NoError(t, bus.Shutdown(context.Background()))

	output := buf.String()
	require.Contains(t, output, "security event")
	for _, eventName := range dto.AllEventNames() {
		require.Contains(t, output, "event="+eventName)
	}
}
