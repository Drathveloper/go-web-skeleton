package service_test

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/argon2"

	"github.com/Drathveloper/go-web-skeleton/common/database"
	commondomain "github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/mocks"
	"github.com/Drathveloper/go-web-skeleton/security/domain"
	"github.com/Drathveloper/go-web-skeleton/security/service"
)

// expectedSaltLength is the salt size the service draws for every new password.
const expectedSaltLength = 16

type UserManagementServiceTestSuite struct {
	suite.Suite

	mockRepository *mocks.MockUserManagementRepository

	service *service.UserManagement

	username string
}

func TestUserManagementServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserManagementServiceTestSuite))
}

func (suite *UserManagementServiceTestSuite) SetupTest() {
	ctrl := gomock.NewController(suite.T())

	suite.mockRepository = mocks.NewMockUserManagementRepository(ctrl)

	suite.service = service.NewUserManagement(suite.mockRepository)

	suite.username = "someUsername"
}

func (suite *UserManagementServiceTestSuite) TestListUser_ShouldSucceed() {
	pagination := &commondomain.Pagination{}
	user := domain.User{
		Username: "username",
		Password: "password",
		Roles:    []commondomain.Role{"someRole"},
	}

	suite.mockRepository.EXPECT().FindAllUsers(suite.T().Context(), pagination).Return([]domain.User{user}, nil)

	users, err := suite.service.ListUsers(suite.T().Context(), pagination)

	suite.Require().NoError(err)
	suite.Require().Equal([]domain.User{user}, users)
}

func (suite *UserManagementServiceTestSuite) TestListUser_ShouldReturnErrorWhenRepositoryFailed() {
	pagination := &commondomain.Pagination{}
	listErr := errors.New("someErr")
	expectedErrMsg := "list users service failed: someErr"

	suite.mockRepository.EXPECT().FindAllUsers(suite.T().Context(), pagination).Return(nil, listErr)

	users, err := suite.service.ListUsers(suite.T().Context(), pagination)

	suite.Require().Nil(users)
	suite.Require().Error(err)
	suite.Require().Equal(expectedErrMsg, err.Error())
}

func (suite *UserManagementServiceTestSuite) TestListUserLookups_ShouldSucceed() {
	users := []domain.User{
		{ID: 1, Username: "alice", Roles: []commondomain.Role{}},
		{ID: 2, Username: "bob", Roles: []commondomain.Role{}},
	}

	suite.mockRepository.EXPECT().FindAllUserLookups(suite.T().Context()).Return(users, nil)

	got, err := suite.service.ListUserLookups(suite.T().Context())

	suite.Require().NoError(err)
	suite.Require().Equal(users, got)
}

func (suite *UserManagementServiceTestSuite) TestListUserLookups_ShouldReturnErrorWhenRepositoryFailed() {
	listErr := errors.New("someErr")
	expectedErrMsg := "list users service failed: someErr"

	suite.mockRepository.EXPECT().FindAllUserLookups(suite.T().Context()).Return(nil, listErr)

	got, err := suite.service.ListUserLookups(suite.T().Context())

	suite.Require().Nil(got)
	suite.Require().Error(err)
	suite.Require().Equal(expectedErrMsg, err.Error())
}

func (suite *UserManagementServiceTestSuite) TestCreateUser_ShouldSucceed() {
	password := "1234"
	roles := []commondomain.Role{"someRole"}
	user := &domain.User{
		Username: suite.username,
		Password: password,
		Roles:    roles,
	}

	suite.mockRepository.EXPECT().CreateUser(suite.T().Context(), gomock.Any()).DoAndReturn(
		func(_ context.Context, userToCreate *domain.User) error {
			suite.Require().Equal(suite.username, userToCreate.Username)
			suite.Require().NotEqual(password, userToCreate.Password)
			suite.Require().Equal(roles, userToCreate.Roles)
			suite.requireStoredPasswordIs(password, userToCreate.Password)

			return nil
		})

	err := suite.service.CreateUser(suite.T().Context(), user)

	suite.Require().NoError(err)
	suite.requireStoredPasswordIs(password, user.Password)
}

// TestCreateUser_ShouldSaltEveryPasswordIndependently: two accounts that share a password
// must not share a hash, otherwise the store leaks which users picked the same one.
func (suite *UserManagementServiceTestSuite) TestCreateUser_ShouldSaltEveryPasswordIndependently() {
	const password = "1234"

	stored := make([]string, 0, 2)

	suite.mockRepository.EXPECT().CreateUser(suite.T().Context(), gomock.Any()).DoAndReturn(
		func(_ context.Context, userToCreate *domain.User) error {
			stored = append(stored, userToCreate.Password)

			return nil
		}).Times(2)

	suite.Require().NoError(suite.service.CreateUser(
		suite.T().Context(), &domain.User{Username: "first", Password: password}))
	suite.Require().NoError(suite.service.CreateUser(
		suite.T().Context(), &domain.User{Username: "second", Password: password}))

	suite.Require().Len(stored, 2)
	suite.Require().NotEqual(stored[0], stored[1])
	suite.requireStoredPasswordIs(password, stored[0])
	suite.requireStoredPasswordIs(password, stored[1])
}

func (suite *UserManagementServiceTestSuite) TestCreateUser_ShouldReturnErrorWhenRepositoryFailed() {
	password := "1234"
	roles := []commondomain.Role{"someRole"}
	user := &domain.User{
		Username: suite.username,
		Password: password,
		Roles:    roles,
	}
	expectedErrMsg := "create user service failed: someErr"

	suite.mockRepository.EXPECT().CreateUser(suite.T().Context(), gomock.Any()).DoAndReturn(
		func(_ context.Context, userToCreate *domain.User) error {
			suite.Require().Equal(suite.username, userToCreate.Username)
			suite.Require().NotEqual(password, userToCreate.Password)
			suite.Require().Equal(roles, userToCreate.Roles)

			return errors.New("someErr")
		})

	err := suite.service.CreateUser(suite.T().Context(), user)

	suite.Require().Error(err)
	suite.Require().Equal(expectedErrMsg, err.Error())
}

func (suite *UserManagementServiceTestSuite) TestUpdateUser_ShouldSucceed() {
	password := "1234"
	roles := []commondomain.Role{"someRole"}
	user := &domain.User{
		ID:       1,
		Username: suite.username,
		Password: password,
		Roles:    roles,
	}

	suite.mockRepository.EXPECT().UpdateUser(suite.T().Context(), gomock.Any()).DoAndReturn(
		func(_ context.Context, userToUpdate *domain.User) error {
			suite.Require().Equal(uint(1), userToUpdate.ID)
			suite.Require().Equal(suite.username, userToUpdate.Username)
			suite.Require().Equal(roles, userToUpdate.Roles)
			suite.requireStoredPasswordIs(password, userToUpdate.Password)

			return nil
		})

	err := suite.service.UpdateUser(suite.T().Context(), user)

	suite.Require().NoError(err)
}

// TestUpdateUser_ShouldLeaveTheStoredPasswordUntouchedWhenBlank: an empty password field
// on the edit form means "keep the current one", so it must reach the repository empty
// rather than being hashed into a value nobody knows.
func (suite *UserManagementServiceTestSuite) TestUpdateUser_ShouldLeaveTheStoredPasswordUntouchedWhenBlank() {
	user := &domain.User{
		ID:       1,
		Username: suite.username,
		Password: "",
		Roles:    []commondomain.Role{"someRole"},
	}

	suite.mockRepository.EXPECT().UpdateUser(suite.T().Context(), gomock.Any()).DoAndReturn(
		func(_ context.Context, userToUpdate *domain.User) error {
			suite.Require().Empty(userToUpdate.Password)

			return nil
		})

	err := suite.service.UpdateUser(suite.T().Context(), user)

	suite.Require().NoError(err)
	suite.Require().Empty(user.Password)
}

func (suite *UserManagementServiceTestSuite) TestUpdateUser_ShouldReturnErrorWhenRepositoryFailed() {
	password := "1234"
	roles := []commondomain.Role{"someRole"}
	user := &domain.User{
		ID:       1,
		Username: suite.username,
		Password: password,
		Roles:    roles,
	}
	updateErr := errors.New("someErr")
	expectedErrMsg := "update user service failed: someErr"

	suite.mockRepository.EXPECT().UpdateUser(suite.T().Context(), gomock.Any()).Return(updateErr)

	err := suite.service.UpdateUser(suite.T().Context(), user)

	suite.Require().Error(err)
	suite.Require().Equal(expectedErrMsg, err.Error())
}

func (suite *UserManagementServiceTestSuite) TestGetUserByID_ShouldSucceed() {
	suite.mockRepository.EXPECT().FindUserByID(suite.T().Context(), uint(1)).Return(&domain.User{ID: 1}, nil)

	user, err := suite.service.GetUserByID(suite.T().Context(), 1)

	suite.Require().NoError(err)
	suite.Require().Equal(&domain.User{ID: 1}, user)
}

// TestGetUserByID_ShouldReturnErrorWhenRecordNotFound: the database-level "no rows"
// becomes domain.ErrUserNotFound, which is what lets the handler answer 404 instead of
// 500 without importing the service package.
func (suite *UserManagementServiceTestSuite) TestGetUserByID_ShouldReturnErrorWhenRecordNotFound() {
	expectedErrMsg := "get user by id service failed: user not found"
	suite.mockRepository.EXPECT().
		FindUserByID(suite.T().Context(), uint(1)).Return(nil, database.ErrRecordNotFound)

	user, err := suite.service.GetUserByID(suite.T().Context(), 1)

	suite.Require().Nil(user)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, domain.ErrUserNotFound)
	suite.Require().NotErrorIs(err, database.ErrRecordNotFound)
	suite.Require().Equal(expectedErrMsg, err.Error())
}

func (suite *UserManagementServiceTestSuite) TestGetUserByID_ShouldReturnErrorWhenRepositoryFailed() {
	getErr := errors.New("someErr")
	expectedErrMsg := "get user by id service failed: someErr"
	suite.mockRepository.EXPECT().FindUserByID(suite.T().Context(), uint(1)).Return(nil, getErr)

	user, err := suite.service.GetUserByID(suite.T().Context(), 1)

	suite.Require().Nil(user)
	suite.Require().Error(err)
	suite.Require().NotErrorIs(err, domain.ErrUserNotFound)
	suite.Require().Equal(expectedErrMsg, err.Error())
}

// requireStoredPasswordIs re-derives the plaintext with the parameters the stored PHC
// string declares and compares the result, so the assertion is that the value really is
// the argon2id hash of that password and not merely that it looks like one.
func (suite *UserManagementServiceTestSuite) requireStoredPasswordIs(plaintext, stored string) {
	suite.T().Helper()

	expectedPrefix := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$", argon2.Version, currentMemory, currentTime, currentParallelism)
	suite.Require().True(
		strings.HasPrefix(stored, expectedPrefix),
		"stored password %q does not start with %q", stored, expectedPrefix)

	fields := strings.Split(stored, "$")
	suite.Require().Len(fields, 6)

	salt, err := base64.RawStdEncoding.Strict().DecodeString(fields[4])
	suite.Require().NoError(err)
	suite.Require().Len(salt, expectedSaltLength)

	hash, err := base64.RawStdEncoding.Strict().DecodeString(fields[5])
	suite.Require().NoError(err)
	suite.Require().Len(hash, currentKeyLength)

	expectedHash := argon2.IDKey([]byte(plaintext), salt, currentTime, currentMemory, currentParallelism, currentKeyLength)
	suite.Require().Equal(expectedHash, hash)
}
