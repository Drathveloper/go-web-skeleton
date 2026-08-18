package domain

import "errors"

type ItemCategory struct {
	Name string
	ID   uint
}

// ErrItemCategoryNotFound lives in the domain, not in the service, so the HTTP layer
// can tell "no such record" from "something broke" and answer 404 rather than
// 500 — a handler never imports the service package.
var ErrItemCategoryNotFound = errors.New("item category not found")
