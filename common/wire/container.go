package wire

import "io/fs"

type Container struct {
	RequiredValidators
	RequiredConfigs
	RequiredDatabaseClients
	RequiredEventClients
	RequiredRepositories
	RequiredServices
	RequiredHTTPHandlers
	RequiredEventHandlers

	fs fs.FS
}
