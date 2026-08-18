package wire

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
)

const injectDatabaseClientsBaseErrMsg = "inject database clients failed"

type RequiredDatabaseClients struct {
	PostgresClient *gorm.DB
	RedisClient    redis.UniversalClient
}

func injectPostgresClient(container *Container) error {
	dsn := container.Store.GetPostgresConnectionString()
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, injectDatabaseClientsBaseErrMsg, err)
	}

	poolConfig.MaxConns = container.Store.GetPostgresPoolConfig().MaxConnections
	poolConfig.MinConns = container.Store.GetPostgresPoolConfig().MinConnections
	poolConfig.MaxConnIdleTime = container.Store.GetPostgresPoolConfig().MaxIdleTime
	poolConfig.HealthCheckPeriod = container.Store.GetPostgresPoolConfig().HealthCheckPeriod
	poolConfig.MaxConnLifetimeJitter = container.Store.GetPostgresPoolConfig().MaxConnLifetimeJitter

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, injectDatabaseClientsBaseErrMsg, err)
	}

	sqlDB := stdlib.OpenDBFromPool(pool)

	database, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, injectDatabaseClientsBaseErrMsg, err)
	}

	container.PostgresClient = database
	return nil
}

func injectRedisClient(container *Container) error {
	redisOpts, err := container.Store.GetRedisOptions()
	if err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, injectDatabaseClientsBaseErrMsg, err)
	}
	container.RedisClient = redis.NewUniversalClient(redisOpts)
	if cmd := container.RedisClient.Ping(context.Background()); cmd.Err() != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, injectDatabaseClientsBaseErrMsg, cmd.Err())
	}
	return nil
}
