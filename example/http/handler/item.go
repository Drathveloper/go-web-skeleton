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

type ItemService interface {
	ListItems(ctx context.Context) ([]domain.Item, error)
	ListItemLookups(ctx context.Context) ([]domain.Item, error)
	GetItemByID(ctx context.Context, id uint) (*domain.Item, error)
	GetItemByIDSummary(ctx context.Context, id uint) (*domain.Item, error)
	CreateItem(ctx context.Context, item *domain.Item) error
	UpdateItem(ctx context.Context, item *domain.Item) error
	DeleteItem(ctx context.Context, id uint) error
}

// ItemCategoryLookupService is the narrow slice of the category service this
// handler needs to fill the relation <select>. A ref field composes services
// through an interface of its own rather than reaching for the whole thing.
type ItemCategoryLookupService interface {
	ListItemCategoryLookups(ctx context.Context) ([]domain.ItemCategory, error)
}

type Item struct {
	service         ItemService
	categoryService ItemCategoryLookupService
}

func NewItem(service ItemService, categoryService ItemCategoryLookupService) *Item {
	return &Item{
		service:         service,
		categoryService: categoryService,
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
		c.HTML(http.StatusOK, itemListPage, mapper.DomainItemsToItemsViewResponse(session, items))
	}
}

func (ctrl *Item) CreateItemView() gin.HandlerFunc {
	return func(c *gin.Context) { //nolint:varnamelen
		session := helper.MustGetSession(c)
		categories, err := ctrl.categoryService.ListItemCategoryLookups(c.Request.Context())
		if err != nil {
			helper.RenderErrorPage(c, session, http.StatusInternalServerError,
				createItemErrMsg, helper.UnexpectedErrorMessageID, err)
			return
		}
		c.HTML(http.StatusOK, itemFormFragment,
			mapper.DomainItemToFormView(session, nil, categories, false, ""))
	}
}

func (ctrl *Item) CreateItemProcess() gin.HandlerFunc {
	return func(c *gin.Context) { //nolint:varnamelen
		session := helper.MustGetSession(c)
		var request dto.Item
		if err := c.ShouldBind(&request); err != nil {
			ctrl.renderFormError(c, &request, false, createItemErrMsg, helper.InvalidFormDataMessageID, err)
			return
		}
		item, err := mapper.DTOItemToDomainItem(&request, 0)
		if err != nil {
			ctrl.renderFormError(c, &request, false, createItemErrMsg, helper.InvalidFormDataMessageID, err)
			return
		}
		if err = ctrl.service.CreateItem(c.Request.Context(), item); err != nil {
			ctrl.renderFormError(c, &request, false, createItemErrMsg, helper.UnexpectedErrorMessageID, err)
			return
		}
		helper.TriggerCloseModal(c)
		c.HTML(http.StatusOK, itemRowFragment, mapper.DomainItemToTableRow(session, item, itemTbodyOOBSwap))
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
		categories, err := ctrl.categoryService.ListItemCategoryLookups(c.Request.Context())
		if err != nil {
			helper.RenderErrorPage(c, session, http.StatusInternalServerError,
				getItemErrMsg, helper.UnexpectedErrorMessageID, err)
			return
		}
		c.HTML(http.StatusOK, itemFormFragment,
			mapper.DomainItemToFormView(session, mapper.DomainItemToDTOItem(item), categories, true, ""))
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
			ctrl.renderFormError(c, &request, true, updateItemErrMsg, helper.InvalidFormDataMessageID, bindErr)
			return
		}
		request.ID = uint(itemID)
		item, err := mapper.DTOItemToDomainItem(&request, uint(itemID))
		if err != nil {
			ctrl.renderFormError(c, &request, true, updateItemErrMsg, helper.InvalidFormDataMessageID, err)
			return
		}
		if err = ctrl.service.UpdateItem(c.Request.Context(), item); err != nil {
			ctrl.renderFormError(c, &request, true, updateItemErrMsg, helper.UnexpectedErrorMessageID, err)
			return
		}
		helper.TriggerCloseModal(c)
		c.HTML(http.StatusOK, itemRowFragment, mapper.DomainItemToTableRow(session, item, itemRowOOBReplace))
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
		c.Status(http.StatusOK)
	}
}

// renderFormError answers 422 with the form redisplayed: the submitted values
// are preserved so nothing is retyped, the message is localized and generic,
// and the real cause goes to the log. The lookups are reloaded because the
// <select> needs its options again on the way back.
//
//nolint:varnamelen // c is the gin context, named c throughout the project
func (ctrl *Item) renderFormError(
	c *gin.Context, request *dto.Item, isEdit bool, logMsg, messageID string, cause error) {
	session := helper.MustGetSession(c)
	helper.LogError(c, logMsg, cause)
	categories, err := ctrl.categoryService.ListItemCategoryLookups(c.Request.Context())
	if err != nil {
		helper.RenderErrorPage(c, session, http.StatusInternalServerError,
			logMsg, helper.UnexpectedErrorMessageID, err)
		return
	}
	c.HTML(http.StatusUnprocessableEntity, itemFormFragment,
		mapper.DomainItemToFormView(session, request, categories, isEdit,
			helper.LocalizedMessage(session, messageID)))
}

//nolint:varnamelen // c is the gin context, named c throughout the project
func (ctrl *Item) renderLoadError(c *gin.Context, err error) {
	session := helper.MustGetSession(c)
	if errors.Is(err, domain.ErrItemNotFound) {
		helper.RenderErrorPage(c, session, http.StatusNotFound,
			getItemErrMsg, helper.NotFoundErrorMessageID, err)
		return
	}
	helper.RenderErrorPage(c, session, http.StatusInternalServerError,
		getItemErrMsg, helper.UnexpectedErrorMessageID, err)
}
