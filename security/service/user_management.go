package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
	"github.com/Drathveloper/go-web-skeleton/common/database"
	commondomain "github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/security/domain"
)

const (
	listUsersErrMsg            = "list users service failed"
	getUserByIDErrMsg          = "get user by id service failed"
	createUserErrMsg           = "create user service failed"
	updateUserErrMsg           = "update user service failed"
	generateSaltErrMsg         = "generate salt failed"
	generatePasswordHashErrMsg = "generate password hash failed"
)

type UserManagementRepository interface {
	FindAllUsers(ctx context.Context, pagination *commondomain.Pagination) ([]domain.User, error)
	FindAllUserLookups(ctx context.Context) ([]domain.User, error)
	FindUserByID(ctx context.Context, id uint) (*domain.User, error)
	CreateUser(ctx context.Context, user *domain.User) error
	UpdateUser(ctx context.Context, user *domain.User) error
}

type UserManagement struct {
	repository UserManagementRepository
}

func NewUserManagement(repository UserManagementRepository) *UserManagement {
	return &UserManagement{
		repository: repository,
	}
}

func (s *UserManagement) ListUsers(ctx context.Context, pagination *commondomain.Pagination) ([]domain.User, error) {
	users, err := s.repository.FindAllUsers(ctx, pagination)
	if err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, listUsersErrMsg, err)
	}
	return users, nil
}

func (s *UserManagement) ListUserLookups(ctx context.Context) ([]domain.User, error) {
	users, err := s.repository.FindAllUserLookups(ctx)
	if err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, listUsersErrMsg, err)
	}
	return users, nil
}

func (s *UserManagement) CreateUser(ctx context.Context, user *domain.User) error {
	passwordHash, err := s.generatePasswordHash(user.Password)
	if err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, createUserErrMsg, err)
	}
	user.Password = passwordHash
	if err = s.repository.CreateUser(ctx, user); err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, createUserErrMsg, err)
	}
	return nil
}

func (s *UserManagement) UpdateUser(ctx context.Context, user *domain.User) error {
	if user.Password != "" {
		passwordHash, err := s.generatePasswordHash(user.Password)
		if err != nil {
			return fmt.Errorf(constants.DefaultWrappedErrorTemplate, updateUserErrMsg, err)
		}
		user.Password = passwordHash
	}
	if err := s.repository.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, updateUserErrMsg, err)
	}
	return nil
}

func (s *UserManagement) GetUserByID(ctx context.Context, userID uint) (*domain.User, error) {
	user, err := s.repository.FindUserByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, getUserByIDErrMsg, domain.ErrUserNotFound)
		default:
			return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, getUserByIDErrMsg, err)
		}
	}
	return user, nil
}

// generatePasswordHash returns the PHC encoding of the derived password:
// `$argon2id$v=19$m=<memory>,t=<time>,p=<parallelism>$<salt>$<hash>`. The cost
// parameters travel with the hash so they can be raised without invalidating the
// passwords already stored.
func (s *UserManagement) generatePasswordHash(plaintextPassword string) (string, error) {
	salt, err := s.generateRandomSalt()
	if err != nil {
		return "", fmt.Errorf(constants.DefaultWrappedErrorTemplate, generatePasswordHashErrMsg, err)
	}
	return hashPassword(plaintextPassword, salt), nil
}

func (s *UserManagement) generateRandomSalt() ([]byte, error) {
	salt := make([]byte, argon2HashSaltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, generateSaltErrMsg, err)
	}
	return salt, nil
}
