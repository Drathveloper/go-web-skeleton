package main

// `scaffold new --no-example` deletes the demonstration module and then strips,
// line by line, the lines the generator had injected into the six shared files.
// Matching by identifier is crude on purpose — the injections are line based,
// so their removal has to be too — and it has already gone wrong once: the
// marker {{/* scaffold:sidebar:items */}} contains the word "item", so it was
// removed along with the example, and the resulting project could never
// generate a module again. There is nothing to notice that: the project still
// builds, and the failure only shows up the next time somebody runs the
// generator.

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// sharedMarkerFiles are the files removeExampleReferences rewrites, relative to
// the repository root. The test reads the real ones: a marker that was renamed
// or a file that moved has to fail here, not in a project generated six months
// from now.
var sharedMarkerFiles = []string{
	filepath.Join("common", "http", "routes", "routes.go"),
	filepath.Join("common", "wire", "inject_repositories.go"),
	filepath.Join("common", "wire", "inject_services.go"),
	filepath.Join("common", "wire", "inject_http_handlers.go"),
	filepath.Join("common", "bootstrap", "db_migration.go"),
	filepath.Join("common", "http", "templates", "files", "layouts", "sidebar.gohtml"),
}

// scaffoldRepoRoot walks up from the test's directory to the module root.
func scaffoldRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		require.NotEqual(t, dir, parent, "no go.mod above %s", dir)
		dir = parent
	}
}

// TestIsExampleReferenceNeverStripsAMarker is the regression: run every line of
// the real shared files through the filter and check that no line carrying a
// scaffold: marker is dropped. It is driven by the files themselves so a new
// marker is covered the day it is added.
func TestIsExampleReferenceNeverStripsAMarker(t *testing.T) {
	root := scaffoldRepoRoot(t)
	total := 0

	for _, relative := range sharedMarkerFiles {
		t.Run(relative, func(t *testing.T) {
			content, err := os.ReadFile(filepath.Join(root, relative))
			require.NoError(t, err)

			markers := 0
			for line := range strings.SplitSeq(string(content), "\n") {
				if !strings.Contains(line, "scaffold:") {
					continue
				}
				markers++
				require.False(t, isExampleReference(line),
					"--no-example would delete the marker %q, and the generated project "+
						"could never inject into %s again", strings.TrimSpace(line), relative)
			}
			require.NotZero(t, markers, "%s carries no scaffold: marker, so this file proves nothing", relative)
			total += markers
		})
	}

	require.GreaterOrEqual(t, total, 14, "markers went missing from the shared files")
}

// TestIsExampleReferenceKeepsMarkersThatContainExampleWords is the same
// property written out by hand, because the regression was specifically a
// marker whose text contains an identifier from the list.
func TestIsExampleReferenceKeepsMarkersThatContainExampleWords(t *testing.T) {
	for _, line := range []string{
		`      {{/* scaffold:sidebar:items */}}`,
		`	// scaffold:migrations:entities`,
		`	// scaffold:repositories:init`,
		`// scaffold:items:whatever a future marker naming Item, /item and exampleservice`,
	} {
		require.False(t, isExampleReference(line), "marker line %q must survive", line)
	}
}

func TestIsExampleReferenceStripsTheGeneratedLines(t *testing.T) {
	lines := map[string]string{
		"route registration": `		registerItemRoutes(webRouter, container)`,
		"repository import":  `	examplerepository "github.com/acme/demo/example/repository/rdbms"`,
		"repository field":   `	ItemRepository *examplerepository.Item`,
		"repository init": `	container.ItemCategoryRepository = ` +
			`examplerepository.NewItemCategory(container.PostgresClient)`,
		"service field":    `	ItemCategoryService *exampleservice.ItemCategory`,
		"handler init":     `	container.ItemHandler = examplehandler.NewItem(container.ItemService)`,
		"migration entity": `		&exampleentity.ItemCategory{},`,
		"sidebar link":     `      <a href="/item" class="sidebar-link" title="{{ localize .Language "item.title" }}">`,
		"sidebar label": `        <span class="sidebar-label">` +
			`{{ localize .Language "item_category.title" }}</span>`,
		"sidebar section": `      <span class="sidebar-section">{{ localize .Language "example.title" }}</span>`,
		"kebab route":     `      <a href="/item-category" class="sidebar-link">`,
	}

	for name, line := range lines {
		t.Run(name, func(t *testing.T) {
			require.True(t, isExampleReference(line), "--no-example must strip %q", line)
		})
	}
}

func TestIsExampleReferenceKeepsEverythingElse(t *testing.T) {
	lines := map[string]string{
		"security route":   `		registerUserRoutes(webRouter, container)`,
		"security field":   `	UserRepository *securityrepository.User`,
		"security sidebar": `        <span class="sidebar-label">{{ localize .Language "sidebar.security.items.users" }}</span>`,
		"home link":        `      <a href="/" class="sidebar-link" title="{{ localize .Language "sidebar.home" }}">`,
		"struct close":     `}`,
		"blank":            ``,
		"package clause":   `package wire`,
	}

	for name, line := range lines {
		t.Run(name, func(t *testing.T) {
			require.False(t, isExampleReference(line), "--no-example must keep %q", line)
		})
	}
}

// TestStrippedSharedFilesStillParse applies the filter to the real Go files and
// parses the result. --no-example promises a project that still builds, and a
// filter that removed, say, a closing brace would take that promise with it.
func TestStrippedSharedFilesStillParse(t *testing.T) {
	root := scaffoldRepoRoot(t)

	for _, relative := range sharedMarkerFiles {
		if filepath.Ext(relative) != ".go" {
			continue
		}
		t.Run(relative, func(t *testing.T) {
			content, err := os.ReadFile(filepath.Join(root, relative))
			require.NoError(t, err)

			kept := make([]string, 0)
			removed := 0
			for line := range strings.SplitSeq(string(content), "\n") {
				if isExampleReference(line) {
					removed++
					continue
				}
				kept = append(kept, line)
			}
			require.NotZero(t, removed, "%s mentions no example module, so the strip proves nothing", relative)

			stripped := strings.Join(kept, "\n")
			_, parseErr := parser.ParseFile(token.NewFileSet(), relative, stripped, parser.SkipObjectResolution)
			require.NoError(t, parseErr, "stripping the example left %s unparseable:\n%s", relative, stripped)

			for _, identifier := range exampleIdentifiers {
				require.NotContains(t, stripped, identifier,
					"%s still mentions the example module after --no-example", relative)
			}
		})
	}
}
