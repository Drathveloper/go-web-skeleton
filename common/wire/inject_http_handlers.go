package wire

import (
	examplehandler "github.com/Drathveloper/go-web-skeleton/example/http/handler"
	securityhandler "github.com/Drathveloper/go-web-skeleton/security/http/handler"
	// scaffold:handlers:imports
)

type RequiredHTTPHandlers struct {
	AuthenticationHandler *securityhandler.Authentication
	UserManagementHandler *securityhandler.UserManagement
	ItemCategoryHandler   *examplehandler.ItemCategory
	ItemHandler           *examplehandler.Item
	// scaffold:handlers:fields
}

func injectHTTPHandlers(container *Container) error {
	container.AuthenticationHandler = securityhandler.NewAuthentication(
		container.AuthenticationService, container.SessionService, container.Store, container.EventBus)
	container.UserManagementHandler = securityhandler.NewUserManagement(container.UserManagementService)
	container.ItemCategoryHandler = examplehandler.NewItemCategory(container.ItemCategoryService)
	container.ItemHandler = examplehandler.NewItem(container.ItemService, container.ItemCategoryService)
	// scaffold:handlers:init
	return nil
}
