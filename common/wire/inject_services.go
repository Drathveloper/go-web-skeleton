package wire

import (
	exampleservice "github.com/Drathveloper/go-web-skeleton/example/service"
	securityservice "github.com/Drathveloper/go-web-skeleton/security/service"
	// scaffold:services:imports
)

type RequiredServices struct {
	AuthenticationService *securityservice.Authentication
	SessionService        *securityservice.Session
	UserManagementService *securityservice.UserManagement
	ItemCategoryService   *exampleservice.ItemCategory
	ItemService           *exampleservice.Item
	// scaffold:services:fields
}

func injectServices(container *Container) error {
	container.AuthenticationService = securityservice.NewAuthentication(container.UserRepository)
	container.SessionService = securityservice.NewSession(container.SessionsRepository)
	container.UserManagementService = securityservice.NewUserManagement(container.UserRepository)
	container.ItemCategoryService = exampleservice.NewItemCategory(container.ItemCategoryRepository)
	container.ItemService = exampleservice.NewItem(container.ItemRepository)
	// scaffold:services:init
	return nil
}
