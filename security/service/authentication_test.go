package service_test

import (
	"encoding/base64"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/argon2"

	"github.com/Drathveloper/go-web-skeleton/common/database"
	commondomain "github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/mocks"
	"github.com/Drathveloper/go-web-skeleton/security/domain"
	"github.com/Drathveloper/go-web-skeleton/security/service"
)

// The cost parameters the service hashes with today. A stored hash carries its own
// parameters, so these are only what a freshly created password looks like.
const (
	currentMemory      = 64 * 1024
	currentTime        = 1
	currentParallelism = 1
	currentKeyLength   = 64
)

// timingSlackFactor is how much slower a locally measured KDF run is allowed to be
// than another before the timing assertion gives up. The property under test is
// "the unknown-user path is not a fast path", which a return-early implementation
// misses by three orders of magnitude, so a factor of four is both generous enough
// not to flake on a loaded machine and tight enough to catch the regression.
const timingSlackFactor = 4

// phcHash builds the stored form of a password exactly as the PHC spec and the
// service's encoder do: $argon2id$v=19$m=..,t=..,p=..$salt$hash.
func phcHash(password, salt string, memory, iterations uint32, parallelism uint8, keyLength uint32) string {
	hash := argon2.IDKey([]byte(password), []byte(salt), iterations, memory, parallelism, keyLength)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		memory,
		iterations,
		parallelism,
		base64.RawStdEncoding.EncodeToString([]byte(salt)),
		base64.RawStdEncoding.EncodeToString(hash))
}

type AuthenticationServiceTestSuite struct {
	suite.Suite

	repository *mocks.MockAuthenticationRepository
	service    *service.Authentication
}

func TestAuthenticationServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AuthenticationServiceTestSuite))
}

func (suite *AuthenticationServiceTestSuite) SetupTest() {
	ctrl := gomock.NewController(suite.T())
	suite.repository = mocks.NewMockAuthenticationRepository(ctrl)

	suite.service = service.NewAuthentication(suite.repository)
}

func (suite *AuthenticationServiceTestSuite) TestLogin_ShouldSucceed() {
	const password = "correct horse battery staple"

	login := &domain.Login{
		Username:   "someUsername",
		Password:   password,
		RememberMe: true,
	}
	expectedUser := &domain.User{
		Username: "someUsername",
		Password: phcHash(
			password, "someSaltValue123", currentMemory, currentTime, currentParallelism, currentKeyLength),
		Roles: []commondomain.Role{commondomain.AdminRole},
		ID:    1,
	}

	suite.repository.EXPECT().FindUserByUsername(suite.T().Context(), login.Username).Return(expectedUser, nil)

	user, err := suite.service.Login(suite.T().Context(), login)

	suite.Require().NoError(err)
	suite.Require().Equal(expectedUser, user)
}

// TestLogin_ShouldSucceedWhenStoredHashUsesOtherCostParameters is the reason the cost
// parameters travel inside the PHC string: raising the constants must not lock out every
// password already in the database. This hash is deliberately produced with parameters
// that differ from the current ones in memory, time and key length.
func (suite *AuthenticationServiceTestSuite) TestLogin_ShouldSucceedWhenStoredHashUsesOtherCostParameters() {
	const (
		legacyMemory      = 8 * 1024
		legacyTime        = 3
		legacyParallelism = 2
		legacyKeyLength   = 32
	)

	suite.Require().NotEqual(uint32(currentMemory), uint32(legacyMemory))
	suite.Require().NotEqual(uint32(currentKeyLength), uint32(legacyKeyLength))

	login := &domain.Login{Username: "legacy", Password: "1234"}
	expectedUser := &domain.User{
		Username: "legacy",
		Password: phcHash("1234", "legacySaltValue1", legacyMemory, legacyTime, legacyParallelism, legacyKeyLength),
		Roles:    []commondomain.Role{commondomain.AdminRole},
		ID:       7,
	}

	suite.repository.EXPECT().FindUserByUsername(suite.T().Context(), login.Username).Return(expectedUser, nil)

	user, err := suite.service.Login(suite.T().Context(), login)

	suite.Require().NoError(err)
	suite.Require().Equal(expectedUser, user)
}

func (suite *AuthenticationServiceTestSuite) TestLogin_ShouldReturnErrorWhenPasswordDoesntMatch() {
	login := &domain.Login{
		Username:   "someUsername",
		Password:   "1235",
		RememberMe: true,
	}
	expectedUser := &domain.User{
		Username: "someUsername",
		Password: phcHash("1234", "someSaltValue123", currentMemory, currentTime, currentParallelism, currentKeyLength),
		Roles:    []commondomain.Role{commondomain.AdminRole},
		ID:       1,
	}
	expectedErrMsg := "login service failed: invalid credentials"

	suite.repository.EXPECT().FindUserByUsername(suite.T().Context(), login.Username).Return(expectedUser, nil)

	user, err := suite.service.Login(suite.T().Context(), login)

	suite.Require().Nil(user)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, service.ErrInvalidCredentials)
	suite.Require().Equal(expectedErrMsg, err.Error())
}

// TestLogin_ShouldReturnErrorWhenStoredHashIsMalformed pins that a stored value that
// cannot be parsed is reported as a failure and never as a wrong password: the two mean
// different things, and only one of them is a failed authentication attempt.
func (suite *AuthenticationServiceTestSuite) TestLogin_ShouldReturnErrorWhenStoredHashIsMalformed() {
	tests := []struct {
		name       string
		stored     string
		wantErrIs  error
		wantErrMsg string
	}{
		{
			name:       "not a PHC string at all",
			stored:     "--fake-password--",
			wantErrIs:  service.ErrMalformedPasswordHash,
			wantErrMsg: "login service failed: decode password hash failed: malformed password hash",
		},
		{
			name:       "empty stored password",
			stored:     "",
			wantErrIs:  service.ErrMalformedPasswordHash,
			wantErrMsg: "login service failed: decode password hash failed: malformed password hash",
		},
		{
			name:       "missing the hash field",
			stored:     "$argon2id$v=19$m=65536,t=1,p=1$c29tZVNhbHRWYWx1ZTEyMw",
			wantErrIs:  service.ErrMalformedPasswordHash,
			wantErrMsg: "login service failed: decode password hash failed: malformed password hash",
		},
		{
			name:       "unparseable cost parameters",
			stored:     "$argon2id$v=19$m=lots,t=1,p=1$c29tZVNhbHRWYWx1ZTEyMw$c29tZUhhc2g",
			wantErrIs:  service.ErrMalformedPasswordHash,
			wantErrMsg: "login service failed: decode password hash failed: malformed password hash",
		},
		{
			name:       "zero cost parameters would panic the KDF",
			stored:     "$argon2id$v=19$m=0,t=0,p=0$c29tZVNhbHRWYWx1ZTEyMw$c29tZUhhc2g",
			wantErrIs:  service.ErrMalformedPasswordHash,
			wantErrMsg: "login service failed: decode password hash failed: malformed password hash",
		},
		{
			name:       "salt that is not base64",
			stored:     "$argon2id$v=19$m=65536,t=1,p=1$not base64!$c29tZUhhc2g",
			wantErrIs:  service.ErrMalformedPasswordHash,
			wantErrMsg: "login service failed: decode password hash failed: malformed password hash",
		},
		{
			name:       "another argon2 variant",
			stored:     "$argon2i$v=19$m=65536,t=1,p=1$c29tZVNhbHRWYWx1ZTEyMw$c29tZUhhc2g",
			wantErrIs:  service.ErrUnsupportedPasswordHash,
			wantErrMsg: "login service failed: decode password hash failed: unsupported password hash",
		},
		{
			name:       "another argon2 version",
			stored:     "$argon2id$v=16$m=65536,t=1,p=1$c29tZVNhbHRWYWx1ZTEyMw$c29tZUhhc2g",
			wantErrIs:  service.ErrUnsupportedPasswordHash,
			wantErrMsg: "login service failed: decode password hash failed: unsupported password hash",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			login := &domain.Login{Username: "someUsername", Password: "1235"}
			storedUser := &domain.User{
				Username: "someUsername",
				Password: tt.stored,
				Roles:    []commondomain.Role{commondomain.AdminRole},
				ID:       1,
			}

			suite.repository.EXPECT().FindUserByUsername(suite.T().Context(), login.Username).Return(storedUser, nil)

			user, err := suite.service.Login(suite.T().Context(), login)

			suite.Require().Nil(user)
			suite.Require().Error(err)
			suite.Require().ErrorIs(err, tt.wantErrIs)
			suite.Require().NotErrorIs(err, service.ErrInvalidCredentials)
			suite.Require().Equal(tt.wantErrMsg, err.Error())
		})
	}
}

func (suite *AuthenticationServiceTestSuite) TestLogin_ShouldReturnErrorWhenUserNotFound() {
	login := &domain.Login{
		Username:   "someUsername",
		Password:   "1235",
		RememberMe: true,
	}
	expectedErrMsg := "login service failed: invalid credentials"

	suite.repository.EXPECT().
		FindUserByUsername(suite.T().Context(), login.Username).Return(nil, database.ErrRecordNotFound)

	user, err := suite.service.Login(suite.T().Context(), login)

	suite.Require().Nil(user)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, service.ErrInvalidCredentials)
	// The unknown username must not leak through the error either: it answers the same
	// thing as a wrong password, so the two cannot be told apart by the response.
	suite.Require().Equal(expectedErrMsg, err.Error())
}

// TestLogin_UnknownUsernameIsNotAFastPath asserts the login does the KDF work even when
// the username does not exist. Returning early there answers in microseconds instead of
// tens of milliseconds, which is enough to enumerate accounts from the outside.
func (suite *AuthenticationServiceTestSuite) TestLogin_UnknownUsernameIsNotAFastPath() {
	knownUser := &domain.User{
		Username: "known",
		Password: phcHash("1234", "someSaltValue123", currentMemory, currentTime, currentParallelism, currentKeyLength),
		ID:       1,
	}

	suite.repository.EXPECT().FindUserByUsername(suite.T().Context(), "known").Return(knownUser, nil).Times(2)
	suite.repository.EXPECT().
		FindUserByUsername(suite.T().Context(), "unknown").Return(nil, database.ErrRecordNotFound)

	// One untimed run first, so the comparison is not paid for by a cold cache.
	_, _ = suite.service.Login(suite.T().Context(), &domain.Login{Username: "known", Password: "wrong"})

	knownElapsed := suite.timeLogin(&domain.Login{Username: "known", Password: "wrong"})
	unknownElapsed := suite.timeLogin(&domain.Login{Username: "unknown", Password: "wrong"})

	suite.Require().Positive(knownElapsed, "the known-user path must actually derive the password")
	suite.Require().Greater(
		unknownElapsed,
		knownElapsed/timingSlackFactor,
		"unknown username answered in %s against %s for a known one: the decoy verification was skipped",
		unknownElapsed,
		knownElapsed)
}

func (suite *AuthenticationServiceTestSuite) TestLogin_ShouldReturnErrorWhenDatabaseFailed() {
	login := &domain.Login{
		Username:   "someUsername",
		Password:   "1235",
		RememberMe: true,
	}
	findUserErr := errors.New("someErr")
	expectedErrMsg := "login service failed: someErr"

	suite.repository.EXPECT().FindUserByUsername(suite.T().Context(), login.Username).Return(nil, findUserErr)

	user, err := suite.service.Login(suite.T().Context(), login)

	suite.Require().Nil(user)
	suite.Require().Error(err)
	suite.Require().NotErrorIs(err, service.ErrInvalidCredentials)
	suite.Require().Equal(expectedErrMsg, err.Error())
}

func (suite *AuthenticationServiceTestSuite) timeLogin(login *domain.Login) time.Duration {
	suite.T().Helper()

	start := time.Now()
	user, err := suite.service.Login(suite.T().Context(), login)
	elapsed := time.Since(start)

	suite.Require().Nil(user)
	suite.Require().ErrorIs(err, service.ErrInvalidCredentials)

	return elapsed
}
