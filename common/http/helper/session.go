package helper

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
	"github.com/Drathveloper/go-web-skeleton/common/domain"
)

var ErrSessionNotFound = errors.New("session not found")

const (
	getSessionErrMsg = "get session failed"

	// OWASP recommends using 32 bytes of randomness for session IDs.
	sessionIDLengthBytes = 32
)

func GetSession(c *gin.Context) (*domain.Session, error) {
	value, ok := c.Get(constants.SessionGinContextKey)
	if !ok {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, getSessionErrMsg, ErrSessionNotFound)
	}
	session, ok := value.(*domain.Session)
	if !ok {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, getSessionErrMsg, ErrSessionNotFound)
	}
	return session, nil
}

func MustGetSession(c *gin.Context) *domain.Session {
	session, err := GetSession(c)
	if err != nil {
		panic(err)
	}
	return session
}

func GenerateSessionID() string {
	sessionID := make([]byte, sessionIDLengthBytes)
	if _, err := io.ReadFull(rand.Reader, sessionID); err != nil {
		panic("failed to generate session id")
	}
	return base64.RawURLEncoding.EncodeToString(sessionID)
}
