package event

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
	"github.com/Drathveloper/go-web-skeleton/common/wire"
)

const runBaseErrMsg = "run event listeners failed"

// HandlerRegistrar subscribes the event handlers of a module to the container event
// bus. Modules pass their own registrar to Run instead of common/event importing them,
// so this package stays free of any dependency on the application modules.
type HandlerRegistrar func(container *wire.Container) error

// Run registers the given event handlers and starts the goroutine that shuts the event
// bus down when sigChan fires, bounded by the configured event shutdown timeout.
func Run(
	container *wire.Container,
	waitGroup *sync.WaitGroup,
	sigChan chan os.Signal,
	registrars ...HandlerRegistrar,
) error {
	for _, registerHandlers := range registrars {
		if err := registerHandlers(container); err != nil {
			return fmt.Errorf(constants.DefaultWrappedErrorTemplate, runBaseErrMsg, err)
		}
	}
	waitGroup.Go(func() {
		<-sigChan
		slog.Debug("shutting down event bus")
		ctx, cancelFunc := context.WithTimeout(context.Background(), container.Store.GetEventShutdownTimeout())
		defer cancelFunc()
		if err := container.EventBus.Shutdown(ctx); err != nil {
			slog.Error("event bus failed to shutdown", slog.String("error", err.Error()))
		}
		slog.Debug("event bus shutdown successfully")
	})
	return nil
}
