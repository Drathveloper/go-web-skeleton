package routes

import (
	"github.com/Drathveloper/go-web-skeleton/common/wire"
	"github.com/Drathveloper/go-web-skeleton/security/event/dto"
)

func InitializeEventHandlers(container *wire.Container) error {
	for _, eventName := range dto.AllEventNames() {
		container.EventBus.Subscribe(eventName, container.AuditLogger.Handle)
	}
	// scaffold:event_routes:register
	return nil
}
