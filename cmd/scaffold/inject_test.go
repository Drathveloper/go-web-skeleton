package main

// Injection is the only part of the generator that edits a file somebody else
// wrote. Everything it can get wrong is silent: the wrong indentation passes
// gofmt in a .go file and shows up crooked in a .gohtml, a lost idempotency
// duplicates a struct field the second time the generator runs, and a missing
// marker is a mistake in the project layout that has to be said out loud.

import (
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// injectionFixture writes content to a temp file and returns its path.
func injectionFixture(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, []byte(content), fileMode))
	return path
}

func readInjected(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(content)
}

const containerSource = `package wire

type Container struct {
	ItemCategoryRepository *string
	// scaffold:repositories:fields
}
`

func TestInjectionAppliesAboveTheMarkerWithItsIndentation(t *testing.T) {
	path := injectionFixture(t, "inject_repositories.go", containerSource)
	injection := Injection{
		Path:   path,
		Marker: "scaffold:repositories:fields",
		Lines:  []string{"InvoiceRepository *string"},
	}

	require.NoError(t, injection.Apply())

	lines := strings.Split(readInjected(t, path), "\n")
	injected, marker := -1, -1
	for index, line := range lines {
		if strings.Contains(line, "InvoiceRepository") {
			injected = index
		}
		if strings.Contains(line, "scaffold:repositories:fields") {
			marker = index
		}
	}

	require.NotEqual(t, -1, injected, "the line was not inserted")
	require.NotEqual(t, -1, marker, "the marker must survive: it is what makes the next run possible")
	require.Equal(t, marker-1, injected, "the line goes immediately above the marker")
	require.True(t, strings.HasPrefix(lines[injected], "\t"),
		"the inserted line must carry the marker's indentation, got %q", lines[injected])
}

// TestInjectionReformatsTheGoFile: inserting a struct field changes the
// alignment gofmt demands of every neighbouring field, so an injection that
// does not reformat leaves the project failing its own lint on the first run.
func TestInjectionReformatsTheGoFile(t *testing.T) {
	path := injectionFixture(t, "inject_repositories.go", containerSource)
	injection := Injection{
		Path:   path,
		Marker: "scaffold:repositories:fields",
		Lines:  []string{"InvoiceRepository *string"},
	}

	require.NoError(t, injection.Apply())

	content := readInjected(t, path)
	formatted, err := format.Source([]byte(content))
	require.NoError(t, err)
	require.Equal(t, string(formatted), content, "the injected file is not gofmt clean")

	// The pre-existing field was realigned around the new one, which is the
	// change a naive line insertion would not have made.
	require.Contains(t, content, "ItemCategoryRepository *string")
	require.Contains(t, content, "InvoiceRepository      *string")
}

func TestInjectionIsIdempotent(t *testing.T) {
	path := injectionFixture(t, "inject_repositories.go", containerSource)
	injection := Injection{
		Path:   path,
		Marker: "scaffold:repositories:fields",
		Lines:  []string{"InvoiceRepository *string"},
	}

	require.NoError(t, injection.Apply())
	afterFirst := readInjected(t, path)

	require.NoError(t, injection.Apply())
	afterSecond := readInjected(t, path)

	require.Equal(t, afterFirst, afterSecond, "a second run must not change the file")
	require.Equal(t, 1, strings.Count(afterSecond, "InvoiceRepository"),
		"the field was inserted twice, which does not compile")
}

// TestInjectionIsIdempotentWhenGofmtPadsTheInsertedLine is the same property
// for the case where the reformat rewrites the very line that was inserted: a
// short field name next to a long one comes back padded, so a dedup check that
// compares the raw text would no longer recognise it.
func TestInjectionIsIdempotentWhenGofmtPadsTheInsertedLine(t *testing.T) {
	path := injectionFixture(t, "inject_repositories.go", containerSource)
	injection := Injection{
		Path:   path,
		Marker: "scaffold:repositories:fields",
		Lines:  []string{"TagRepository *string"},
	}

	require.NoError(t, injection.Apply())
	require.NoError(t, injection.Apply())

	content := readInjected(t, path)
	require.Equal(t, 1, strings.Count(content, "TagRepository"),
		"the padded field was inserted twice, which does not compile:\n%s", content)
}

// TestInjectionKeepsIndentationInTemplates: the sidebar is a .gohtml, it is not
// gofmt'ed, and crooked markup is the only thing that would show it.
func TestInjectionKeepsIndentationInTemplates(t *testing.T) {
	const sidebar = `<nav>
      <a href="/item">item</a>
      {{/* scaffold:sidebar:items */}}
</nav>
`
	path := injectionFixture(t, "sidebar.gohtml", sidebar)
	injection := Injection{
		Path:   path,
		Marker: "scaffold:sidebar:items",
		Lines: []string{
			`<a href="/invoice" class="sidebar-link">`,
			`  <span>Invoice</span>`,
			`</a>`,
		},
	}

	require.NoError(t, injection.Apply())

	content := readInjected(t, path)
	require.Contains(t, content, "\n      <a href=\"/invoice\" class=\"sidebar-link\">\n")
	require.Contains(t, content, "\n        <span>Invoice</span>\n",
		"a nested line keeps its own relative indentation on top of the marker's")
	require.Contains(t, content, "{{/* scaffold:sidebar:items */}}")
}

func TestInjectionMissingMarker(t *testing.T) {
	path := injectionFixture(t, "routes.go", "package routes\n\nfunc Register() {}\n")
	injection := Injection{
		Path:   path,
		Marker: "scaffold:routes:register",
		Lines:  []string{"registerInvoiceRoutes(webRouter, container)"},
	}

	err := injection.Apply()

	require.ErrorIs(t, err, ErrMarkerNotFound)
	require.Contains(t, err.Error(), "scaffold:routes:register", "the error has to name the missing marker")
	require.Contains(t, err.Error(), path, "and the file it looked in")
	require.Equal(t, "package routes\n\nfunc Register() {}\n", readInjected(t, path),
		"a failed injection must not have written anything")
}

func TestInjectionMissingFile(t *testing.T) {
	injection := Injection{
		Path:   filepath.Join(t.TempDir(), "absent.go"),
		Marker: "scaffold:routes:register",
		Lines:  []string{"registerInvoiceRoutes(webRouter, container)"},
	}

	err := injection.Apply()

	require.Error(t, err)
	require.ErrorIs(t, err, os.ErrNotExist)
}

// TestInjectionRejectsCodeItCannotFormat: writing out Go that does not parse
// would leave the project broken with no clue where it happened.
func TestInjectionRejectsCodeItCannotFormat(t *testing.T) {
	path := injectionFixture(t, "inject_repositories.go", containerSource)
	injection := Injection{
		Path:   path,
		Marker: "scaffold:repositories:fields",
		Lines:  []string{"InvoiceRepository *string}}}"},
	}

	err := injection.Apply()

	require.Error(t, err)
	require.Contains(t, err.Error(), "gofmt")
	require.Equal(t, containerSource, readInjected(t, path), "the file must be left untouched")
}
