package mapper

import (
	"gorm.io/gorm"

	commondomain "github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/security/domain"
	"github.com/Drathveloper/go-web-skeleton/security/repository/rdbms/entity"
)

func EntityUsersToDomainUsers(users []entity.User) []domain.User {
	if len(users) == 0 {
		return []domain.User{}
	}
	domains := make([]domain.User, 0, len(users))
	for _, user := range users {
		domains = append(domains, *EntityUserToDomainUser(&user))
	}
	return domains
}

func EntityUserToDomainUser(user *entity.User) *domain.User {
	if user != nil {
		return &domain.User{
			ID:       user.ID,
			Username: user.Username,
			Password: user.Password,
			Roles:    entityRolesToDomainRoles(user.Roles),
		}
	}
	return &domain.User{}
}

func DomainUserToEntityUser(user *domain.User) *entity.User {
	return &entity.User{
		Model:    gorm.Model{ID: user.ID},
		Username: user.Username,
		Password: user.Password,
		Roles:    domainRolesToEntityRoles(user.Roles),
	}
}

func entityRolesToDomainRoles(roles []string) []commondomain.Role {
	domainRoles := make([]commondomain.Role, 0, len(roles))
	for _, role := range roles {
		domainRoles = append(domainRoles, commondomain.Role(role))
	}
	return domainRoles
}

func domainRolesToEntityRoles(roles []commondomain.Role) []string {
	entityRoles := make([]string, 0, len(roles))
	for _, role := range roles {
		entityRoles = append(entityRoles, string(role))
	}
	return entityRoles
}
