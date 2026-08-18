package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	commondomain "github.com/Drathveloper/go-web-skeleton/common/domain"
	commondto "github.com/Drathveloper/go-web-skeleton/common/http/dto"
	"github.com/Drathveloper/go-web-skeleton/common/http/helper"
	commonmapper "github.com/Drathveloper/go-web-skeleton/common/http/mapper"
	"github.com/Drathveloper/go-web-skeleton/security/domain"
	eventdto "github.com/Drathveloper/go-web-skeleton/security/event/dto"
	"github.com/Drathveloper/go-web-skeleton/security/http/dto"
	"github.com/Drathveloper/go-web-skeleton/security/http/mapper"
)

const (
	userFormFragment = "fragments/auth/form"
	userRowFragment  = "fragments/auth/row"
)

const (
	userListPath = "/auth/user"
	userNewPath  = "/auth/user/new"
)

const (
	listUsersErrMsg  = "list users error"
	getUserErrMsg    = "get user error"
	createUserErrMsg = "create user error"
	updateUserErrMsg = "update user error"

	userCreatedMessageID = "auth.messages.user_created"
	userUpdatedMessageID = "auth.messages.user_updated"

	// createUserOOB inserts the brand new row at the top of the table; updateUserOOB
	// replaces the row that already carries the id of the edited user.
	createUserOOB = "afterbegin:#users-tbody"
	updateUserOOB = "true"

	userIDBase    = 10
	userIDBitSize = 64
)

type UserManagementService interface {
	ListUsers(ctx context.Context, pagination *commondomain.Pagination) ([]domain.User, error)
	GetUserByID(ctx context.Context, ID uint) (*domain.User, error)
	CreateUser(ctx context.Context, user *domain.User) error
	UpdateUser(ctx context.Context, user *domain.User) error
}

type UserManagement struct {
	service        UserManagementService
	eventPublisher EventPublisher
}

func NewUserManagement(service UserManagementService, eventPublisher EventPublisher) *UserManagement {
	return &UserManagement{
		service:        service,
		eventPublisher: eventPublisher,
	}
}

func newUserCreatedEvent(actor string, request *dto.CreateUserProcessRequest, isSuccess bool) *eventdto.UserCreatedEvent {
	return &eventdto.UserCreatedEvent{
		ActorUsername: actor,
		Username:      request.Username,
		Roles:         request.Roles,
		IsSuccess:     isSuccess,
	}
}

func newUserUpdatedEvent(
	actor string, userID uint, request *dto.UpdateUserProcessRequest, isSuccess bool) *eventdto.UserUpdatedEvent {
	return &eventdto.UserUpdatedEvent{
		ActorUsername: actor,
		Username:      request.Username,
		Roles:         request.Roles,
		UserID:        userID,
		IsSuccess:     isSuccess,
	}
}

func userEditPath(userID string) string {
	return userListPath + "/" + userID + "/edit"
}

func userFormData(session *commondomain.Session, user *dto.User, roles []string, isEdit bool, errMsg string) gin.H {
	var userRoles []string
	if user != nil {
		userRoles = user.Roles
	}
	return gin.H{
		"Language":     session.Language,
		"CSRFToken":    session.CSRFToken,
		"User":         user,
		"UserRoles":    userRoles,
		"AllowedRoles": roles,
		"IsEdit":       isEdit,
		"Error":        errMsg,
	}
}

func userRowData(session *commondomain.Session, user *dto.User, oob string) gin.H {
	return gin.H{
		"Language":  session.Language,
		"CSRFToken": session.CSRFToken,
		"User":      user,
		"OOB":       oob,
	}
}

func (ctrl *UserManagement) ListUsersView() gin.HandlerFunc {
	return func(c *gin.Context) { //nolint:varnamelen
		session := helper.MustGetSession(c)
		var queryPagination commondto.Pagination
		// A malformed page/size in the query string is not worth an error page: the
		// zero value means "first page, default size" downstream.
		_ = c.ShouldBindQuery(&queryPagination)
		pagination := commonmapper.MapDTOPaginationToDomainPagination(&queryPagination)
		users, err := ctrl.service.ListUsers(c.Request.Context(), pagination)
		if err != nil {
			helper.RenderErrorPage(
				c, session, http.StatusInternalServerError,
				listUsersErrMsg, helper.UnexpectedErrorMessageID, err)
			return
		}
		response := mapper.DomainUsersToUsersViewResponse(users, session)
		c.HTML(http.StatusOK, "auth/list_users", response)
	}
}

func (ctrl *UserManagement) CreateUserView() gin.HandlerFunc {
	return func(c *gin.Context) { //nolint:varnamelen
		session := helper.MustGetSession(c)
		if helper.IsHTMXRequest(c) {
			c.HTML(http.StatusOK, userFormFragment,
				userFormData(session, nil, commondomain.GetAllowedRoles(), false, ""))
			return
		}
		response := mapper.SessionDomainToCreateUserViewResponse(session)
		c.HTML(http.StatusOK, "auth/create_user", response)
	}
}

func (ctrl *UserManagement) CreateUserProcess() gin.HandlerFunc {
	return func(c *gin.Context) { //nolint:varnamelen
		session := helper.MustGetSession(c)
		var request dto.CreateUserProcessRequest
		htmx := helper.IsHTMXRequest(c)
		if err := c.ShouldBind(&request); err != nil {
			helper.LogError(c, invalidFormDataErrMsg, err)
			if htmx {
				partial := &dto.User{Username: request.Username, Roles: request.Roles}
				c.HTML(http.StatusUnprocessableEntity, userFormFragment,
					userFormData(session, partial, commondomain.GetAllowedRoles(), false,
						helper.LocalizedMessage(session, helper.InvalidFormDataMessageID)))
				return
			}
			helper.FlashError(c, session, http.StatusBadRequest, invalidFormDataErrMsg, helper.InvalidFormDataMessageID, err)
			c.Redirect(http.StatusFound, userNewPath)
			return
		}
		user := mapper.DTOCreateUserProcessRequestToDomainUser(&request)
		if err := ctrl.service.CreateUser(c.Request.Context(), user); err != nil {
			ctrl.eventPublisher.Publish(newUserCreatedEvent(session.Username, &request, false))
			if htmx {
				helper.LogError(c, createUserErrMsg, err)
				partial := &dto.User{Username: request.Username, Roles: request.Roles}
				c.HTML(http.StatusUnprocessableEntity, userFormFragment,
					userFormData(session, partial, commondomain.GetAllowedRoles(), false,
						helper.LocalizedMessage(session, helper.UnexpectedErrorMessageID)))
				return
			}
			helper.FlashError(c, session, http.StatusInternalServerError, createUserErrMsg, helper.UnexpectedErrorMessageID, err)
			c.Redirect(http.StatusFound, userNewPath)
			return
		}
		ctrl.eventPublisher.Publish(newUserCreatedEvent(session.Username, &request, true))
		if htmx {
			helper.TriggerCloseModal(c)
			c.HTML(http.StatusOK, userRowFragment,
				userRowData(session, mapper.DomainUserToDTOUser(user), createUserOOB))
			return
		}
		session.AddAlertMessages(
			commondomain.NewSuccessAlertMessage(helper.LocalizedMessage(session, userCreatedMessageID), ""))
		c.Redirect(http.StatusFound, userListPath)
	}
}

func (ctrl *UserManagement) UpdateUserView() gin.HandlerFunc {
	return func(c *gin.Context) { //nolint:varnamelen
		session := helper.MustGetSession(c)
		userID, err := strconv.ParseUint(c.Param("id"), userIDBase, userIDBitSize)
		if err != nil {
			helper.RenderErrorPage(
				c, session, http.StatusBadRequest,
				invalidFormDataErrMsg, helper.InvalidFormDataMessageID, err)
			return
		}
		user, err := ctrl.service.GetUserByID(c.Request.Context(), uint(userID))
		if err != nil {
			// A user that does not exist is a 404, not a server failure. The
			// sentinel lives in security/domain precisely so this layer can
			// tell the two apart without importing the service package.
			if errors.Is(err, domain.ErrUserNotFound) {
				helper.RenderErrorPage(c, session, http.StatusNotFound, getUserErrMsg, helper.NotFoundErrorMessageID, err)
				return
			}
			helper.RenderErrorPage(
				c, session, http.StatusInternalServerError,
				getUserErrMsg, helper.UnexpectedErrorMessageID, err)
			return
		}
		if helper.IsHTMXRequest(c) {
			c.HTML(http.StatusOK, userFormFragment,
				userFormData(session, mapper.DomainUserToDTOUser(user), commondomain.GetAllowedRoles(), true, ""))
			return
		}
		response := mapper.DomainUserToUpdateUserViewResponse(user, session)
		c.HTML(http.StatusOK, "auth/update_user", response)
	}
}

func (ctrl *UserManagement) UpdateUserProcess() gin.HandlerFunc {
	return func(c *gin.Context) { //nolint:varnamelen
		session := helper.MustGetSession(c)
		htmx := helper.IsHTMXRequest(c)
		userID, err := strconv.ParseUint(c.Param("id"), userIDBase, userIDBitSize)
		if err != nil {
			helper.LogError(c, invalidFormDataErrMsg, err)
			if htmx {
				c.HTML(http.StatusBadRequest, userFormFragment,
					userFormData(session, nil, commondomain.GetAllowedRoles(), true,
						helper.LocalizedMessage(session, helper.InvalidFormDataMessageID)))
				return
			}
			helper.FlashError(c, session, http.StatusBadRequest, invalidFormDataErrMsg, helper.InvalidFormDataMessageID, err)
			c.Redirect(http.StatusFound, userListPath)
			return
		}
		var request dto.UpdateUserProcessRequest
		if err = c.ShouldBind(&request); err != nil {
			helper.LogError(c, invalidFormDataErrMsg, err)
			if htmx {
				partial := &dto.User{ID: uint(userID), Username: request.Username, Roles: request.Roles}
				c.HTML(http.StatusUnprocessableEntity, userFormFragment,
					userFormData(session, partial, commondomain.GetAllowedRoles(), true,
						helper.LocalizedMessage(session, helper.InvalidFormDataMessageID)))
				return
			}
			helper.FlashError(c, session, http.StatusBadRequest, invalidFormDataErrMsg, helper.InvalidFormDataMessageID, err)
			c.Redirect(http.StatusFound, userEditPath(c.Param("id")))
			return
		}
		user := mapper.DTOUpdateUserProcessRequestToDomainUser(uint(userID), &request)
		if err = ctrl.service.UpdateUser(c.Request.Context(), user); err != nil {
			ctrl.eventPublisher.Publish(newUserUpdatedEvent(session.Username, uint(userID), &request, false))
			if htmx {
				helper.LogError(c, updateUserErrMsg, err)
				partial := &dto.User{ID: uint(userID), Username: request.Username, Roles: request.Roles}
				c.HTML(http.StatusUnprocessableEntity, userFormFragment,
					userFormData(session, partial, commondomain.GetAllowedRoles(), true,
						helper.LocalizedMessage(session, helper.UnexpectedErrorMessageID)))
				return
			}
			helper.FlashError(c, session, http.StatusInternalServerError, updateUserErrMsg, helper.UnexpectedErrorMessageID, err)
			c.Redirect(http.StatusFound, userEditPath(c.Param("id")))
			return
		}
		ctrl.eventPublisher.Publish(newUserUpdatedEvent(session.Username, uint(userID), &request, true))
		if htmx {
			helper.TriggerCloseModal(c)
			c.HTML(http.StatusOK, userRowFragment,
				userRowData(session, mapper.DomainUserToDTOUser(user), updateUserOOB))
			return
		}
		session.AddAlertMessages(
			commondomain.NewSuccessAlertMessage(helper.LocalizedMessage(session, userUpdatedMessageID), ""))
		c.Redirect(http.StatusFound, userListPath)
	}
}
