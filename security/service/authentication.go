package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
	"github.com/Drathveloper/go-web-skeleton/common/database"
	"github.com/Drathveloper/go-web-skeleton/security/domain"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

const loginErrMsg = "login service failed"

type AuthenticationRepository interface {
	FindUserByUsername(ctx context.Context, username string) (*domain.User, error)
}

type Authentication struct {
	repository AuthenticationRepository
}

func NewAuthentication(repository AuthenticationRepository) *Authentication {
	return &Authentication{
		repository: repository,
	}
}

func (s *Authentication) Login(ctx context.Context, login *domain.Login) (*domain.User, error) {
	user, err := s.repository.FindUserByUsername(ctx, login.Username)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			// Returning here directly would answer an unknown username far
			// faster than a known one with a wrong password, which is enough
			// to enumerate accounts. Do the KDF work anyway, against a hash
			// generated at startup, so both paths cost the same.
			_, _ = verifyPassword(login.Password, enumerationDecoyHash)
			return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, loginErrMsg, ErrInvalidCredentials)
		default:
			return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, loginErrMsg, err)
		}
	}
	isPasswordValid, err := verifyPassword(login.Password, user.Password)
	if err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, loginErrMsg, err)
	}
	if !isPasswordValid {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, loginErrMsg, ErrInvalidCredentials)
	}
	return user, nil
}
