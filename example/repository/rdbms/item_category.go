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
	findAllItemCategoriesErrMsg = "find all item categories repository failed"
	findItemCategoryByIDErrMsg  = "find item category by id repository failed"
	createItemCategoryErrMsg    = "create item category repository failed"
	updateItemCategoryErrMsg    = "update item category repository failed"
	deleteItemCategoryErrMsg    = "delete item category repository failed"
)

type ItemCategory struct {
	db rdbms.PostgresClient
}

func NewItemCategory(db rdbms.PostgresClient) *ItemCategory {
	return &ItemCategory{
		db: db,
	}
}

func (i *ItemCategory) FindAllItemCategories(
	ctx context.Context) ([]domain.ItemCategory, error) {
	var itemCategoriesEntity []entity.ItemCategory

	err := i.db.WithContext(ctx).
		Find(&itemCategoriesEntity).Error
	if err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, findAllItemCategoriesErrMsg, err)
	}

	return mapper.EntityItemCategoriesToDomainItemCategories(itemCategoriesEntity), nil
}

func (i *ItemCategory) FindAllItemCategoryLookups(
	ctx context.Context) ([]domain.ItemCategory, error) {
	var itemCategoriesEntity []entity.ItemCategory

	err := i.db.WithContext(ctx).
		Select("id", "name").
		Find(&itemCategoriesEntity).Error
	if err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, findAllItemCategoriesErrMsg, err)
	}

	return mapper.EntityItemCategoriesToDomainItemCategories(itemCategoriesEntity), nil
}

func (i *ItemCategory) FindItemCategoryByID(
	ctx context.Context, itemCategoryID uint) (*domain.ItemCategory, error) {
	var itemCategoryEntity entity.ItemCategory

	err := i.db.WithContext(ctx).
		First(&itemCategoryEntity, itemCategoryID).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, fmt.Errorf(
				constants.DefaultWrappedErrorTemplate, findItemCategoryByIDErrMsg, database.ErrRecordNotFound)
		default:
			return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, findItemCategoryByIDErrMsg, err)
		}
	}

	return mapper.EntityItemCategoryToDomainItemCategory(&itemCategoryEntity), nil
}

func (i *ItemCategory) FindItemCategoryByIDSummary(
	ctx context.Context, itemCategoryID uint) (*domain.ItemCategory, error) {
	var itemCategoryEntity entity.ItemCategory

	err := i.db.WithContext(ctx).
		Select("id", "name").
		First(&itemCategoryEntity, itemCategoryID).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, fmt.Errorf(
				constants.DefaultWrappedErrorTemplate, findItemCategoryByIDErrMsg, database.ErrRecordNotFound)
		default:
			return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, findItemCategoryByIDErrMsg, err)
		}
	}

	return mapper.EntityItemCategoryToDomainItemCategory(&itemCategoryEntity), nil
}

func (i *ItemCategory) CreateItemCategory(
	ctx context.Context, itemCategory *domain.ItemCategory) (*domain.ItemCategory, error) {
	itemCategoryEntity := mapper.DomainItemCategoryToEntityItemCategory(itemCategory)

	err := i.db.WithContext(ctx).
		Create(itemCategoryEntity).Error
	if err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, createItemCategoryErrMsg, err)
	}

	itemCategory, err = i.FindItemCategoryByID(ctx, itemCategoryEntity.ID)
	if err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, createItemCategoryErrMsg, err)
	}

	return itemCategory, nil
}

func (i *ItemCategory) UpdateItemCategory(
	ctx context.Context, itemCategory *domain.ItemCategory) (*domain.ItemCategory, error) {
	itemCategoryEntity := mapper.DomainItemCategoryToEntityItemCategory(itemCategory)

	err := i.db.WithContext(ctx).
		Where("id = ?", itemCategoryEntity.ID).
		Select("name", "updated_at").
		Updates(itemCategoryEntity).Error
	if err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, updateItemCategoryErrMsg, err)
	}

	itemCategory, err = i.FindItemCategoryByID(ctx, itemCategoryEntity.ID)
	if err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, updateItemCategoryErrMsg, err)
	}

	return itemCategory, nil
}

func (i *ItemCategory) DeleteItemCategory(
	ctx context.Context, itemCategoryID uint) error {
	err := i.db.WithContext(ctx).
		Delete(&entity.ItemCategory{}, itemCategoryID).Error
	if err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, deleteItemCategoryErrMsg, err)
	}
	return nil
}
