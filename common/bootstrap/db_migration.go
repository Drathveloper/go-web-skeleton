package bootstrap

import (
	"fmt"

	"github.com/Drathveloper/go-web-skeleton/common/wire"
	exampleentity "github.com/Drathveloper/go-web-skeleton/example/repository/rdbms/entity"
	securityentity "github.com/Drathveloper/go-web-skeleton/security/repository/rdbms/entity"
	// scaffold:migrations:imports
)

func runDatabaseMigrations(container *wire.Container) error {
	entitiesToMigrate := []any{
		&securityentity.User{},
		&exampleentity.ItemCategory{},
		&exampleentity.Item{},
		// scaffold:migrations:entities
	}
	if len(entitiesToMigrate) == 0 {
		return nil
	}
	if err := container.PostgresClient.AutoMigrate(entitiesToMigrate...); err != nil {
		return fmt.Errorf("database migration failed: %w", err)
	}
	return nil
}
