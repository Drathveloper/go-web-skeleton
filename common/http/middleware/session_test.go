package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
	"github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/common/http/helper"
	"github.com/Drathveloper/go-web-skeleton/common/http/middleware"
	"github.com/Drathveloper/go-web-skeleton/mocks"
)

// errInfrastructure stands for what the session service really fails with. None
// of it may reach the browser.
var errInfrastructure = errors.New("dial tcp 10.0.3.14:6379: connect: connection refused")

type SessionMiddlewareTestSuite struct {
	suite.Suite

	mockSessionService        *mocks.MockSessionService
	mockSessionConfigProvider *mocks.MockSessionConfigProvider
	cookie                    *http.Cookie
	sessionID                 string
}

func TestSessionMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(SessionMiddlewareTestSuite))
}

func (suite *SessionMiddlewareTestSuite) SetupTest() {
	mockCtrl := gomock.NewController(suite.T())

	suite.mockSessionConfigProvider = mocks.NewMockSessionConfigProvider(mockCtrl)
	suite.mockSessionService = mocks.NewMockSessionService(mockCtrl)

	suite.sessionID = "someSessionID"
	middleware.SessionIDGeneratorFunc = func() string { return suite.sessionID }
	suite.T().Cleanup(func() {
		middleware.SessionIDGeneratorFunc = helper.GenerateSessionID
	})
	suite.cookie = &http.Cookie{
		Name:     constants.SessionIDKey,
		Value:    suite.sessionID,
		Path:     "/",
		Domain:   "localhost",
		MaxAge:   3600,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}

func (suite *SessionMiddlewareTestSuite) TestSession_ShouldCreateAnAnonymousSessionWhenNoCookieIsPresent() {
	createdSession := &domain.Session{ID: suite.sessionID}

	suite.mockSessionConfigProvider.EXPECT().GetSessionTTL().Return(3600 * time.Second)
	suite.mockSessionConfigProvider.EXPECT().IsSecureCookie().Return(true)
	suite.mockSessionConfigProvider.EXPECT().GetCookieDomain().Return("localhost")
	suite.mockSessionService.EXPECT().CreateAnonymousSession(gomock.Any(), suite.sessionID).Return(createdSession, nil)
	suite.mockSessionService.EXPECT().UpdateSession(gomock.Any(), createdSession).Return(nil)

	engine, handlerCalled := suite.setupServer(nil)
	recorder := serve(suite.T(), engine, suite.newRequest(false))

	suite.Require().Equal(http.StatusOK, recorder.Code)
	suite.Require().True(*handlerCalled)

	// The cookie is the session: without HttpOnly any script on the page can read
	// it, and without SameSite=Strict it rides along on cross-site requests.
	cookies := (&http.Response{Header: recorder.Header()}).Cookies()
	suite.Require().Len(cookies, 1)
	suite.Require().Equal(constants.SessionIDKey, cookies[0].Name)
	suite.Require().Equal(suite.sessionID, cookies[0].Value)
	suite.Require().True(cookies[0].HttpOnly)
	suite.Require().True(cookies[0].Secure)
	suite.Require().Equal(http.SameSiteStrictMode, cookies[0].SameSite)
	suite.Require().Equal(3600, cookies[0].MaxAge)

	// A page that varies by session must never be cached as if it did not.
	suite.Require().Equal("Cookie", recorder.Header().Get("Vary"))
	suite.Require().Equal(`no-cache="Set-Cookie"`, recorder.Header().Get("Cache-Control"))
}

func (suite *SessionMiddlewareTestSuite) TestSession_ShouldLoadTheSessionWhenTheCookieIsPresent() {
	foundSession := &domain.Session{ID: suite.sessionID, Username: "someUser", UserID: 1}

	suite.mockSessionService.EXPECT().LoadSession(gomock.Any(), suite.sessionID).Return(foundSession, nil)
	suite.mockSessionService.EXPECT().UpdateSession(gomock.Any(), foundSession).Return(nil)

	var seen *domain.Session
	engine, handlerCalled := suite.setupServer(func(c *gin.Context) {
		seen = helper.MustGetSession(c)
	})

	recorder := serve(suite.T(), engine, suite.newRequest(true))

	suite.Require().Equal(http.StatusOK, recorder.Code)
	suite.Require().True(*handlerCalled)
	suite.Require().Same(foundSession, seen, "the handler must see the loaded session")
	// No new cookie: the one the browser sent is still the valid one.
	suite.Require().Empty(recorder.Header().Values("Set-Cookie"))
}

// A handler that replaces the session in the context is doing a login rotation:
// the anonymous entry has been deleted and a new session ID minted. Persisting
// the pointer captured before c.Next() would write the deleted anonymous session
// back to Redis and leave the new one unsaved, which is what made a successful
// login land back on the login page.
func (suite *SessionMiddlewareTestSuite) TestSession_ShouldPersistTheSessionTheHandlerLeftInTheContext() {
	loadedSession := &domain.Session{ID: suite.sessionID}
	rotatedSession := &domain.Session{ID: "rotatedSessionID", Username: "someUser", UserID: 1}

	suite.mockSessionService.EXPECT().LoadSession(gomock.Any(), suite.sessionID).Return(loadedSession, nil)
	suite.mockSessionService.EXPECT().UpdateSession(gomock.Any(), rotatedSession).Return(nil)

	engine, _ := suite.setupServer(func(c *gin.Context) {
		c.Set(constants.SessionGinContextKey, rotatedSession)
	})

	recorder := serve(suite.T(), engine, suite.newRequest(true))

	suite.Require().Equal(http.StatusOK, recorder.Code)
}

// Logout clears the session from the context. There is nothing left to persist,
// and writing the captured pointer back would resurrect the session that was
// just destroyed.
func (suite *SessionMiddlewareTestSuite) TestSession_ShouldNotPersistAnythingWhenTheHandlerClearedTheSession() {
	loadedSession := &domain.Session{ID: suite.sessionID, UserID: 1}

	suite.mockSessionService.EXPECT().LoadSession(gomock.Any(), suite.sessionID).Return(loadedSession, nil)
	// No UpdateSession expectation: gomock fails the test if it is called at all.

	engine, _ := suite.setupServer(func(c *gin.Context) {
		c.Set(constants.SessionGinContextKey, nil)
	})

	recorder := serve(suite.T(), engine, suite.newRequest(true))

	suite.Require().Equal(http.StatusOK, recorder.Code)
}

func (suite *SessionMiddlewareTestSuite) TestSession_ShouldRenderErrorWhenCreateAnonymousSessionFailed() {
	suite.mockSessionService.EXPECT().
		CreateAnonymousSession(gomock.Any(), suite.sessionID).
		Return(nil, errInfrastructure)

	engine, handlerCalled := suite.setupServer(nil)
	recorder := serve(suite.T(), engine, suite.newRequest(false))

	suite.requireErrorPage(recorder)
	suite.Require().False(*handlerCalled, "the request must not reach the handler without a session")
}

func (suite *SessionMiddlewareTestSuite) TestSession_ShouldRenderErrorWhenLoadSessionFailed() {
	suite.mockSessionService.EXPECT().LoadSession(gomock.Any(), suite.sessionID).Return(nil, errInfrastructure)

	engine, handlerCalled := suite.setupServer(nil)
	recorder := serve(suite.T(), engine, suite.newRequest(true))

	suite.requireErrorPage(recorder)
	suite.Require().False(*handlerCalled)
}

func (suite *SessionMiddlewareTestSuite) TestSession_ShouldRenderErrorWhenUpdateSessionFailedOnANewSession() {
	createdSession := &domain.Session{ID: suite.sessionID}

	suite.mockSessionConfigProvider.EXPECT().GetSessionTTL().Return(3600 * time.Second)
	suite.mockSessionConfigProvider.EXPECT().IsSecureCookie().Return(true)
	suite.mockSessionConfigProvider.EXPECT().GetCookieDomain().Return("localhost")
	suite.mockSessionService.EXPECT().CreateAnonymousSession(gomock.Any(), suite.sessionID).Return(createdSession, nil)
	suite.mockSessionService.EXPECT().UpdateSession(gomock.Any(), createdSession).Return(errInfrastructure)

	engine, _ := suite.setupServer(nil)
	recorder := serve(suite.T(), engine, suite.newRequest(false))

	suite.requireErrorPage(recorder)
}

func (suite *SessionMiddlewareTestSuite) TestSession_ShouldRenderErrorWhenUpdateSessionFailedOnALoadedSession() {
	foundSession := &domain.Session{ID: suite.sessionID}

	suite.mockSessionService.EXPECT().LoadSession(gomock.Any(), suite.sessionID).Return(foundSession, nil)
	suite.mockSessionService.EXPECT().UpdateSession(gomock.Any(), foundSession).Return(errInfrastructure)

	engine, _ := suite.setupServer(nil)
	recorder := serve(suite.T(), engine, suite.newRequest(true))

	suite.requireErrorPage(recorder)
}

func (suite *SessionMiddlewareTestSuite) requireErrorPage(recorder *httptest.ResponseRecorder) {
	suite.T().Helper()

	suite.Require().Equal(http.StatusInternalServerError, recorder.Code)
	body := recorder.Body.String()
	suite.Require().Contains(body, "Internal server error.")
	suite.Require().NotContains(body, "10.0.3.14", "the error page must not leak infrastructure detail")
	suite.Require().NotContains(body, "connection refused")
}

func (suite *SessionMiddlewareTestSuite) newRequest(withCookie bool) *http.Request {
	request := httptest.NewRequestWithContext(suite.T().Context(), http.MethodGet, "/test", nil)
	if withCookie {
		request.AddCookie(suite.cookie)
	}

	return request
}

// setupServer wires the middleware in front of a handler that reports whether it
// ran, so an aborted request can be told apart from one that merely failed later.
func (suite *SessionMiddlewareTestSuite) setupServer(handler gin.HandlerFunc) (*gin.Engine, *bool) {
	called := false
	engine := newEngine(suite.T())
	engine.GET("/test",
		middleware.SessionHandler(suite.mockSessionService, suite.mockSessionConfigProvider),
		func(c *gin.Context) {
			called = true
			if handler != nil {
				handler(c)
			}
			c.Status(http.StatusOK)
		},
	)

	return engine, &called
}
