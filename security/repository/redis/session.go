package redis

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
	"github.com/Drathveloper/go-web-skeleton/common/database"
	"github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/security/repository/redis/entity"
	"github.com/Drathveloper/go-web-skeleton/security/repository/redis/mapper"
)

const (
	getSessionErrMsg    = "repository get session failed"
	saveSessionErrMsg   = "repository save session failed"
	deleteSessionErrMsg = "repository delete session failed"
	sessionKeyPrefix    = "session:"
)

type SessionConfigProvider interface {
	GetSessionTTL() time.Duration
}

type Session struct {
	client     redis.UniversalClient
	sessionTTL time.Duration
}

func NewSession(client redis.UniversalClient, config SessionConfigProvider) *Session {
	return &Session{
		client:     client,
		sessionTTL: config.GetSessionTTL(),
	}
}

// sessionKey derives the Redis key of a session from its ID. Every read, write and
// delete path goes through here so the three can never diverge again: the session
// ID is a bearer credential, so anyone able to list the keyspace must not be able
// to read it back. Note sha512.Sum512, not hash.Hash.Sum — the latter *appends*
// the current digest to its argument, so hashing this way with a fresh hash.Hash
// returns the plaintext key followed by the digest of the empty string.
func sessionKey(id string) string {
	sum := sha512.Sum512([]byte(sessionKeyPrefix + id))
	return base64.StdEncoding.EncodeToString(sum[:])
}

func (s *Session) Get(ctx context.Context, key string) (*domain.Session, error) {
	cmd := s.client.Get(ctx, sessionKey(key))
	if cmd.Err() != nil {
		switch {
		case errors.Is(cmd.Err(), redis.Nil):
			return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, getSessionErrMsg, database.ErrRecordNotFound)
		default:
			return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, getSessionErrMsg, cmd.Err())
		}
	}
	var sessionEntity entity.Session
	if err := json.Unmarshal([]byte(cmd.Val()), &sessionEntity); err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, getSessionErrMsg, err)
	}
	return mapper.SessionEntityToSessionDomain(key, &sessionEntity), nil
}

func (s *Session) Save(ctx context.Context, session *domain.Session) error {
	sessionEntity := mapper.SessionDomainToSessionEntity(session)
	value, err := json.Marshal(sessionEntity)
	if err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, saveSessionErrMsg, err)
	}
	if cmd := s.client.Set(ctx, sessionKey(session.ID), value, s.sessionTTL); cmd.Err() != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, saveSessionErrMsg, cmd.Err())
	}
	return nil
}

// Delete removes a session by ID. It is what the login flow uses to drop the
// anonymous session it is promoting, which otherwise lingers for the whole TTL.
func (s *Session) Delete(ctx context.Context, key string) error {
	if cmd := s.client.Del(ctx, sessionKey(key)); cmd.Err() != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, deleteSessionErrMsg, cmd.Err())
	}
	return nil
}
