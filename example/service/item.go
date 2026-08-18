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
	listItemsErrMsg   = "list items service failed"
	getItemByIDErrMsg = "get item by id service failed"
	createItemErrMsg  = "create item service failed"
	updateItemErrMsg  = "update item service failed"
	deleteItemErrMsg  = "delete item service failed"
)

// ItemRepository is declared here, by the consumer, not next to its
// implementation: the service never imports the repository package and only
// common/wire sees both sides.
type ItemRepository interface {
	FindItemByID(ctx context.Context, id uint) (*domain.Item, error)
	FindItemByIDSummary(ctx context.Context, id uint) (*domain.Item, error)
	FindAllItems(ctx context.Context) ([]domain.Item, error)
	FindAllItemLookups(ctx context.Context) ([]domain.Item, error)
	CreateItem(ctx context.Context, item *domain.Item) (*domain.Item, error)
	UpdateItem(ctx context.Context, item *domain.Item) (*domain.Item, error)
	DeleteItem(ctx context.Context, id uint) error
}

type Item struct {
	repository ItemRepository
}

func NewItem(repository ItemRepository) *Item {
	return &Item{
		repository: repository,
	}
}

func (i *Item) ListItems(
	ctx context.Context) ([]domain.Item, error) {
	items, err := i.repository.FindAllItems(ctx)
	if err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, listItemsErrMsg, err)
	}
	return items, nil
}

func (i *Item) ListItemLookups(
	ctx context.Context) ([]domain.Item, error) {
	items, err := i.repository.FindAllItemLookups(ctx)
	if err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, listItemsErrMsg, err)
	}
	return items, nil
}

func (i *Item) GetItemByID(
	ctx context.Context, id uint) (*domain.Item, error) {
	item, err := i.repository.FindItemByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			return nil, fmt.Errorf(
				constants.DefaultWrappedErrorTemplate, getItemByIDErrMsg, domain.ErrItemNotFound)
		default:
			return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, getItemByIDErrMsg, err)
		}
	}
	return item, nil
}

func (i *Item) GetItemByIDSummary(
	ctx context.Context, id uint) (*domain.Item, error) {
	item, err := i.repository.FindItemByIDSummary(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			return nil, fmt.Errorf(
				constants.DefaultWrappedErrorTemplate, getItemByIDErrMsg, domain.ErrItemNotFound)
		default:
			return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, getItemByIDErrMsg, err)
		}
	}
	return item, nil
}

func (i *Item) CreateItem(
	ctx context.Context, item *domain.Item) error {
	created, err := i.repository.CreateItem(ctx, item)
	if err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, createItemErrMsg, err)
	}
	*item = *created
	return nil
}

func (i *Item) UpdateItem(
	ctx context.Context, item *domain.Item) error {
	updated, err := i.repository.UpdateItem(ctx, item)
	if err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, updateItemErrMsg, err)
	}
	*item = *updated
	return nil
}

func (i *Item) DeleteItem(
	ctx context.Context, id uint) error {
	err := i.repository.DeleteItem(ctx, id)
	if err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, deleteItemErrMsg, err)
	}
	return nil
}
