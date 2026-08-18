package wire

import (
	examplerepository "github.com/Drathveloper/go-web-skeleton/example/repository/rdbms"
	securityRDBMSrepository "github.com/Drathveloper/go-web-skeleton/security/repository/rdbms"
	securityRedisRepository "github.com/Drathveloper/go-web-skeleton/security/repository/redis"
	// scaffold:repositories:imports
)

// Markers below are insertion points for `scaffold module`. The generator
// appends above each marker, so the surrounding code must stay valid with
// zero entries: that is why `container` is named rather than blanked out,
// even while nothing reads it yet.

type RequiredRepositories struct {
	UserRepository         *securityRDBMSrepository.User
	SessionsRepository     *securityRedisRepository.Session
	ItemCategoryRepository *examplerepository.ItemCategory
	ItemRepository         *examplerepository.Item
	// scaffold:repositories:fields
}

func injectRDBMSRepositories(container *Container) error {
	container.UserRepository = securityRDBMSrepository.NewUser(container.PostgresClient)
	container.ItemCategoryRepository = examplerepository.NewItemCategory(container.PostgresClient)
	container.ItemRepository = examplerepository.NewItem(container.PostgresClient)
	// scaffold:repositories:init
	return nil
}

func injectRedisRepositories(container *Container) error {
	container.SessionsRepository = securityRedisRepository.NewSession(container.RedisClient, container.Store)
	// scaffold:repositories:redis:init
	return nil
}
