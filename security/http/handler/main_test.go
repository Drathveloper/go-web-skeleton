package handler_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Drathveloper/go-web-skeleton/common/http/validation"
	"github.com/Drathveloper/go-web-skeleton/common/i18n"
	"github.com/Drathveloper/go-web-skeleton/pkg/event"
)

// publisherSpy records every published event so the tests can assert exactly
// what a handler emitted. The handlers publish from the request goroutine, so
// no locking is needed as long as assertions run after ServeHTTP returns.
type publisherSpy struct {
	events []event.Event
}

func (s *publisherSpy) Publish(evt event.Event) {
	s.events = append(s.events, evt)
}

// The i18n bundle is a package-level map and the gin binding validator is a
// package-level engine. Loading both once here, before any test runs, is what
// lets the tests below use t.Parallel without racing on either.
func TestMain(m *testing.M) {
	if err := i18n.InitializeI18n(); err != nil {
		fmt.Fprintf(os.Stderr, "initialize i18n failed: %v\n", err)
		os.Exit(1)
	}
	validation.RegisterValidators()
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
