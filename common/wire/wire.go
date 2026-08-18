package wire

import (
	"fmt"
	"io/fs"

	"github.com/Drathveloper/go-web-skeleton/common/config/model"
)

func Wire(fileSystem fs.FS, buildInfo model.BuildInfo) (*Container, error) {
	var injectables = []injectableFunc{
		injectValidators,
		injectConfig,
		injectPostgresClient,
		injectRedisClient,
		injectEventClients,
		injectRDBMSRepositories,
		injectRedisRepositories,
		injectServices,
		injectHTTPHandlers,
		injectEventHandlers,
	}

	container := &Container{
		fs:        fileSystem,
		buildInfo: buildInfo,
	}

	for _, injectable := range injectables {
		if err := injectable(container); err != nil {
			return nil, fmt.Errorf("dependency injection failed: %w", err)
		}
	}
	return container, nil
}
