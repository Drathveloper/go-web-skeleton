package mapper_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/Drathveloper/go-web-skeleton/common/i18n"
)

// Loading the i18n bundle once, before any test runs, keeps the package-level
// localizer map out of reach of the parallel tests below: initializing it from
// several of them at once is a data race.
func TestMain(m *testing.M) {
	if err := i18n.InitializeI18n(); err != nil {
		fmt.Fprintf(os.Stderr, "initialize i18n failed: %v\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
