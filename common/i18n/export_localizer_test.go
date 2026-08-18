package i18n //nolint:testpackage

import (
	"maps"
	"testing"

	i18nlib "github.com/nicksnyder/go-i18n/v2/i18n"
)

// SetLocalizers swaps the package-level catalog for the duration of one test and
// puts the previous one back when it ends.
//
// The catalog is a global. The source's version of this hook was a bare setter
// called from subtests that had all declared t.Parallel(), so several of them
// wrote and read that map at once: a data race that happened to pass because
// every case assigned before it read. Taking *testing.T and restoring through
// Cleanup keeps the mutation scoped to one test, and the tests that use it run
// sequentially — see the comment at the top of localizer_test.go.
func SetLocalizers(t *testing.T, localizersByLocale map[string]*i18nlib.Localizer) {
	t.Helper()

	previous := localizers
	t.Cleanup(func() {
		localizers = previous
	})
	localizers = maps.Clone(localizersByLocale)
}
