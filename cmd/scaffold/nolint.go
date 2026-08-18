package main

import (
	"strings"
)

// funlenThreshold and varnamelenSpan mirror the linter defaults the project
// runs with. They are only used to decide whether a suppression is warranted.
const (
	funlenThreshold = 60
	varnamelenSpan  = 5
)

// pruneUnusedNolint removes //nolint directives the generated code does not
// actually need.
//
// A suppression is only correct for some shapes of a module: a form with nine
// fields trips funlen, one with a single field does not, and nolintlint then
// fails the build for the *unused* directive. Emitting them unconditionally
// would make `make lint` red for small modules; omitting them would make it red
// for large ones. So the generator measures its own output and keeps only the
// directives that are earned.
func pruneUnusedNolint(source string) string {
	lines := strings.Split(source, "\n")
	kept := make([]string, 0, len(lines))

	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "//nolint:funlen"):
			if functionLength(lines, index) > funlenThreshold {
				kept = append(kept, line)
			}
		case strings.HasPrefix(trimmed, "//nolint:varnamelen"):
			if functionLength(lines, index) > varnamelenSpan {
				kept = append(kept, line)
			}
		case strings.HasSuffix(trimmed, "//nolint:varnamelen"):
			if functionLength(lines, index) > varnamelenSpan {
				kept = append(kept, line)
			} else {
				kept = append(kept, strings.TrimRight(
					strings.TrimSuffix(line, "//nolint:varnamelen"), " \t"))
			}
		default:
			kept = append(kept, line)
		}
	}
	return strings.Join(kept, "\n")
}

// functionLength counts the lines of the function body that follows or contains
// the directive at index, by tracking brace depth from the first opening brace.
func functionLength(lines []string, index int) int {
	start := index
	for start < len(lines) && !strings.Contains(lines[start], "{") {
		start++
	}
	if start >= len(lines) {
		return 0
	}

	depth := 0
	for offset := start; offset < len(lines); offset++ {
		depth += strings.Count(lines[offset], "{") - strings.Count(lines[offset], "}")
		if depth <= 0 {
			return offset - start + 1
		}
	}
	return len(lines) - start
}
