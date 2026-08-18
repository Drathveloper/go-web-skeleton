package wire

import (
	securityhandler "github.com/Drathveloper/go-web-skeleton/security/http/handler"
	// scaffold:handlers:imports
)

type RequiredHTTPHandlers struct {
	AuthenticationHandler *securityhandler.Authentication
	UserManagementHandler *securityhandler.UserManagement
	// scaffold:handlers:fields
}

func injectHTTPHandlers(container *Container) error {
	container.AuthenticationHandler = securityhandler.NewAuthentication(
		container.AuthenticationService, container.SessionService, container.Store, container.EventBus)
	container.UserManagementHandler = securityhandler.NewUserManagement(container.UserManagementService)
	// scaffold:handlers:init
	return nil
}
