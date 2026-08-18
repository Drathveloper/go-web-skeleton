package domain

import (
	"errors"

	"github.com/Drathveloper/go-web-skeleton/common/domain"
)

type Login struct {
	Username   string
	Password   string
	RememberMe bool
}

type User struct {
	Username string
	Password string
	Roles    []domain.Role
	ID       uint
}

// ErrUserNotFound lives in the domain rather than in the service so the HTTP
// layer can distinguish "no such user" from "something broke" without
// importing the service package — the consumer-declares-the-interface rule
// means a handler never sees service types. Without this, a missing user is
// indistinguishable from a failure and answers 500 instead of 404.
var ErrUserNotFound = errors.New("user not found")
