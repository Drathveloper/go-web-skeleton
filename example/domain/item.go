package domain

import (
	"errors"
	"time"
)

type Item struct {
	ReleasedAt time.Time
	StartsAt   time.Time
	Category   *ItemCategory
	Name       string
	Notes      string
	Contact    string
	ID         uint
	Stock      uint
	Price      uint
	CategoryID uint
	Active     bool
}

// ErrItemNotFound lives in the domain, not in the service, so the HTTP layer
// can tell "no such record" from "something broke" and answer 404 rather than
// 500 — a handler never imports the service package.
var ErrItemNotFound = errors.New("item not found")
