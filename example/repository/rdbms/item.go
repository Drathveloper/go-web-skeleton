package rdbms

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
	"github.com/Drathveloper/go-web-skeleton/common/database"
	"github.com/Drathveloper/go-web-skeleton/common/database/rdbms"
	"github.com/Drathveloper/go-web-skeleton/example/domain"
	"github.com/Drathveloper/go-web-skeleton/example/repository/rdbms/entity"
	"github.com/Drathveloper/go-web-skeleton/example/repository/rdbms/mapper"
)

const (
	findAllItemsErrMsg = "find all items repository failed"
	findItemByIDErrMsg = "find item by id repository failed"
	createItemErrMsg   = "create item repository failed"
	updateItemErrMsg   = "update item repository failed"
	deleteItemErrMsg   = "delete item repository failed"
)

// The relations below are preloaded so a listing can show the related name
// without the handler issuing one query per row.
const (
	itemCategoryRelation = "Category"
)

type Item struct {
	db rdbms.PostgresClient
}

func NewItem(db rdbms.PostgresClient) *Item {
	return &Item{
		db: db,
	}
}

func (i *Item) FindAllItems(
	ctx context.Context) ([]domain.Item, error) {
	var itemsEntity []entity.Item

	err := i.db.WithContext(ctx).
		Preload(itemCategoryRelation).
		Find(&itemsEntity).Error
	if err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, findAllItemsErrMsg, err)
	}

	return mapper.EntityItemsToDomainItems(itemsEntity), nil
}

func (i *Item) FindAllItemLookups(
	ctx context.Context) ([]domain.Item, error) {
	var itemsEntity []entity.Item

	err := i.db.WithContext(ctx).
		Select("id", "name").
		Find(&itemsEntity).Error
	if err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, findAllItemsErrMsg, err)
	}

	return mapper.EntityItemsToDomainItems(itemsEntity), nil
}

func (i *Item) FindItemByID(
	ctx context.Context, itemID uint) (*domain.Item, error) {
	var itemEntity entity.Item

	err := i.db.WithContext(ctx).
		Preload(itemCategoryRelation).
		First(&itemEntity, itemID).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, fmt.Errorf(
				constants.DefaultWrappedErrorTemplate, findItemByIDErrMsg, database.ErrRecordNotFound)
		default:
			return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, findItemByIDErrMsg, err)
		}
	}

	return mapper.EntityItemToDomainItem(&itemEntity), nil
}

func (i *Item) FindItemByIDSummary(
	ctx context.Context, itemID uint) (*domain.Item, error) {
	var itemEntity entity.Item

	err := i.db.WithContext(ctx).
		Select("id", "name").
		First(&itemEntity, itemID).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, fmt.Errorf(
				constants.DefaultWrappedErrorTemplate, findItemByIDErrMsg, database.ErrRecordNotFound)
		default:
			return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, findItemByIDErrMsg, err)
		}
	}

	return mapper.EntityItemToDomainItem(&itemEntity), nil
}

func (i *Item) CreateItem(
	ctx context.Context, item *domain.Item) (*domain.Item, error) {
	itemEntity := mapper.DomainItemToEntityItem(item)

	err := i.db.WithContext(ctx).
		Create(itemEntity).Error
	if err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, createItemErrMsg, err)
	}

	item, err = i.FindItemByID(ctx, itemEntity.ID)
	if err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, createItemErrMsg, err)
	}

	return item, nil
}

func (i *Item) UpdateItem(
	ctx context.Context, item *domain.Item) (*domain.Item, error) {
	itemEntity := mapper.DomainItemToEntityItem(item)

	err := i.db.WithContext(ctx).
		Where("id = ?", itemEntity.ID).
		Select("name", "notes", "stock", "price", "contact", "released_at", "starts_at", "category_id", "active",
			"updated_at").
		Updates(itemEntity).Error
	if err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, updateItemErrMsg, err)
	}

	item, err = i.FindItemByID(ctx, itemEntity.ID)
	if err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, updateItemErrMsg, err)
	}

	return item, nil
}

func (i *Item) DeleteItem(
	ctx context.Context, itemID uint) error {
	err := i.db.WithContext(ctx).
		Delete(&entity.Item{}, itemID).Error
	if err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, deleteItemErrMsg, err)
	}
	return nil
}
