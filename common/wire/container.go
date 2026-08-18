package wire

import (
	"io/fs"

	"github.com/Drathveloper/go-web-skeleton/common/config/model"
)

type Container struct {
	RequiredValidators
	RequiredConfigs
	RequiredDatabaseClients
	RequiredEventClients
	RequiredRepositories
	RequiredServices
	RequiredHTTPHandlers
	RequiredEventHandlers

	fs        fs.FS
	buildInfo model.BuildInfo
}
