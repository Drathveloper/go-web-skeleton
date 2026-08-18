package wire

import (
	"fmt"
	"io/fs"
)

func Wire(fileSystem fs.FS) (*Container, error) {
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
		fs: fileSystem,
	}

	for _, injectable := range injectables {
		if err := injectable(container); err != nil {
			return nil, fmt.Errorf("dependency injection failed: %w", err)
		}
	}
	return container, nil
}
