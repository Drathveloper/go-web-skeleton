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
	listItemCategoriesErrMsg = "list item categories handler failed"
	getItemCategoryErrMsg    = "get item category handler failed"
	createItemCategoryErrMsg = "create item category handler failed"
	updateItemCategoryErrMsg = "update item category handler failed"
	deleteItemCategoryErrMsg = "delete item category handler failed"

	itemCategoryListPage      = "item_category/list_item_categories"
	itemCategoryFormFragment  = "fragments/form/modal"
	itemCategoryRowFragment   = "fragments/table/row"
	itemCategoryTbodyOOBSwap  = "afterbegin:#item-categories-tbody"
	itemCategoryRowOOBReplace = "true"
)

// ItemCategoryService is declared by its consumer. The handler never imports the
// service package; only common/wire knows both sides.
type ItemCategoryService interface {
	ListItemCategories(ctx context.Context) ([]domain.ItemCategory, error)
	ListItemCategoryLookups(ctx context.Context) ([]domain.ItemCategory, error)
	GetItemCategoryByID(ctx context.Context, id uint) (*domain.ItemCategory, error)
	GetItemCategoryByIDSummary(ctx context.Context, id uint) (*domain.ItemCategory, error)
	CreateItemCategory(ctx context.Context, itemCategory *domain.ItemCategory) error
	UpdateItemCategory(ctx context.Context, itemCategory *domain.ItemCategory) error
	DeleteItemCategory(ctx context.Context, id uint) error
}

type ItemCategory struct {
	service ItemCategoryService
}

func NewItemCategory(service ItemCategoryService) *ItemCategory {
	return &ItemCategory{
		service: service,
	}
}

func (ctrl *ItemCategory) ListItemCategoriesView() gin.HandlerFunc {
	return func(c *gin.Context) { //nolint:varnamelen
		session := helper.MustGetSession(c)
		itemCategories, err := ctrl.service.ListItemCategories(c.Request.Context())
		if err != nil {
			helper.RenderErrorPage(c, session, http.StatusInternalServerError,
				listItemCategoriesErrMsg, helper.UnexpectedErrorMessageID, err)
			return
		}
		response := mapper.DomainItemCategoriesToItemCategoriesViewResponse(session, itemCategories)
		c.HTML(http.StatusOK, itemCategoryListPage, response)
	}
}

func (ctrl *ItemCategory) CreateItemCategoryView() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := helper.MustGetSession(c)
		c.HTML(http.StatusOK, itemCategoryFormFragment,
			mapper.DomainItemCategoryToFormView(session, nil, false, ""))
	}
}

func (ctrl *ItemCategory) CreateItemCategoryProcess() gin.HandlerFunc {
	return func(c *gin.Context) { //nolint:varnamelen
		session := helper.MustGetSession(c)
		var request dto.ItemCategory
		if err := c.ShouldBind(&request); err != nil {
			ctrl.renderFormError(
				c, &request, false, createItemCategoryErrMsg, helper.InvalidFormDataMessageID, err)
			return
		}
		itemCategory := mapper.DTOItemCategoryToDomainItemCategory(&request, 0)
		if err := ctrl.service.CreateItemCategory(c.Request.Context(), itemCategory); err != nil {
			ctrl.renderFormError(
				c, &request, false, createItemCategoryErrMsg, helper.UnexpectedErrorMessageID, err)
			return
		}
		helper.TriggerCloseModal(c)
		c.HTML(http.StatusOK, itemCategoryRowFragment,
			mapper.DomainItemCategoryToTableRow(session, itemCategory, itemCategoryTbodyOOBSwap))
	}
}

func (ctrl *ItemCategory) UpdateItemCategoryView() gin.HandlerFunc {
	return func(c *gin.Context) { //nolint:varnamelen
		session := helper.MustGetSession(c)
		itemCategoryID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			helper.RenderErrorPage(c, session, http.StatusBadRequest,
				getItemCategoryErrMsg, helper.InvalidFormDataMessageID, err)
			return
		}
		itemCategory, err := ctrl.service.GetItemCategoryByID(c.Request.Context(), uint(itemCategoryID))
		if err != nil {
			ctrl.renderLoadError(c, err)
			return
		}
		c.HTML(http.StatusOK, itemCategoryFormFragment,
			mapper.DomainItemCategoryToFormView(session, mapper.DomainItemCategoryToDTOItemCategory(itemCategory), true, ""))
	}
}

func (ctrl *ItemCategory) UpdateItemCategoryProcess() gin.HandlerFunc {
	return func(c *gin.Context) { //nolint:varnamelen
		session := helper.MustGetSession(c)
		itemCategoryID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			helper.RenderErrorPage(c, session, http.StatusBadRequest,
				updateItemCategoryErrMsg, helper.InvalidFormDataMessageID, err)
			return
		}
		var request dto.ItemCategory
		if bindErr := c.ShouldBind(&request); bindErr != nil {
			request.ID = uint(itemCategoryID)
			ctrl.renderFormError(
				c, &request, true, updateItemCategoryErrMsg, helper.InvalidFormDataMessageID, bindErr)
			return
		}
		request.ID = uint(itemCategoryID)
		itemCategory := mapper.DTOItemCategoryToDomainItemCategory(&request, uint(itemCategoryID))
		if err = ctrl.service.UpdateItemCategory(c.Request.Context(), itemCategory); err != nil {
			ctrl.renderFormError(
				c, &request, true, updateItemCategoryErrMsg, helper.UnexpectedErrorMessageID, err)
			return
		}
		helper.TriggerCloseModal(c)
		c.HTML(http.StatusOK, itemCategoryRowFragment,
			mapper.DomainItemCategoryToTableRow(session, itemCategory, itemCategoryRowOOBReplace))
	}
}

func (ctrl *ItemCategory) DeleteItemCategoryProcess() gin.HandlerFunc {
	return func(c *gin.Context) { //nolint:varnamelen
		itemCategoryID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			helper.LogError(c, deleteItemCategoryErrMsg, err)
			c.Status(http.StatusBadRequest)
			return
		}
		if err = ctrl.service.DeleteItemCategory(c.Request.Context(), uint(itemCategoryID)); err != nil {
			helper.LogError(c, deleteItemCategoryErrMsg, err)
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
func (ctrl *ItemCategory) renderFormError(
	ginCtx *gin.Context, request *dto.ItemCategory, isEdit bool, logMsg, messageID string, cause error) {
	session := helper.MustGetSession(ginCtx)
	helper.LogError(ginCtx, logMsg, cause)
	ginCtx.HTML(http.StatusUnprocessableEntity, itemCategoryFormFragment,
		mapper.DomainItemCategoryToFormView(session, request, isEdit,
			helper.LocalizedMessage(session, messageID)))
}

// renderLoadError keeps the 404/500 split in one place: a record that does not
// exist is not a server failure, and the sentinel lives in the domain so this
// layer can tell them apart without importing the service package.
func (ctrl *ItemCategory) renderLoadError(ginCtx *gin.Context, err error) {
	session := helper.MustGetSession(ginCtx)
	if errors.Is(err, domain.ErrItemCategoryNotFound) {
		helper.RenderErrorPage(ginCtx, session, http.StatusNotFound,
			getItemCategoryErrMsg, helper.NotFoundErrorMessageID, err)
		return
	}
	helper.RenderErrorPage(ginCtx, session, http.StatusInternalServerError,
		getItemCategoryErrMsg, helper.UnexpectedErrorMessageID, err)
}
