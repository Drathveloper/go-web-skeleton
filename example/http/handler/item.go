package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/Drathveloper/go-web-skeleton/common/http/helper"
	"github.com/Drathveloper/go-web-skeleton/example/domain"
	"github.com/Drathveloper/go-web-skeleton/example/http/dto"
	"github.com/Drathveloper/go-web-skeleton/example/http/mapper"
)

const (
	listItemsErrMsg  = "list items handler failed"
	getItemErrMsg    = "get item handler failed"
	createItemErrMsg = "create item handler failed"
	updateItemErrMsg = "update item handler failed"
	deleteItemErrMsg = "delete item handler failed"

	itemListPage      = "item/list_items"
	itemFormFragment  = "fragments/form/modal"
	itemRowFragment   = "fragments/table/row"
	itemTbodyOOBSwap  = "afterbegin:#items-tbody"
	itemRowOOBReplace = "true"
)

// ItemService is declared by its consumer. The handler never imports the
// service package; only common/wire knows both sides.
type ItemService interface {
	ListItems(ctx context.Context) ([]domain.Item, error)
	ListItemLookups(ctx context.Context) ([]domain.Item, error)
	GetItemByID(ctx context.Context, id uint) (*domain.Item, error)
	GetItemByIDSummary(ctx context.Context, id uint) (*domain.Item, error)
	CreateItem(ctx context.Context, item *domain.Item) error
	UpdateItem(ctx context.Context, item *domain.Item) error
	DeleteItem(ctx context.Context, id uint) error
}

// ItemCategoryLookupService is the narrow slice of the item category service this
// handler needs to fill the relation <select>. A ref field composes services
// through an interface of its own rather than reaching for the whole thing.
type ItemCategoryLookupService interface {
	ListItemCategoryLookups(ctx context.Context) ([]domain.ItemCategory, error)
}

type Item struct {
	service             ItemService
	itemCategoryService ItemCategoryLookupService
}

func NewItem(
	service ItemService,
	itemCategoryService ItemCategoryLookupService) *Item {
	return &Item{
		service:             service,
		itemCategoryService: itemCategoryService,
	}
}

func (ctrl *Item) ListItemsView() gin.HandlerFunc {
	return func(c *gin.Context) { //nolint:varnamelen
		session := helper.MustGetSession(c)
		items, err := ctrl.service.ListItems(c.Request.Context())
		if err != nil {
			helper.RenderErrorPage(c, session, http.StatusInternalServerError,
				listItemsErrMsg, helper.UnexpectedErrorMessageID, err)
			return
		}
		response := mapper.DomainItemsToItemsViewResponse(session, items)
		c.HTML(http.StatusOK, itemListPage, response)
	}
}

func (ctrl *Item) CreateItemView() gin.HandlerFunc {
	return func(c *gin.Context) { //nolint:varnamelen
		session := helper.MustGetSession(c)
		itemCategories, err := ctrl.itemCategoryService.ListItemCategoryLookups(c.Request.Context())
		if err != nil {
			helper.RenderErrorPage(c, session, http.StatusInternalServerError,
				createItemErrMsg, helper.UnexpectedErrorMessageID, err)
			return
		}
		c.HTML(http.StatusOK, itemFormFragment,
			mapper.DomainItemToFormView(session, nil, itemCategories, false, ""))
	}
}

func (ctrl *Item) CreateItemProcess() gin.HandlerFunc {
	return func(c *gin.Context) { //nolint:varnamelen
		session := helper.MustGetSession(c)
		var request dto.Item
		if err := c.ShouldBind(&request); err != nil {
			ctrl.renderFormError(
				c, &request, false, createItemErrMsg, helper.InvalidFormDataMessageID, err)
			return
		}
		item, err := mapper.DTOItemToDomainItem(&request, 0)
		if err != nil {
			ctrl.renderFormError(
				c, &request, false, createItemErrMsg, helper.InvalidFormDataMessageID, err)
			return
		}
		if err = ctrl.service.CreateItem(c.Request.Context(), item); err != nil {
			ctrl.renderFormError(
				c, &request, false, createItemErrMsg, helper.UnexpectedErrorMessageID, err)
			return
		}
		helper.TriggerCloseModal(c)
		c.HTML(http.StatusOK, itemRowFragment,
			mapper.DomainItemToTableRow(session, item, itemTbodyOOBSwap))
	}
}

func (ctrl *Item) UpdateItemView() gin.HandlerFunc {
	return func(c *gin.Context) { //nolint:varnamelen
		session := helper.MustGetSession(c)
		itemID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			helper.RenderErrorPage(c, session, http.StatusBadRequest,
				getItemErrMsg, helper.InvalidFormDataMessageID, err)
			return
		}
		item, err := ctrl.service.GetItemByID(c.Request.Context(), uint(itemID))
		if err != nil {
			ctrl.renderLoadError(c, err)
			return
		}
		itemCategories, err := ctrl.itemCategoryService.ListItemCategoryLookups(c.Request.Context())
		if err != nil {
			helper.RenderErrorPage(c, session, http.StatusInternalServerError,
				getItemErrMsg, helper.UnexpectedErrorMessageID, err)
			return
		}
		c.HTML(http.StatusOK, itemFormFragment,
			mapper.DomainItemToFormView(session, mapper.DomainItemToDTOItem(item), itemCategories, true, ""))
	}
}

func (ctrl *Item) UpdateItemProcess() gin.HandlerFunc {
	return func(c *gin.Context) { //nolint:varnamelen
		session := helper.MustGetSession(c)
		itemID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			helper.RenderErrorPage(c, session, http.StatusBadRequest,
				updateItemErrMsg, helper.InvalidFormDataMessageID, err)
			return
		}
		var request dto.Item
		if bindErr := c.ShouldBind(&request); bindErr != nil {
			request.ID = uint(itemID)
			ctrl.renderFormError(
				c, &request, true, updateItemErrMsg, helper.InvalidFormDataMessageID, bindErr)
			return
		}
		request.ID = uint(itemID)
		item, err := mapper.DTOItemToDomainItem(&request, uint(itemID))
		if err != nil {
			ctrl.renderFormError(
				c, &request, true, updateItemErrMsg, helper.InvalidFormDataMessageID, err)
			return
		}
		if err = ctrl.service.UpdateItem(c.Request.Context(), item); err != nil {
			ctrl.renderFormError(
				c, &request, true, updateItemErrMsg, helper.UnexpectedErrorMessageID, err)
			return
		}
		helper.TriggerCloseModal(c)
		c.HTML(http.StatusOK, itemRowFragment,
			mapper.DomainItemToTableRow(session, item, itemRowOOBReplace))
	}
}

func (ctrl *Item) DeleteItemProcess() gin.HandlerFunc {
	return func(c *gin.Context) { //nolint:varnamelen
		itemID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			helper.LogError(c, deleteItemErrMsg, err)
			c.Status(http.StatusBadRequest)
			return
		}
		if err = ctrl.service.DeleteItem(c.Request.Context(), uint(itemID)); err != nil {
			helper.LogError(c, deleteItemErrMsg, err)
			c.Status(http.StatusInternalServerError)
			return
		}
		// The row is removed client side by hx-swap="outerHTML" on an empty
		// 200 body; there is nothing to render.
		c.Status(http.StatusOK)
	}
}

// renderFormError answers 422 with the form redisplayed: the submitted values
// are preserved so nothing is retyped, the message is localized and generic,
// and the real cause goes to the log.
func (ctrl *Item) renderFormError(
	ginCtx *gin.Context, request *dto.Item, isEdit bool, logMsg, messageID string, cause error) {
	session := helper.MustGetSession(ginCtx)
	helper.LogError(ginCtx, logMsg, cause)
	itemCategories, err := ctrl.itemCategoryService.ListItemCategoryLookups(ginCtx.Request.Context())
	if err != nil {
		helper.RenderErrorPage(ginCtx, session, http.StatusInternalServerError,
			logMsg, helper.UnexpectedErrorMessageID, err)
		return
	}
	ginCtx.HTML(http.StatusUnprocessableEntity, itemFormFragment,
		mapper.DomainItemToFormView(session, request, itemCategories, isEdit,
			helper.LocalizedMessage(session, messageID)))
}

// renderLoadError keeps the 404/500 split in one place: a record that does not
// exist is not a server failure, and the sentinel lives in the domain so this
// layer can tell them apart without importing the service package.
func (ctrl *Item) renderLoadError(ginCtx *gin.Context, err error) {
	session := helper.MustGetSession(ginCtx)
	if errors.Is(err, domain.ErrItemNotFound) {
		helper.RenderErrorPage(ginCtx, session, http.StatusNotFound,
			getItemErrMsg, helper.NotFoundErrorMessageID, err)
		return
	}
	helper.RenderErrorPage(ginCtx, session, http.StatusInternalServerError,
		getItemErrMsg, helper.UnexpectedErrorMessageID, err)
}
