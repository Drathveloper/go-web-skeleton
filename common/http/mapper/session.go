package mapper

import (
	"github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/common/http/dto"
)

func MapDataToViewResponse[T any](data *T, breadcrumbs []string, session *domain.Session) *dto.ViewResponse[T] {
	msgs := mapDomainAlertMessagesToDTOAlertMessages(session.AlertMessages)
	return &dto.ViewResponse[T]{
		Data:        data,
		Language:    session.Language,
		CSRFToken:   session.CSRFToken,
		User:        session.Username,
		Msgs:        msgs,
		Breadcrumbs: breadcrumbs,
		IsLogged:    session.UserID != 0,
	}
}

func MapDomainErrorToViewResponse(code int, title string, err error) *dto.ViewResponse[any] {
	return &dto.ViewResponse[any]{
		Msgs: []dto.AlertMessage{dto.NewErrorMsg(code, title, err)},
	}
}

func mapDomainAlertMessagesToDTOAlertMessages(alerts domain.AlertMessages) []dto.AlertMessage {
	dtos := make([]dto.AlertMessage, 0, len(alerts))
	for _, alert := range alerts {
		dtos = append(dtos, dto.AlertMessage{
			Code:    alert.Code,
			Title:   alert.Title,
			Message: alert.Message,
			Type:    alert.Type,
		})
	}
	return dtos
}
