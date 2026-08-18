package service_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/Drathveloper/go-web-skeleton/common/database"
	commondomain "github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/mocks"
	"github.com/Drathveloper/go-web-skeleton/security/domain"
	"github.com/Drathveloper/go-web-skeleton/security/service"
)

type SessionServiceTestSuite struct {
	suite.Suite

	repository *mocks.MockSessionRepository
	service    *service.Session

	sessionID string
}

func TestSessionServiceTestSuite(t *testing.T) {
	suite.Run(t, new(SessionServiceTestSuite))
}

func (suite *SessionServiceTestSuite) SetupTest() {
	ctrl := gomock.NewController(suite.T())
	suite.repository = mocks.NewMockSessionRepository(ctrl)

	suite.service = service.NewSession(suite.repository)

	suite.sessionID = "someSessionID"
}

func (suite *SessionServiceTestSuite) TestLoadSession_ShouldSucceedWhenSessionFoundInRepository() {
	expectedSession := &commondomain.Session{
		ID:            suite.sessionID,
		AlertMessages: make(commondomain.AlertMessages, 0),
		Roles:         make([]commondomain.Role, 0),
	}

	suite.repository.EXPECT().Get(suite.T().Context(), suite.sessionID).Return(expectedSession, nil)

	session, err := suite.service.LoadSession(suite.T().Context(), suite.sessionID)

	suite.Require().NoError(err)
	suite.Require().Equal(expectedSession, session)
}

func (suite *SessionServiceTestSuite) TestLoadSession_ShouldSucceedWhenSessionNotFoundInRepository() {
	expectedSession := &commondomain.Session{
		ID:            suite.sessionID,
		AlertMessages: make(commondomain.AlertMessages, 0),
		Roles:         make([]commondomain.Role, 0),
	}

	suite.repository.EXPECT().Get(suite.T().Context(), suite.sessionID).Return(nil, database.ErrRecordNotFound)
	suite.repository.EXPECT().Save(suite.T().Context(), expectedSession).Return(nil)

	session, err := suite.service.LoadSession(suite.T().Context(), suite.sessionID)

	suite.Require().NoError(err)
	suite.Require().Equal(expectedSession, session)
}

func (suite *SessionServiceTestSuite) TestLoadSession_ShouldReturnErrorWhenSessionNotFoundInRepositoryAndRepositoryFailed() {
	getErr := errors.New("someErr")
	expectedErrMsg := "load session service failed: someErr"

	suite.repository.EXPECT().Get(suite.T().Context(), suite.sessionID).Return(nil, getErr)

	session, err := suite.service.LoadSession(suite.T().Context(), suite.sessionID)

	suite.Require().Nil(session)
	suite.Require().Error(err)
	suite.Require().Equal(expectedErrMsg, err.Error())
}

func (suite *SessionServiceTestSuite) TestLoadSession_ShouldReturnErrorWhenSessionNotFoundInRepositoryAndSaveFailed() {
	expectedSession := &commondomain.Session{
		ID:            suite.sessionID,
		AlertMessages: make(commondomain.AlertMessages, 0),
		Roles:         make([]commondomain.Role, 0),
	}
	saveErr := errors.New("someErr")
	expectedErrMsg := "load session service failed: create anonymous session service failed: someErr"

	suite.repository.EXPECT().Get(suite.T().Context(), suite.sessionID).Return(nil, database.ErrRecordNotFound)
	suite.repository.EXPECT().Save(suite.T().Context(), expectedSession).Return(saveErr)

	session, err := suite.service.LoadSession(suite.T().Context(), suite.sessionID)

	suite.Require().Nil(session)
	suite.Require().Error(err)
	suite.Require().Equal(expectedErrMsg, err.Error())
}

func (suite *SessionServiceTestSuite) TestCreateAnonymousSession_ShouldSucceed() {
	expectedSession := &commondomain.Session{
		ID:            suite.sessionID,
		AlertMessages: make(commondomain.AlertMessages, 0),
		Roles:         make([]commondomain.Role, 0),
	}

	suite.repository.EXPECT().Save(suite.T().Context(), expectedSession).Return(nil)

	session, err := suite.service.CreateAnonymousSession(suite.T().Context(), suite.sessionID)

	suite.Require().NoError(err)
	suite.Require().Equal(expectedSession, session)
}

func (suite *SessionServiceTestSuite) TestCreateAnonymousSession_ShouldReturnErrorWhenSaveFailed() {
	expectedSession := &commondomain.Session{
		ID:            suite.sessionID,
		AlertMessages: make(commondomain.AlertMessages, 0),
		Roles:         make([]commondomain.Role, 0),
	}
	saveErr := errors.New("someErr")
	expectedErrMsg := "create anonymous session service failed: someErr"

	suite.repository.EXPECT().Save(suite.T().Context(), expectedSession).Return(saveErr)

	session, err := suite.service.CreateAnonymousSession(suite.T().Context(), suite.sessionID)

	suite.Require().Nil(session)
	suite.Require().Error(err)
	suite.Require().Equal(expectedErrMsg, err.Error())
}

// TestCreateUserSession_ShouldRotateTheSessionID is the session fixation guard: the
// authenticated session is stored under a brand new identifier, and the anonymous entry
// the visitor arrived with is dropped instead of surviving for the whole TTL.
func (suite *SessionServiceTestSuite) TestCreateUserSession_ShouldRotateTheSessionID() {
	const (
		anonymousSessionID = "anonymousSessionID"
		userSessionID      = "userSessionID"
	)

	user := &domain.User{
		Username: "joseluis",
		Roles:    []commondomain.Role{commondomain.AdminRole},
		ID:       1,
	}
	expectedSession := &commondomain.Session{
		ID:            userSessionID,
		UserID:        1,
		Username:      "joseluis",
		AlertMessages: make(commondomain.AlertMessages, 0),
		Roles:         []commondomain.Role{commondomain.AdminRole},
	}

	// The delete comes first: the opposite order can leave an authenticated session
	// stored that nobody holds a cookie for.
	gomock.InOrder(
		suite.repository.EXPECT().Delete(suite.T().Context(), anonymousSessionID).Return(nil),
		suite.repository.EXPECT().Save(suite.T().Context(), expectedSession).Return(nil),
	)

	session, err := suite.service.CreateUserSession(suite.T().Context(), anonymousSessionID, userSessionID, user)

	suite.Require().NoError(err)
	suite.Require().Equal(expectedSession, session)
	suite.Require().NotEqual(anonymousSessionID, session.ID)
}

// TestCreateUserSession_ShouldNotSaveWhenDroppingTheAnonymousSessionFailed: if the
// anonymous entry cannot be dropped, no authenticated session may be written at all,
// otherwise the pre-authentication identifier stays valid alongside the new one.
func (suite *SessionServiceTestSuite) TestCreateUserSession_ShouldNotSaveWhenDroppingTheAnonymousSessionFailed() {
	const (
		anonymousSessionID = "anonymousSessionID"
		userSessionID      = "userSessionID"
	)

	user := &domain.User{
		Username: "joseluis",
		Roles:    []commondomain.Role{commondomain.AdminRole},
		ID:       1,
	}
	deleteErr := errors.New("someErr")
	expectedErrMsg := "create user session service failed: someErr"

	suite.repository.EXPECT().Delete(suite.T().Context(), anonymousSessionID).Return(deleteErr)
	suite.repository.EXPECT().Save(gomock.Any(), gomock.Any()).Times(0)

	session, err := suite.service.CreateUserSession(suite.T().Context(), anonymousSessionID, userSessionID, user)

	suite.Require().Nil(session)
	suite.Require().Error(err)
	suite.Require().Equal(expectedErrMsg, err.Error())
}

func (suite *SessionServiceTestSuite) TestCreateUserSession_ShouldNotDeleteWhenThereIsNoAnonymousSession() {
	user := &domain.User{
		Username: "joseluis",
		Roles:    []commondomain.Role{commondomain.AdminRole},
		ID:       1,
	}
	expectedSession := &commondomain.Session{
		ID:            suite.sessionID,
		UserID:        1,
		Username:      "joseluis",
		AlertMessages: make(commondomain.AlertMessages, 0),
		Roles:         []commondomain.Role{commondomain.AdminRole},
	}

	suite.repository.EXPECT().Delete(gomock.Any(), gomock.Any()).Times(0)
	suite.repository.EXPECT().Save(suite.T().Context(), expectedSession).Return(nil)

	session, err := suite.service.CreateUserSession(suite.T().Context(), "", suite.sessionID, user)

	suite.Require().NoError(err)
	suite.Require().Equal(expectedSession, session)
}

// TestCreateUserSession_ShouldNotDeleteTheSessionItIsAboutToWrite covers the caller that
// passes the same identifier twice: deleting it would remove the entry just promoted.
func (suite *SessionServiceTestSuite) TestCreateUserSession_ShouldNotDeleteTheSessionItIsAboutToWrite() {
	user := &domain.User{
		Username: "joseluis",
		Roles:    []commondomain.Role{commondomain.AdminRole},
		ID:       1,
	}
	expectedSession := &commondomain.Session{
		ID:            suite.sessionID,
		UserID:        1,
		Username:      "joseluis",
		AlertMessages: make(commondomain.AlertMessages, 0),
		Roles:         []commondomain.Role{commondomain.AdminRole},
	}

	suite.repository.EXPECT().Delete(gomock.Any(), gomock.Any()).Times(0)
	suite.repository.EXPECT().Save(suite.T().Context(), expectedSession).Return(nil)

	session, err := suite.service.CreateUserSession(suite.T().Context(), suite.sessionID, suite.sessionID, user)

	suite.Require().NoError(err)
	suite.Require().Equal(expectedSession, session)
}

func (suite *SessionServiceTestSuite) TestCreateUserSession_ShouldReturnErrorWhenSaveFailed() {
	user := &domain.User{
		Username: "joseluis",
		Roles:    []commondomain.Role{commondomain.AdminRole},
		ID:       1,
	}
	expectedSession := &commondomain.Session{
		ID:            suite.sessionID,
		UserID:        1,
		Username:      "joseluis",
		AlertMessages: make(commondomain.AlertMessages, 0),
		Roles:         []commondomain.Role{commondomain.AdminRole},
	}
	saveErr := errors.New("someErr")
	expectedErrMsg := "create user session service failed: someErr"

	suite.repository.EXPECT().Save(suite.T().Context(), expectedSession).Return(saveErr)

	session, err := suite.service.CreateUserSession(suite.T().Context(), "", suite.sessionID, user)

	suite.Require().Nil(session)
	suite.Require().Error(err)
	suite.Require().Equal(expectedErrMsg, err.Error())
}

func (suite *SessionServiceTestSuite) TestUpdateSession_ShouldSucceed() {
	session := &commondomain.Session{
		ID: suite.sessionID,
	}

	suite.repository.EXPECT().Save(suite.T().Context(), session).Return(nil)

	err := suite.service.UpdateSession(suite.T().Context(), session)

	suite.Require().NoError(err)
}

func (suite *SessionServiceTestSuite) TestUpdateSession_ShouldReturnErrorWhenSaveFailed() {
	session := &commondomain.Session{
		ID: suite.sessionID,
	}
	saveErr := errors.New("someErr")
	expectedErrMsg := "update session service failed: someErr"

	suite.repository.EXPECT().Save(suite.T().Context(), session).Return(saveErr)

	err := suite.service.UpdateSession(suite.T().Context(), session)

	suite.Require().Error(err)
	suite.Require().Equal(expectedErrMsg, err.Error())
}

func (suite *SessionServiceTestSuite) TestDestroySession_ShouldSucceed() {
	session := &commondomain.Session{
		ID: suite.sessionID,
	}

	suite.repository.EXPECT().Delete(suite.T().Context(), session.ID).Return(nil)

	err := suite.service.DestroySession(suite.T().Context(), session)

	suite.Require().NoError(err)
}

func (suite *SessionServiceTestSuite) TestDestroySession_ShouldReturnErrorWhenDeleteFailed() {
	session := &commondomain.Session{
		ID: suite.sessionID,
	}
	deleteErr := errors.New("someErr")
	expectedErrMsg := "destroy session service failed: someErr"

	suite.repository.EXPECT().Delete(suite.T().Context(), session.ID).Return(deleteErr)

	err := suite.service.DestroySession(suite.T().Context(), session)

	suite.Require().Error(err)
	suite.Require().Equal(expectedErrMsg, err.Error())
}
