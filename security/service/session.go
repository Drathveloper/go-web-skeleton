package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
	"github.com/Drathveloper/go-web-skeleton/common/database"
	commondomain "github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/security/domain"
)

const (
	loadSessionErrMsg            = "load session service failed"
	createAnonymousSessionErrMsg = "create anonymous session service failed"
	createUserSessionErrMsg      = "create user session service failed"
	updateSessionErrMsg          = "update session service failed"
	destroySessionErrMsg         = "destroy session service failed"
)

type SessionRepository interface {
	Get(ctx context.Context, key string) (*commondomain.Session, error)
	Save(ctx context.Context, session *commondomain.Session) error
	Delete(ctx context.Context, key string) error
}

type Session struct {
	repository SessionRepository
}

func NewSession(repository SessionRepository) *Session {
	return &Session{
		repository: repository,
	}
}

func (s *Session) LoadSession(ctx context.Context, sessionID string) (*commondomain.Session, error) {
	sessionInformation, err := s.repository.Get(ctx, sessionID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			sessionInformation, err = s.CreateAnonymousSession(ctx, sessionID)
			if err != nil {
				return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, loadSessionErrMsg, err)
			}
		default:
			return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, loadSessionErrMsg, err)
		}
	}
	return sessionInformation, nil
}

func (s *Session) CreateAnonymousSession(ctx context.Context, sessionID string) (*commondomain.Session, error) {
	session := &commondomain.Session{
		ID:            sessionID,
		Roles:         make([]commondomain.Role, 0),
		AlertMessages: make(commondomain.AlertMessages, 0),
	}
	if err := s.repository.Save(ctx, session); err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, createAnonymousSessionErrMsg, err)
	}
	return session, nil
}

// CreateUserSession promotes an anonymous visitor into an authenticated one. The caller
// supplies a freshly generated userSessionID — reusing the anonymous identifier across
// the authentication boundary is session fixation — and the anonymous entry is dropped
// here, so it does not survive in the store for the whole session TTL. The delete comes
// first on purpose: if the save fails afterwards the visitor simply gets a new anonymous
// session on the next request, whereas the opposite order would leave an authenticated
// session behind that nobody holds a cookie for.
func (s *Session) CreateUserSession(
	ctx context.Context,
	anonymousSessionID, userSessionID string,
	user *domain.User) (*commondomain.Session, error) {
	if anonymousSessionID != "" && anonymousSessionID != userSessionID {
		if err := s.repository.Delete(ctx, anonymousSessionID); err != nil {
			return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, createUserSessionErrMsg, err)
		}
	}
	session := &commondomain.Session{
		ID:            userSessionID,
		UserID:        user.ID,
		Username:      user.Username,
		Roles:         user.Roles,
		AlertMessages: make(commondomain.AlertMessages, 0),
	}
	if err := s.repository.Save(ctx, session); err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, createUserSessionErrMsg, err)
	}
	return session, nil
}

func (s *Session) UpdateSession(ctx context.Context, session *commondomain.Session) error {
	if err := s.repository.Save(ctx, session); err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, updateSessionErrMsg, err)
	}
	return nil
}

func (s *Session) DestroySession(ctx context.Context, session *commondomain.Session) error {
	if err := s.repository.Delete(ctx, session.ID); err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, destroySessionErrMsg, err)
	}
	return nil
}
