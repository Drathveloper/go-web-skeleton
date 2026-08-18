package mapper

import (
	"github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/common/http/dto"
	"github.com/Drathveloper/go-web-skeleton/common/i18n"
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

// MapDomainErrorToViewResponse builds the payload for the error page.
//
// It takes i18n keys, not text, and resolves them here: that way the error
// page is localized like every other screen, and there is no signature a
// caller could use to push a raw error onto it. Log the underlying error at
// the call site; only these two keys reach the user.
func MapDomainErrorToViewResponse(lang string, code int, titleKey, messageKey string) *dto.ViewResponse[any] {
	if lang == "" {
		lang = i18n.DefaultLanguage
	}
	return &dto.ViewResponse[any]{
		Language: lang,
		Msgs: []dto.AlertMessage{
			dto.NewErrorMsg(code, i18n.LocalizeMessage(lang, titleKey), i18n.LocalizeMessage(lang, messageKey)),
		},
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
