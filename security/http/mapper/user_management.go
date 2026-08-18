package mapper

import (
	commondomain "github.com/Drathveloper/go-web-skeleton/common/domain"
	commondto "github.com/Drathveloper/go-web-skeleton/common/http/dto"
	commonmapper "github.com/Drathveloper/go-web-skeleton/common/http/mapper"
	"github.com/Drathveloper/go-web-skeleton/security/domain"
	"github.com/Drathveloper/go-web-skeleton/security/http/dto"
)

func DomainUsersToUsersViewResponse(
	users []domain.User, session *commondomain.Session) *commondto.ViewResponse[dto.Users] {
	result := make([]dto.User, 0, len(users))
	for _, user := range users {
		result = append(result, *DomainUserToDTOUser(&user))
	}
	data := &dto.Users{Users: result}
	return commonmapper.MapDataToViewResponse(data, getUserManagementBreadcrumb(), session)
}

func SessionDomainToCreateUserViewResponse(
	session *commondomain.Session) *commondto.ViewResponse[dto.CreateUserViewResponse] {
	data := &dto.CreateUserViewResponse{
		AllowedRoles: commondomain.GetAllowedRoles(),
	}
	return commonmapper.MapDataToViewResponse(data, getUserManagementBreadcrumb(), session)
}

func DomainUserToUpdateUserViewResponse(
	user *domain.User, session *commondomain.Session) *commondto.ViewResponse[dto.UpdateUserViewResponse] {
	data := &dto.UpdateUserViewResponse{
		User:         *DomainUserToDTOUser(user),
		AllowedRoles: commondomain.GetAllowedRoles(),
	}
	return commonmapper.MapDataToViewResponse(data, getUserManagementBreadcrumb(), session)
}

func DTOCreateUserProcessRequestToDomainUser(request *dto.CreateUserProcessRequest) *domain.User {
	return &domain.User{
		Username: request.Username,
		Password: request.Password,
		Roles:    dtoRolesToDomainRoles(request.Roles),
	}
}

func DTOUpdateUserProcessRequestToDomainUser(userID uint, request *dto.UpdateUserProcessRequest) *domain.User {
	return &domain.User{
		Username: request.Username,
		Password: request.Password,
		Roles:    dtoRolesToDomainRoles(request.Roles),
		ID:       userID,
	}
}

// DomainUserToDTOUser is exported because the handler needs it for the row
// fragment: the domain type is the only thing that crosses the service boundary,
// so the handler must not assemble the DTO by hand.
func DomainUserToDTOUser(user *domain.User) *dto.User {
	return &dto.User{
		ID:       user.ID,
		Username: user.Username,
		Roles:    domainRolesToDTORoles(user.Roles),
	}
}

func dtoRolesToDomainRoles(formRoles []string) []commondomain.Role {
	roles := make([]commondomain.Role, 0, len(formRoles))
	for _, formRole := range formRoles {
		roles = append(roles, commondomain.Role(formRole))
	}
	return roles
}

func domainRolesToDTORoles(roles []commondomain.Role) []string {
	result := make([]string, 0, len(roles))
	for _, role := range roles {
		result = append(result, string(role))
	}
	return result
}

func getUserManagementBreadcrumb() []string {
	return []string{"Security", "Users"}
}
