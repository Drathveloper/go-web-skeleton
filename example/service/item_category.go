package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
	"github.com/Drathveloper/go-web-skeleton/common/database"
	"github.com/Drathveloper/go-web-skeleton/example/domain"
)

const (
	listItemCategoriesErrMsg  = "list item categories service failed"
	getItemCategoryByIDErrMsg = "get item category by id service failed"
	createItemCategoryErrMsg  = "create item category service failed"
	updateItemCategoryErrMsg  = "update item category service failed"
	deleteItemCategoryErrMsg  = "delete item category service failed"
)

// ItemCategoryRepository is declared here, by the consumer, not next to its
// implementation: the service never imports the repository package and only
// common/wire sees both sides.
type ItemCategoryRepository interface {
	FindItemCategoryByID(ctx context.Context, id uint) (*domain.ItemCategory, error)
	FindItemCategoryByIDSummary(ctx context.Context, id uint) (*domain.ItemCategory, error)
	FindAllItemCategories(ctx context.Context) ([]domain.ItemCategory, error)
	FindAllItemCategoryLookups(ctx context.Context) ([]domain.ItemCategory, error)
	CreateItemCategory(ctx context.Context, itemCategory *domain.ItemCategory) (*domain.ItemCategory, error)
	UpdateItemCategory(ctx context.Context, itemCategory *domain.ItemCategory) (*domain.ItemCategory, error)
	DeleteItemCategory(ctx context.Context, id uint) error
}

type ItemCategory struct {
	repository ItemCategoryRepository
}

func NewItemCategory(repository ItemCategoryRepository) *ItemCategory {
	return &ItemCategory{
		repository: repository,
	}
}

func (i *ItemCategory) ListItemCategories(ctx context.Context) ([]domain.ItemCategory, error) {
	itemCategories, err := i.repository.FindAllItemCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, listItemCategoriesErrMsg, err)
	}
	return itemCategories, nil
}

func (i *ItemCategory) ListItemCategoryLookups(ctx context.Context) ([]domain.ItemCategory, error) {
	itemCategories, err := i.repository.FindAllItemCategoryLookups(ctx)
	if err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, listItemCategoriesErrMsg, err)
	}
	return itemCategories, nil
}

func (i *ItemCategory) GetItemCategoryByID(ctx context.Context, id uint) (*domain.ItemCategory, error) {
	itemCategory, err := i.repository.FindItemCategoryByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			return nil, fmt.Errorf(
				constants.DefaultWrappedErrorTemplate, getItemCategoryByIDErrMsg, domain.ErrItemCategoryNotFound)
		default:
			return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, getItemCategoryByIDErrMsg, err)
		}
	}
	return itemCategory, nil
}

func (i *ItemCategory) GetItemCategoryByIDSummary(ctx context.Context, id uint) (*domain.ItemCategory, error) {
	itemCategory, err := i.repository.FindItemCategoryByIDSummary(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			return nil, fmt.Errorf(
				constants.DefaultWrappedErrorTemplate, getItemCategoryByIDErrMsg, domain.ErrItemCategoryNotFound)
		default:
			return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, getItemCategoryByIDErrMsg, err)
		}
	}
	return itemCategory, nil
}

func (i *ItemCategory) CreateItemCategory(ctx context.Context, itemCategory *domain.ItemCategory) error {
	created, err := i.repository.CreateItemCategory(ctx, itemCategory)
	if err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, createItemCategoryErrMsg, err)
	}
	*itemCategory = *created
	return nil
}

func (i *ItemCategory) UpdateItemCategory(ctx context.Context, itemCategory *domain.ItemCategory) error {
	updated, err := i.repository.UpdateItemCategory(ctx, itemCategory)
	if err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, updateItemCategoryErrMsg, err)
	}
	*itemCategory = *updated
	return nil
}

func (i *ItemCategory) DeleteItemCategory(ctx context.Context, id uint) error {
	err := i.repository.DeleteItemCategory(ctx, id)
	if err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, deleteItemCategoryErrMsg, err)
	}
	return nil
}
