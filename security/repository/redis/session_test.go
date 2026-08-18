package redis_test

import (
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/Drathveloper/go-web-skeleton/common/database"
	"github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/mocks"
	redisRepository "github.com/Drathveloper/go-web-skeleton/security/repository/redis"
)

const sessionKeyPrefix = "session:"

// expectedSessionKey rebuilds the key the repository must derive: the base64 of the
// SHA-512 *digest* of the prefixed session ID.
//
// It is deliberately not written as sha512.New().Sum([]byte(prefix+id)), which is what
// this repository used to do. hash.Hash.Sum appends the current digest to its argument
// instead of hashing it, so that expression returns the plaintext key followed by the
// digest of the empty string — the session ID, a bearer credential, sat in clear text in
// the keyspace. TestSessionKey_IsADigestAndNotThePlaintext pins that it cannot come back.
func expectedSessionKey(sessionID string) string {
	sum := sha512.Sum512([]byte(sessionKeyPrefix + sessionID))
	return base64.StdEncoding.EncodeToString(sum[:])
}

type SessionRepositoryTestSuite struct {
	suite.Suite

	mockSessionConfigProvider *mocks.MockRedisSessionConfigProvider
	mockUniversalClient       *mocks.MockUniversalClient
	repository                *redisRepository.Session

	sessionID  string
	hashedKey  string
	sessionTTL time.Duration
}

func TestSessionRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(SessionRepositoryTestSuite))
}

func (suite *SessionRepositoryTestSuite) SetupTest() {
	mockCtrl := gomock.NewController(suite.T())

	suite.mockSessionConfigProvider = mocks.NewMockRedisSessionConfigProvider(mockCtrl)
	suite.mockUniversalClient = mocks.NewMockUniversalClient(mockCtrl)

	suite.sessionID = "someKey"
	suite.hashedKey = expectedSessionKey(suite.sessionID)
	suite.sessionTTL = 24 * time.Hour
	suite.mockSessionConfigProvider.EXPECT().GetSessionTTL().Return(suite.sessionTTL)

	suite.repository = redisRepository.NewSession(suite.mockUniversalClient, suite.mockSessionConfigProvider)
}

// TestSessionKey_IsADigestAndNotThePlaintext is the regression guard for the key
// derivation. Anyone who can list the keyspace must not be able to read a session ID
// back out of it, so the key has to be the digest and nothing else.
func (suite *SessionRepositoryTestSuite) TestSessionKey_IsADigestAndNotThePlaintext() {
	brokenKey := base64.StdEncoding.EncodeToString(sha512.New().Sum([]byte(sessionKeyPrefix + suite.sessionID)))
	resultCmd := redis.NewIntCmd(suite.T().Context())

	var usedKey string

	suite.mockUniversalClient.EXPECT().Del(suite.T().Context(), gomock.Any()).
		DoAndReturn(func(_ any, keys ...string) *redis.IntCmd {
			suite.Require().Len(keys, 1)
			usedKey = keys[0]

			return resultCmd
		})

	err := suite.repository.Delete(suite.T().Context(), suite.sessionID)

	suite.Require().NoError(err)
	suite.Require().Equal(suite.hashedKey, usedKey)

	decoded, decodeErr := base64.StdEncoding.DecodeString(usedKey)
	suite.Require().NoError(decodeErr)
	suite.Require().Len(decoded, sha512.Size, "the key must be a bare SHA-512 digest")

	suite.Require().NotContains(usedKey, suite.sessionID, "the raw session id must not survive in the key")
	suite.Require().NotContains(
		string(decoded), suite.sessionID, "the raw session id must not survive in the decoded key")
	suite.Require().False(strings.HasPrefix(usedKey, base64.StdEncoding.EncodeToString([]byte(sessionKeyPrefix))))
	suite.Require().NotEqual(brokenKey, usedKey, "hash.Hash.Sum appends the digest, it does not hash")
}

// TestSessionKey_DiffersPerSession keeps the derivation from collapsing to a constant,
// which a "hash" that ignores its input would do.
func (suite *SessionRepositoryTestSuite) TestSessionKey_DiffersPerSession() {
	otherSessionID := "someOtherKey"
	resultCmd := redis.NewIntCmd(suite.T().Context())

	suite.mockUniversalClient.EXPECT().Del(suite.T().Context(), suite.hashedKey).Return(resultCmd)
	suite.mockUniversalClient.EXPECT().Del(suite.T().Context(), expectedSessionKey(otherSessionID)).Return(resultCmd)

	suite.Require().NoError(suite.repository.Delete(suite.T().Context(), suite.sessionID))
	suite.Require().NoError(suite.repository.Delete(suite.T().Context(), otherSessionID))

	suite.Require().NotEqual(expectedSessionKey(otherSessionID), suite.hashedKey)
}

func (suite *SessionRepositoryTestSuite) TestGet_ShouldSucceed() {
	resultCmd := redis.NewStringCmd(suite.T().Context())
	resultCmd.SetVal(`{"user_id":1234}`)
	expectedSession := &domain.Session{
		ID:            suite.sessionID,
		UserID:        1234,
		Roles:         []domain.Role{},
		AlertMessages: domain.AlertMessages{},
	}

	suite.mockUniversalClient.EXPECT().Get(suite.T().Context(), suite.hashedKey).Return(resultCmd)

	session, err := suite.repository.Get(suite.T().Context(), suite.sessionID)

	suite.Require().NoError(err)
	suite.Require().Equal(expectedSession, session)
}

func (suite *SessionRepositoryTestSuite) TestGet_ShouldReturnErrorWhenUnmarshalFailed() {
	resultCmd := redis.NewStringCmd(suite.T().Context())
	resultCmd.SetVal(`{"user_id":"1234"}`)
	expectedErrMsg := "repository get session failed: json: cannot unmarshal string " +
		"into Go struct field Session.user_id of type uint"

	suite.mockUniversalClient.EXPECT().Get(suite.T().Context(), suite.hashedKey).Return(resultCmd)

	session, err := suite.repository.Get(suite.T().Context(), suite.sessionID)

	suite.Require().Nil(session)
	suite.Require().Error(err)
	suite.Require().Equal(expectedErrMsg, err.Error())
}

func (suite *SessionRepositoryTestSuite) TestGet_ShouldReturnErrorWhenRecordNotFound() {
	resultCmd := redis.NewStringCmd(suite.T().Context())
	resultCmd.SetErr(redis.Nil)
	expectedErrMsg := "repository get session failed: database record not found"

	suite.mockUniversalClient.EXPECT().Get(suite.T().Context(), suite.hashedKey).Return(resultCmd)

	session, err := suite.repository.Get(suite.T().Context(), suite.sessionID)

	suite.Require().Nil(session)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, database.ErrRecordNotFound)
	suite.Require().Equal(expectedErrMsg, err.Error())
}

func (suite *SessionRepositoryTestSuite) TestGet_ShouldReturnErrorWhenUnexpectedError() {
	resultCmd := redis.NewStringCmd(suite.T().Context())
	resultCmd.SetErr(errors.New("someErr"))
	expectedErrMsg := "repository get session failed: someErr"

	suite.mockUniversalClient.EXPECT().Get(suite.T().Context(), suite.hashedKey).Return(resultCmd)

	session, err := suite.repository.Get(suite.T().Context(), suite.sessionID)

	suite.Require().Nil(session)
	suite.Require().Error(err)
	suite.Require().NotErrorIs(err, database.ErrRecordNotFound)
	suite.Require().Equal(expectedErrMsg, err.Error())
}

func (suite *SessionRepositoryTestSuite) TestSave_ShouldSucceed() {
	value := []byte(`{"user_id":1234}`)
	resultCmd := redis.NewStatusCmd(suite.T().Context())
	session := &domain.Session{
		ID:            suite.sessionID,
		UserID:        1234,
		Roles:         []domain.Role{},
		AlertMessages: domain.AlertMessages{},
	}

	suite.mockUniversalClient.EXPECT().
		Set(suite.T().Context(), suite.hashedKey, value, suite.sessionTTL).Return(resultCmd)

	err := suite.repository.Save(suite.T().Context(), session)

	suite.Require().NoError(err)
}

// TestSave_ShouldNotStoreTheSessionIDInTheValue keeps the ID out of the stored payload:
// it is the key material, and writing it into the value would put it back in clear text.
func (suite *SessionRepositoryTestSuite) TestSave_ShouldNotStoreTheSessionIDInTheValue() {
	resultCmd := redis.NewStatusCmd(suite.T().Context())
	session := &domain.Session{
		ID:                   suite.sessionID,
		UserID:               1234,
		Username:             "johndoe",
		Roles:                []domain.Role{domain.AdminRole},
		CSRFToken:            "csrf",
		Language:             "es",
		IsLanguageOverridden: true,
	}

	var storedValue []byte

	suite.mockUniversalClient.EXPECT().
		Set(suite.T().Context(), suite.hashedKey, gomock.Any(), suite.sessionTTL).
		DoAndReturn(func(_ any, _ string, value any, _ ...any) *redis.StatusCmd {
			raw, ok := value.([]byte)
			suite.Require().True(ok)
			storedValue = raw

			return resultCmd
		})

	err := suite.repository.Save(suite.T().Context(), session)

	suite.Require().NoError(err)
	suite.Require().NotContains(string(storedValue), suite.sessionID)
	suite.Require().Contains(string(storedValue), `"username":"johndoe"`)
	suite.Require().Contains(string(storedValue), `"is_language_overridden":true`)
}

func (suite *SessionRepositoryTestSuite) TestSave_ShouldReturnErrorWhenUnexpectedError() {
	value := []byte(`{"user_id":1234}`)
	resultCmd := redis.NewStatusCmd(suite.T().Context())
	resultCmd.SetErr(errors.New("someErr"))
	session := &domain.Session{
		ID:            suite.sessionID,
		UserID:        1234,
		Roles:         []domain.Role{},
		AlertMessages: domain.AlertMessages{},
	}
	expectedErrMsg := "repository save session failed: someErr"

	suite.mockUniversalClient.EXPECT().
		Set(suite.T().Context(), suite.hashedKey, value, suite.sessionTTL).Return(resultCmd)

	err := suite.repository.Save(suite.T().Context(), session)

	suite.Require().Error(err)
	suite.Require().Equal(expectedErrMsg, err.Error())
}

func (suite *SessionRepositoryTestSuite) TestDelete_ShouldSucceed() {
	resultCmd := redis.NewIntCmd(suite.T().Context())

	suite.mockUniversalClient.EXPECT().Del(suite.T().Context(), suite.hashedKey).Return(resultCmd)

	err := suite.repository.Delete(suite.T().Context(), suite.sessionID)

	suite.Require().NoError(err)
}

func (suite *SessionRepositoryTestSuite) TestDelete_ShouldReturnErrorWhenUnexpectedError() {
	resultCmd := redis.NewIntCmd(suite.T().Context())
	resultCmd.SetErr(errors.New("someErr"))
	expectedErrMsg := "repository delete session failed: someErr"

	suite.mockUniversalClient.EXPECT().Del(suite.T().Context(), suite.hashedKey).Return(resultCmd)

	err := suite.repository.Delete(suite.T().Context(), suite.sessionID)

	suite.Require().Error(err)
	suite.Require().Equal(expectedErrMsg, err.Error())
}
