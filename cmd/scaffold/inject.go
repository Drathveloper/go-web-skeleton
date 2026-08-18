package main

import (
	"errors"
	"fmt"
	"go/format"
	"os"
	"strings"
)

var ErrMarkerNotFound = errors.New("marker not found")

// Injection is one insertion into a shared file. The generator appends above a
// marker comment rather than manipulating the AST: it is far easier to read, it
// works on templates as well as Go, and it makes the seam visible to whoever
// opens the file next.
type Injection struct {
	Path   string
	Marker string
	Lines  []string
}

// Apply inserts the lines immediately above the marker, matching the marker's
// own indentation. It is idempotent: a line already present is not added again,
// so re-running the generator after an interrupted run is safe.
func (i Injection) Apply() error {
	content, err := os.ReadFile(i.Path)
	if err != nil {
		return fmt.Errorf("inject into %s failed: %w", i.Path, err)
	}

	lines := strings.Split(string(content), "\n")
	markerIndex := -1
	for index, line := range lines {
		if strings.Contains(line, i.Marker) {
			markerIndex = index
			break
		}
	}
	if markerIndex < 0 {
		return fmt.Errorf("inject into %s failed: %w: %s", i.Path, ErrMarkerNotFound, i.Marker)
	}

	markerLine := lines[markerIndex]
	indent := markerLine[:len(markerLine)-len(strings.TrimLeft(markerLine, " \t"))]

	existing := make(map[string]bool, len(lines))
	for _, line := range lines {
		existing[normalizeInjected(line)] = true
	}

	pending := make([]string, 0, len(i.Lines))
	for _, line := range i.Lines {
		if line == "" {
			pending = append(pending, "")
			continue
		}
		if existing[normalizeInjected(line)] {
			continue
		}
		pending = append(pending, indent+line)
	}
	if len(pending) == 0 {
		return nil
	}

	updated := make([]string, 0, len(lines)+len(pending))
	updated = append(updated, lines[:markerIndex]...)
	updated = append(updated, pending...)
	updated = append(updated, lines[markerIndex:]...)

	result := strings.Join(updated, "\n")

	// Inserting a struct field changes the alignment gofmt requires of every
	// other field, so an unformatted file is the normal outcome of a correct
	// injection, not an edge case. Reformat rather than leaving the project
	// failing its own lint on the first run.
	if strings.HasSuffix(i.Path, ".go") {
		formatted, formatErr := format.Source([]byte(result))
		if formatErr != nil {
			return fmt.Errorf("gofmt %s after injection failed: %w", i.Path, formatErr)
		}
		result = string(formatted)
	}

	if err = os.WriteFile(i.Path, []byte(result), fileMode); err != nil {
		return fmt.Errorf("inject into %s failed: %w", i.Path, err)
	}
	return nil
}

// normalizeInjected reduces a line to what makes it the same line for the
// purpose of not inserting it twice.
//
// Comparing the raw text does not work: the reformat step above realigns the
// struct fields around the one just inserted, so the inserted line comes back
// padded with a different number of spaces than the generator produced. The
// next run would not recognise it and would add it again, and a struct with
// the same field declared twice does not compile.
func normalizeInjected(line string) string {
	return strings.Join(strings.Fields(line), " ")
}
