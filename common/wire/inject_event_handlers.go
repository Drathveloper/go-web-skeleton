package wire

import (
	securityeventhandler "github.com/Drathveloper/go-web-skeleton/security/event/handler"
)

type RequiredEventHandlers struct {
	AuditLogger *securityeventhandler.AuditLogger
	// scaffold:event_handlers:fields
}

func injectEventHandlers(container *Container) error {
	container.AuditLogger = securityeventhandler.NewAuditLogger()
	// scaffold:event_handlers:init
	return nil
}
