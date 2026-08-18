package mapper

import (
	"net/url"

	commondomain "github.com/Drathveloper/go-web-skeleton/common/domain"
	commondto "github.com/Drathveloper/go-web-skeleton/common/http/dto"
	commonmapper "github.com/Drathveloper/go-web-skeleton/common/http/mapper"
	"github.com/Drathveloper/go-web-skeleton/security/domain"
)

func DomainSessionToLoginViewResponse(session *commondomain.Session) *commondto.ViewResponse[any] {
	return commonmapper.MapDataToViewResponse[any](nil, getLoginBreadcrumb(), session)
}

func FormLoginToDomainLogin(values url.Values) (*domain.Login, error) {
	rememberMe := values.Get("remember_me") != ""
	return &domain.Login{
		Username:   values.Get("username"),
		Password:   values.Get("password"),
		RememberMe: rememberMe,
	}, nil
}

func getLoginBreadcrumb() []string {
	return []string{"Security", "Login"}
}
