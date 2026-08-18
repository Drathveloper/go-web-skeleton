package wire

import (
	securityservice "github.com/Drathveloper/go-web-skeleton/security/service"
	// scaffold:services:imports
)

type RequiredServices struct {
	AuthenticationService *securityservice.Authentication
	SessionService        *securityservice.Session
	UserManagementService *securityservice.UserManagement
	// scaffold:services:fields
}

func injectServices(container *Container) error {
	container.AuthenticationService = securityservice.NewAuthentication(container.UserRepository)
	container.SessionService = securityservice.NewSession(container.SessionsRepository)
	container.UserManagementService = securityservice.NewUserManagement(container.UserRepository)
	// scaffold:services:init
	return nil
}
