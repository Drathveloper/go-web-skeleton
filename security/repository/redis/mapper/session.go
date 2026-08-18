package mapper

import (
	"github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/security/repository/redis/entity"
)

// SessionEntityToSessionDomain reads back every field SessionDomainToSessionEntity
// writes. Language and IsLanguageOverridden used to be write-only, so the language
// chosen by the user was persisted and then silently dropped on the next request.
func SessionEntityToSessionDomain(id string, sessionEntity *entity.Session) *domain.Session {
	return &domain.Session{
		ID:                   id,
		UserID:               sessionEntity.UserID,
		Username:             sessionEntity.Username,
		Roles:                rolesToDomainRoles(sessionEntity.Roles),
		CSRFToken:            sessionEntity.CSRFToken,
		Language:             sessionEntity.Language,
		IsLanguageOverridden: sessionEntity.IsLanguageOverridden,
		AlertMessages:        alertMessagesEntityToAlertMessagesDomain(sessionEntity.AlertMessages),
	}
}

// SessionDomainToSessionEntity persists IsLanguageOverridden explicitly instead of
// deriving it from a non-empty Language: deciding whether the language is a user
// override belongs to the middleware, not to a mapper.
func SessionDomainToSessionEntity(session *domain.Session) *entity.Session {
	return &entity.Session{
		UserID:               session.UserID,
		Username:             session.Username,
		Roles:                sessionRoleListDomainToStringList(session.Roles),
		CSRFToken:            session.CSRFToken,
		Language:             session.Language,
		IsLanguageOverridden: session.IsLanguageOverridden,
		AlertMessages:        alertMessagesDomainToAlertMessagesEntity(session.AlertMessages),
	}
}

func sessionRoleListDomainToStringList(roles []domain.Role) []string {
	sessionRoles := make([]string, 0, len(roles))
	for _, role := range roles {
		sessionRoles = append(sessionRoles, string(role))
	}
	return sessionRoles
}

func rolesToDomainRoles(roles []string) []domain.Role {
	sessionRoles := make([]domain.Role, 0, len(roles))
	for _, role := range roles {
		sessionRoles = append(sessionRoles, domain.Role(role))
	}
	return sessionRoles
}

func alertMessagesDomainToAlertMessagesEntity(messages domain.AlertMessages) []entity.AlertMessages {
	entities := make([]entity.AlertMessages, 0, len(messages))
	for _, message := range messages {
		entities = append(entities, entity.AlertMessages{
			Code:    message.Code,
			Title:   message.Title,
			Message: message.Message,
			Type:    message.Type,
		})
	}
	return entities
}

func alertMessagesEntityToAlertMessagesDomain(messages []entity.AlertMessages) domain.AlertMessages {
	dtos := make(domain.AlertMessages, 0, len(messages))
	for _, message := range messages {
		dtos = append(dtos, domain.AlertMessage{
			Code:    message.Code,
			Title:   message.Title,
			Message: message.Message,
			Type:    message.Type,
		})
	}
	return dtos
}
