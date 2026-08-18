package wire

import (
	"fmt"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
	"github.com/Drathveloper/go-web-skeleton/pkg/event"
)

const injectEventClientsBaseErrMsg = "inject event clients failed"

type RequiredEventClients struct {
	EventBus *event.Bus
}

func injectEventClients(container *Container) error {
	opts := container.Store.GetEventBusConfig()
	eventBus, err := event.NewBus(opts)
	if err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, injectEventClientsBaseErrMsg, err)
	}
	container.EventBus = eventBus

	return nil
}
