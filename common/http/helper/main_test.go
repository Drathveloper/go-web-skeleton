package helper_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/Drathveloper/go-web-skeleton/common/i18n"
)

// The i18n bundle is a package-level map. Loading it once here, before any test
// runs, is what lets the tests below use t.Parallel: calling InitializeI18n from
// several parallel tests writes that map concurrently, which is a data race the
// -race build reports and a plain run only hides.
func TestMain(m *testing.M) {
	if err := i18n.InitializeI18n(); err != nil {
		fmt.Fprintf(os.Stderr, "initialize i18n failed: %v\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
