package main

// A suppression is only correct for some shapes of a module. Emitting them
// unconditionally makes `make lint` red for small modules (nolintlint reports
// the unused directive); omitting them makes it red for large ones. So the
// generator measures its own output, and these tests are what says it measured
// it right — a wrong threshold here is a lint failure in generated code that
// the author did not write and cannot see the cause of.

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// bodyLines is n lines of a function body with no braces of their own, so the
// brace depth tracking in functionLength sees exactly one block.
func bodyLines(n int) string {
	lines := make([]string, 0, n)
	for range n {
		lines = append(lines, "\t_ = 1")
	}
	return strings.Join(lines, "\n")
}

// funlenSource is the shape the generator emits: the directive on its own line
// above the function, with its explanation.
func funlenSource(body int) string {
	return "package generated\n\n" +
		"//nolint:funlen // a declarative list with one literal per module field\n" +
		"func toView() string {\n" +
		bodyLines(body) + "\n" +
		"\treturn \"\"\n}\n"
}

func TestPruneUnusedNolintKeepsFunlenOnALongFunction(t *testing.T) {
	// The measured length is the body plus the signature and closing brace, so
	// this function is 61 lines: one over the threshold the linter uses.
	source := funlenSource(58)

	pruned := pruneUnusedNolint(source)

	require.Contains(t, pruned, "//nolint:funlen // a declarative list with one literal per module field")
	require.Equal(t, source, pruned, "nothing else may change")
}

func TestPruneUnusedNolintDropsFunlenOnAShortFunction(t *testing.T) {
	source := funlenSource(3)

	pruned := pruneUnusedNolint(source)

	require.NotContains(t, pruned, "nolint:funlen",
		"nolintlint fails the build for a directive that suppresses nothing")
	require.Contains(t, pruned, "func toView() string {", "only the directive line goes")
	require.Contains(t, pruned, "\treturn \"\"")
	require.Equal(t, strings.Count(source, "\n")-1, strings.Count(pruned, "\n"),
		"exactly one line was removed")
}

// TestPruneUnusedNolintFunlenBoundary pins the threshold itself: at exactly the
// limit the linter does not complain, so the directive would be unused.
func TestPruneUnusedNolintFunlenBoundary(t *testing.T) {
	atLimit := funlenSource(57)   // 60 lines measured
	overLimit := funlenSource(58) // 61 lines measured

	require.NotContains(t, pruneUnusedNolint(atLimit), "nolint:funlen")
	require.Contains(t, pruneUnusedNolint(overLimit), "nolint:funlen")
}

func TestPruneUnusedNolintVarnamelenOnItsOwnLine(t *testing.T) {
	source := func(body int) string {
		return "package generated\n\n" +
			"//nolint:varnamelen\n" +
			"func handle(c int) {\n" +
			bodyLines(body) + "\n}\n"
	}

	require.Contains(t, pruneUnusedNolint(source(10)), "nolint:varnamelen",
		"the short name spans more than five lines, so the directive is earned")
	require.NotContains(t, pruneUnusedNolint(source(1)), "nolint:varnamelen",
		"three lines: varnamelen does not complain, so the directive is unused")
}

// TestPruneUnusedNolintVarnamelenTrailing covers the shape every generated
// handler uses: the directive sits at the end of the line that opens the
// closure, so what gets measured is the closure, not the enclosing function.
func TestPruneUnusedNolintVarnamelenTrailing(t *testing.T) {
	source := func(body int) string {
		return "package generated\n\n" +
			"func Handle() gin.HandlerFunc {\n" +
			"\treturn func(c *gin.Context) { //nolint:varnamelen\n" +
			bodyLines(body) + "\n" +
			"\t}\n}\n"
	}

	long := pruneUnusedNolint(source(6))
	require.Contains(t, long, "return func(c *gin.Context) { //nolint:varnamelen")

	short := pruneUnusedNolint(source(1))
	require.NotContains(t, short, "nolint:varnamelen")
	require.Contains(t, short, "\treturn func(c *gin.Context) {\n",
		"the code has to survive with only the directive stripped, and no trailing space")
	require.Contains(t, short, "\t_ = 1", "the closure body is untouched")
}

func TestPruneUnusedNolintLeavesOtherDirectivesAlone(t *testing.T) {
	source := "package generated\n\n" +
		"//nolint:gochecknoglobals // the entity list is a registry\n" +
		"var entities = []any{}\n\n" +
		"func short() {\n\t_ = 1\n}\n"

	require.Equal(t, source, pruneUnusedNolint(source))
}

func TestFunctionLengthCountsToTheMatchingBrace(t *testing.T) {
	lines := strings.Split("//nolint:funlen\nfunc f() {\n\tif x {\n\t\t_ = 1\n\t}\n}\nfunc g() {\n}\n", "\n")

	// func f spans five lines, signature through closing brace: the nested if
	// block must not end the count early.
	require.Equal(t, 5, functionLength(lines, 0))
}
