package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
	commondomain "github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/common/http/helper"
	"github.com/Drathveloper/go-web-skeleton/pkg/event"
	"github.com/Drathveloper/go-web-skeleton/security/domain"
	"github.com/Drathveloper/go-web-skeleton/security/event/dto"
	"github.com/Drathveloper/go-web-skeleton/security/http/mapper"
)

const loginPath = "/auth/login"

const (
	loginErrMsg         = "login error"
	createSessionErrMsg = "create session error"
	logoutErrMsg        = "logout error"
)

type SessionConfigProvider interface {
	GetSessionTTL() time.Duration
	IsSecureCookie() bool
	GetCookieDomain() string
}

type AuthenticationService interface {
	Login(ctx context.Context, login *domain.Login) (*domain.User, error)
}

type SessionService interface {
	CreateUserSession(
		ctx context.Context,
		anonymousSessionID, userSessionID string,
		user *domain.User) (*commondomain.Session, error)
	DestroySession(ctx context.Context, session *commondomain.Session) error
}

type EventPublisher interface {
	Publish(event event.Event)
}

type Authentication struct {
	eventPublisher EventPublisher
	authService    AuthenticationService
	sessionService SessionService
	cookieDomain   string
	sessionTTL     time.Duration
	isSecureCookie bool
}

func NewAuthentication(
	authService AuthenticationService,
	sessionService SessionService,
	configProvider SessionConfigProvider,
	eventPublisher EventPublisher) *Authentication {
	return &Authentication{
		eventPublisher: eventPublisher,
		authService:    authService,
		sessionService: sessionService,
		isSecureCookie: configProvider.IsSecureCookie(),
		sessionTTL:     configProvider.GetSessionTTL(),
		cookieDomain:   configProvider.GetCookieDomain(),
	}
}

func (ctrl *Authentication) LoginView() gin.HandlerFunc {
	return func(c *gin.Context) { //nolint:varnamelen
		session := helper.MustGetSession(c)
		if session.UserID != 0 {
			c.Redirect(http.StatusFound, constants.HomePath)
			return
		}
		response := mapper.DomainSessionToLoginViewResponse(session)
		c.HTML(http.StatusOK, "auth/login", response)
	}
}

func (ctrl *Authentication) LoginProcess() gin.HandlerFunc {
	return func(c *gin.Context) { //nolint:varnamelen
		session := helper.MustGetSession(c)
		if err := c.Request.ParseForm(); err != nil {
			flashError(c, session, http.StatusBadRequest, invalidFormDataErrMsg, invalidFormDataMessageID, err)
			c.Redirect(http.StatusFound, loginPath)
			return
		}
		login, err := mapper.FormLoginToDomainLogin(c.Request.Form)
		if err != nil {
			flashError(c, session, http.StatusBadRequest, invalidFormDataErrMsg, invalidFormDataMessageID, err)
			c.Redirect(http.StatusFound, loginPath)
			return
		}
		user, err := ctrl.authService.Login(c.Request.Context(), login)
		if err != nil {
			// An unknown user and a wrong password answer exactly the same thing:
			// telling them apart turns the login form into a user enumeration oracle.
			ctrl.eventPublisher.Publish(&dto.LoginEvent{Username: login.Username, IsSuccess: false})
			flashError(c, session, http.StatusUnauthorized, loginErrMsg, invalidCredentialsMessageID, err)
			c.Redirect(http.StatusFound, loginPath)
			return
		}
		// A brand new session ID is what prevents session fixation; handing the old
		// one over lets the service drop the anonymous entry instead of leaving it
		// orphaned in Redis until its TTL expires, which is a year by default.
		sessionID := helper.GenerateSessionID()
		userSession, err := ctrl.sessionService.CreateUserSession(
			c.Request.Context(), session.ID, sessionID, user)
		if err != nil {
			flashError(c, session, http.StatusInternalServerError, createSessionErrMsg, unexpectedErrorMessageID, err)
			c.Redirect(http.StatusFound, loginPath)
			return
		}
		c.Set(constants.SessionGinContextKey, userSession)
		c.SetSameSite(http.SameSiteStrictMode)
		c.SetCookie(
			constants.SessionIDKey,
			sessionID,
			int(ctrl.sessionTTL.Seconds()),
			"/",
			ctrl.cookieDomain,
			ctrl.isSecureCookie,
			true)
		ctrl.eventPublisher.Publish(&dto.LoginEvent{Username: login.Username, IsSuccess: true})
		c.Redirect(http.StatusFound, constants.HomePath)
	}
}

func (ctrl *Authentication) LogoutProcess() gin.HandlerFunc {
	return func(c *gin.Context) { //nolint:varnamelen
		session := helper.MustGetSession(c)
		if err := ctrl.sessionService.DestroySession(c.Request.Context(), session); err != nil {
			flashError(c, session, http.StatusInternalServerError, logoutErrMsg, unexpectedErrorMessageID, err)
			c.Redirect(http.StatusFound, constants.HomePath)
			return
		}
		c.Set(constants.SessionGinContextKey, nil)
		c.SetSameSite(http.SameSiteStrictMode)
		c.SetCookie(constants.SessionIDKey, "", -1, "/", ctrl.cookieDomain, ctrl.isSecureCookie, true)
		c.Redirect(http.StatusFound, loginPath)
	}
}
