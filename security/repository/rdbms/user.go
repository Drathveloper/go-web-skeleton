package rdbms

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
	"github.com/Drathveloper/go-web-skeleton/common/database"
	"github.com/Drathveloper/go-web-skeleton/common/database/rdbms"
	commondomain "github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/security/domain"
	"github.com/Drathveloper/go-web-skeleton/security/repository/rdbms/entity"
	"github.com/Drathveloper/go-web-skeleton/security/repository/rdbms/mapper"
)

const (
	findAllUsersErrMsg       = "repository find all users failed"
	findUserByUsernameErrMsg = "repository find user by username failed"
	findUserByIDErrMsg       = "repository find user by id failed"
	createUserErrMsg         = "repository create user failed"
	updateUserErrMsg         = "repository update user failed"
)

type User struct {
	db rdbms.PostgresClient
}

func NewUser(db rdbms.PostgresClient) *User {
	return &User{
		db: db,
	}
}

func (u *User) FindAllUsers(ctx context.Context, pagination *commondomain.Pagination) ([]domain.User, error) {
	var userEntities []entity.User
	err := u.db.WithContext(ctx).Scopes(rdbms.Paginate(pagination.Page, pagination.Size)).Find(&userEntities).Error
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return make([]domain.User, 0), nil
		default:
			return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, findAllUsersErrMsg, err)
		}
	}
	return mapper.EntityUsersToDomainUsers(userEntities), nil
}

func (u *User) FindAllUserLookups(ctx context.Context) ([]domain.User, error) {
	var userEntities []entity.User

	err := u.db.WithContext(ctx).
		Select("id", "username").
		Find(&userEntities).Error
	if err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, findAllUsersErrMsg, err)
	}

	return mapper.EntityUsersToDomainUsers(userEntities), nil
}

func (u *User) FindUserByID(ctx context.Context, id uint) (*domain.User, error) {
	var userEntity entity.User
	if err := u.db.WithContext(ctx).First(&userEntity, id).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, findUserByIDErrMsg, database.ErrRecordNotFound)
		default:
			return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, findUserByIDErrMsg, err)
		}
	}
	return mapper.EntityUserToDomainUser(&userEntity), nil
}

func (u *User) CreateUser(ctx context.Context, user *domain.User) error {
	userEntity := mapper.DomainUserToEntityUser(user)
	if err := u.db.WithContext(ctx).Create(userEntity).Error; err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, createUserErrMsg, err)
	}
	return nil
}

func (u *User) UpdateUser(ctx context.Context, user *domain.User) error {
	userEntity := mapper.DomainUserToEntityUser(user)
	if err := u.db.WithContext(ctx).
		Where("id = ?", userEntity.ID).
		Updates(userEntity).Error; err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, updateUserErrMsg, err)
	}
	return nil
}

func (u *User) FindUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	var userEntity entity.User
	if err := u.db.WithContext(ctx).First(&userEntity, "username = ?", username).Error; err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, findUserByUsernameErrMsg, database.ErrRecordNotFound)
		default:
			return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, findUserByUsernameErrMsg, err)
		}
	}
	return mapper.EntityUserToDomainUser(&userEntity), nil
}
