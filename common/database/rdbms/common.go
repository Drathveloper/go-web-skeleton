package rdbms

import (
	"context"

	"gorm.io/gorm"
)

type PostgresClient interface {
	WithContext(ctx context.Context) *gorm.DB
}
