package templates //nolint:testpackage // the scan needs the unexported templatesFS

// Nothing in the toolchain notices a missing translation. go-i18n answers an
// unknown key with the key itself, so the page still renders and the bug ships
// as a literal "auth.login.title" on screen. These tests are what stands between
// a typo in a .gohtml and that page.
//
// The scan is static: it parses every embedded template and walks the parse
// tree looking for calls to the `localize` template function. That is stronger
// than a grep, because it also resolves the keys that are *composed*:
// components/auth/user_form.gohtml builds its keys with
// `printf "%s.form.username.label" $prefix`, where $prefix is one of two string
// literals assigned in the same template. A grep sees none of those fourteen
// keys. Any localize call this scanner cannot resolve fails
// TestEveryLocalizeCallIsStaticallyResolvable rather than being skipped in
// silence, so the coverage claim below stays true or the suite goes red.

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"maps"
	"path"
	"slices"
	"sort"
	"strings"
	"testing"
	"text/template"
	"text/template/parse"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/Drathveloper/go-web-skeleton/common/i18n"
)

// localizeFuncName and printfFuncName are the two template functions the
// scanner understands. localize takes (language, key).
const (
	localizeFuncName = "localize"
	printfFuncName   = "printf"

	// localizeArgCount is `localize` plus its two arguments.
	localizeArgCount = 3

	// catalogGlob matches every per-module catalog. Catalogs are named
	// <module>.<lang>.json, so the language cannot be assumed from a fixed
	// file list: every generated module drops another pair in here.
	catalogGlob = "files/*.json"

	// catalogNameParts is <module>.<lang>.json split on the dot.
	catalogNameParts = 3
)

// keyUse is one i18n key a template asks for, and where it asked for it.
type keyUse struct {
	key      string
	source   string // files/pages/auth/login.gohtml:12
	composed bool   // built with printf rather than written as a literal
}

// unresolvedUse is a localize call whose key the scanner could not work out,
// which means that key is not covered by the checks below.
type unresolvedUse struct {
	source string
	expr   string
}

// catalog is one <module>.<lang>.json, flattened. The nesting in the file is
// only a way of writing prefixes down once: {"item":{"fields":{"name":...}}}
// provides the key item.fields.name, which is what a template asks for.
type catalog struct {
	keys map[string]string
	file string
	lang string
}

func TestTemplateI18nKeysExistInEveryLocale(t *testing.T) {
	catalogs := loadCatalogs(t)
	byLang := keysByLanguage(catalogs)
	require.GreaterOrEqual(t, len(byLang), 2,
		"a key coverage test with fewer than two locales cannot fail: found %v", slices.Sorted(maps.Keys(byLang)))

	uses, _ := collectTemplateKeys(t)
	languages := slices.Sorted(maps.Keys(byLang))

	problems := make([]string, 0)
	for _, use := range uses {
		for _, lang := range languages {
			if _, ok := byLang[lang][use.key]; !ok {
				problems = append(problems,
					fmt.Sprintf("  %-8s is missing %-45q used by %s", lang, use.key, use.source))
			}
		}
	}
	problems = sortedUnique(problems)

	if len(problems) > 0 {
		t.Fatalf("%d i18n key(s) used by a template are absent from a catalog:\n%s\n"+
			"add them to the matching common/i18n/files/<module>.<lang>.json",
			len(problems), strings.Join(problems, "\n"))
	}
}

// TestEveryLocalizeCallIsStaticallyResolvable is what keeps the test above
// honest. A key the scanner cannot resolve is a key nothing checks, and the
// failure mode of a static scan is to quietly see less than it claims to. If a
// template starts composing a key the scanner does not understand, this fails
// and says so instead of shrinking the coverage in silence.
func TestEveryLocalizeCallIsStaticallyResolvable(t *testing.T) {
	_, unresolved := collectTemplateKeys(t)

	lines := make([]string, 0, len(unresolved))
	for _, use := range unresolved {
		lines = append(lines, fmt.Sprintf("  %s: %s", use.source, use.expr))
	}
	if len(lines) > 0 {
		t.Fatalf("%d localize call(s) build their key in a way the key coverage scan cannot follow, "+
			"so those keys are checked by nothing:\n%s\n"+
			"either write the key as a literal or teach resolveKeyArg how to read it",
			len(lines), strings.Join(sortedUnique(lines), "\n"))
	}
}

// TestTemplateKeyScanResolvesComposedKeys pins the printf resolution itself.
// Without it, a regression in the scanner would show up as a suite that still
// passes while covering fourteen fewer keys.
func TestTemplateKeyScanResolvesComposedKeys(t *testing.T) {
	uses, _ := collectTemplateKeys(t)

	composed := make(map[string]string)
	for _, use := range uses {
		if use.composed {
			composed[use.key] = use.source
		}
	}

	// Both branches of the create/update prefix in components/auth/user_form.gohtml.
	for _, key := range []string{
		"auth.create_user.title",
		"auth.create_user.form.username.label",
		"auth.create_user.form.roles.label",
		"auth.update_user.title",
		"auth.update_user.form.confirm_password.placeholder",
		"auth.update_user.form.submit.value",
	} {
		require.Contains(t, composed, key,
			"the scanner no longer resolves printf-composed keys; it now sees only %v", slices.Sorted(maps.Keys(composed)))
	}
}

// TestLocaleCatalogsAreSymmetric is the reverse direction: a key that exists in
// one language and not in another. The template scan cannot see it, because the
// key it looks up resolves in the language the test happened to render.
func TestLocaleCatalogsAreSymmetric(t *testing.T) {
	catalogs := loadCatalogs(t)
	byLang := keysByLanguage(catalogs)
	languages := slices.Sorted(maps.Keys(byLang))
	require.GreaterOrEqual(t, len(languages), 2, "found only %v", languages)

	every := make(map[string][]string)
	for lang, keys := range byLang {
		for key := range keys {
			every[key] = append(every[key], lang)
		}
	}

	problems := make([]string, 0)
	for key, present := range every {
		if len(present) == len(languages) {
			continue
		}
		missing := make([]string, 0, len(languages))
		for _, lang := range languages {
			if !slices.Contains(present, lang) {
				missing = append(missing, lang)
			}
		}
		problems = append(problems, fmt.Sprintf("  %-45q defined in %v, missing in %v",
			key, sortedUnique(present), missing))
	}

	if len(problems) > 0 {
		t.Fatalf("%d i18n key(s) are not defined in every locale:\n%s",
			len(problems), strings.Join(sortedUnique(problems), "\n"))
	}
}

// TestLocaleCatalogsDefineEachKeyOnce guards a hazard that grows with every
// generated module: go-i18n merges all catalogs of a language into one bundle,
// so two files defining the same key do not conflict — one silently wins, and
// which one depends on walk order.
func TestLocaleCatalogsDefineEachKeyOnce(t *testing.T) {
	catalogs := loadCatalogs(t)

	owners := make(map[string]map[string][]string) // lang -> key -> files
	for _, cat := range catalogs {
		if owners[cat.lang] == nil {
			owners[cat.lang] = make(map[string][]string)
		}
		for key := range cat.keys {
			owners[cat.lang][key] = append(owners[cat.lang][key], cat.file)
		}
	}

	problems := make([]string, 0)
	for _, lang := range slices.Sorted(maps.Keys(owners)) {
		for key, files := range owners[lang] {
			if len(files) > 1 {
				problems = append(problems,
					fmt.Sprintf("  %-8s %-45q defined in %v", lang, key, sortedUnique(files)))
			}
		}
	}

	if len(problems) > 0 {
		t.Fatalf("%d i18n key(s) are defined by more than one catalog of the same language:\n%s\n"+
			"go-i18n merges them, so one definition wins at random",
			len(problems), strings.Join(sortedUnique(problems), "\n"))
	}
}

// loadCatalogs reads every catalog out of the embedded locale filesystem and
// flattens it. The language comes from the file name, exactly as go-i18n takes
// it, so the two can never disagree.
func loadCatalogs(t *testing.T) []catalog {
	t.Helper()

	entries, err := fs.Glob(i18n.LocaleFS, catalogGlob)
	require.NoError(t, err)
	require.NotEmpty(t, entries, "no catalogs matched %s: the checks below would all pass vacuously", catalogGlob)

	catalogs := make([]catalog, 0, len(entries))
	for _, entry := range entries {
		name := path.Base(entry)
		parts := strings.Split(name, ".")
		require.Len(t, parts, catalogNameParts, "catalog %s is not named <module>.<lang>.json, "+
			"so go-i18n cannot derive its language", name)

		raw, readErr := i18n.LocaleFS.ReadFile(entry)
		require.NoError(t, readErr)

		var tree map[string]any
		require.NoError(t, json.Unmarshal(raw, &tree), "catalog %s is not valid JSON", name)

		keys := make(map[string]string)
		flattenCatalog("", tree, keys)
		require.NotEmpty(t, keys, "catalog %s defines no keys", name)

		catalogs = append(catalogs, catalog{keys: keys, file: name, lang: parts[1]})
	}
	return catalogs
}

// keysByLanguage merges the per-module catalogs the way go-i18n does: one key
// space per language.
func keysByLanguage(catalogs []catalog) map[string]map[string]string {
	byLang := make(map[string]map[string]string)
	for _, cat := range catalogs {
		if byLang[cat.lang] == nil {
			byLang[cat.lang] = make(map[string]string)
		}
		maps.Copy(byLang[cat.lang], cat.keys)
	}
	return byLang
}

// flattenCatalog turns the nested JSON into the dotted keys templates use.
func flattenCatalog(prefix string, node map[string]any, into map[string]string) {
	for name, value := range node {
		key := name
		if prefix != "" {
			key = prefix + "." + name
		}
		if child, isObject := value.(map[string]any); isObject && !isPluralMessage(child) {
			flattenCatalog(key, child, into)
			continue
		}
		into[key] = fmt.Sprint(value)
	}
}

// isPluralMessage tells a namespace apart from a go-i18n message written in
// long form. {"one": "...", "other": "..."} is one message with plural forms,
// not two keys, and flattening it would invent key.one and key.other.
func isPluralMessage(node map[string]any) bool {
	messageFields := map[string]bool{
		"zero": true, "one": true, "two": true, "few": true, "many": true, "other": true,
		"description": true, "hash": true, "id": true, "translation": true,
		"leftdelim": true, "rightdelim": true,
	}
	if len(node) == 0 {
		return false
	}
	for name := range node {
		if !messageFields[strings.ToLower(name)] {
			return false
		}
	}
	return true
}

// collectTemplateKeys parses every embedded template and returns the keys its
// localize calls ask for, plus the calls it could not resolve.
func collectTemplateKeys(t *testing.T) ([]keyUse, []unresolvedUse) {
	t.Helper()

	funcs := templateFuncs()
	uses := make([]keyUse, 0)
	unresolved := make([]unresolvedUse, 0)

	err := fs.WalkDir(templatesFS, "files", func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || path.Ext(filePath) != templateExtension {
			return nil
		}
		source, readErr := templatesFS.ReadFile(filePath)
		if readErr != nil {
			return fmt.Errorf("read template %s failed: %w", filePath, readErr)
		}
		parsed, parseErr := template.New(path.Base(filePath)).Funcs(funcs).Parse(string(source))
		if parseErr != nil {
			return fmt.Errorf("parse template %s failed: %w", filePath, parseErr)
		}
		fileUses, fileUnresolved := scanTemplate(filePath, string(source), parsed)
		uses = append(uses, fileUses...)
		unresolved = append(unresolved, fileUnresolved...)
		return nil
	})
	require.NoError(t, err)
	require.NotEmpty(t, uses, "no localize call was found in any template: the scanner is broken, "+
		"not the templates")

	return uses, unresolved
}

// templateFuncs is the function map the real renderer installs. Taking it from
// registerFuncs rather than listing the names here means a new template
// function cannot make every template fail to parse in this test only.
func templateFuncs() template.FuncMap {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	registerFuncs(engine)
	return engine.FuncMap
}

// scanTemplate walks the parse trees of one file. Every {{ define }} block is a
// tree of its own, so all of them are visited.
func scanTemplate(file, source string, parsed *template.Template) ([]keyUse, []unresolvedUse) {
	pipes := make([]*parse.PipeNode, 0)
	for _, tmpl := range parsed.Templates() {
		if tmpl.Tree == nil {
			continue
		}
		collectPipes(tmpl.Root, &pipes)
	}

	variables := literalVariables(pipes)
	uses := make([]keyUse, 0)
	unresolved := make([]unresolvedUse, 0)

	for _, pipe := range pipes {
		for _, command := range pipe.Cmds {
			if !isCall(command, localizeFuncName) {
				continue
			}
			at := fmt.Sprintf("%s:%d", file, lineOf(source, int(command.Position())))
			if len(command.Args) != localizeArgCount {
				unresolved = append(unresolved, unresolvedUse{source: at, expr: command.String()})
				continue
			}
			keys, composed, ok := resolveKeyArg(command.Args[2], variables)
			if !ok {
				unresolved = append(unresolved, unresolvedUse{source: at, expr: command.String()})
				continue
			}
			for _, key := range keys {
				uses = append(uses, keyUse{key: key, source: at, composed: composed})
			}
		}
	}
	return uses, unresolved
}

// resolveKeyArg works out which keys one localize argument can stand for. A
// string literal is one key; a printf whose format is a literal and whose
// arguments are variables holding literals is one key per combination, which is
// how the create/update user form gets both of its prefixes covered.
func resolveKeyArg(arg parse.Node, variables map[string][]string) ([]string, bool, bool) {
	if literal, ok := arg.(*parse.StringNode); ok {
		return []string{literal.Text}, false, true
	}

	pipe, ok := arg.(*parse.PipeNode)
	if !ok || len(pipe.Cmds) != 1 {
		return nil, false, false
	}
	command := pipe.Cmds[0]
	if !isCall(command, printfFuncName) || len(command.Args) < 2 {
		return nil, false, false
	}
	format, ok := command.Args[1].(*parse.StringNode)
	if !ok {
		return nil, false, false
	}

	candidates := make([][]string, 0, len(command.Args)-2)
	for _, formatArg := range command.Args[2:] {
		variable, isVariable := formatArg.(*parse.VariableNode)
		if !isVariable || len(variable.Ident) != 1 {
			return nil, false, false
		}
		values := variables[variable.Ident[0]]
		if len(values) == 0 {
			return nil, false, false
		}
		candidates = append(candidates, values)
	}

	keys := make([]string, 0)
	for _, combination := range combinations(candidates) {
		values := make([]any, 0, len(combination))
		for _, value := range combination {
			values = append(values, value)
		}
		keys = append(keys, fmt.Sprintf(format.Text, values...))
	}
	return keys, true, len(keys) > 0
}

// literalVariables collects every template variable that is ever assigned a
// string literal, together with *all* the values it is assigned. $prefix in
// user_form.gohtml holds "auth.create_user" on one branch and
// "auth.update_user" on the other, and both have to be covered.
func literalVariables(pipes []*parse.PipeNode) map[string][]string {
	variables := make(map[string][]string)
	for _, pipe := range pipes {
		if len(pipe.Decl) == 0 || len(pipe.Cmds) != 1 || len(pipe.Cmds[0].Args) != 1 {
			continue
		}
		literal, ok := pipe.Cmds[0].Args[0].(*parse.StringNode)
		if !ok {
			continue
		}
		for _, declared := range pipe.Decl {
			if len(declared.Ident) != 1 {
				continue
			}
			name := declared.Ident[0]
			if !slices.Contains(variables[name], literal.Text) {
				variables[name] = append(variables[name], literal.Text)
			}
		}
	}
	return variables
}

// combinations is the cartesian product of the candidate values of every printf
// argument. With no arguments it yields the single empty combination, so a
// format with no verbs still resolves to itself.
func combinations(candidates [][]string) [][]string {
	result := [][]string{{}}
	for _, options := range candidates {
		next := make([][]string, 0, len(result)*len(options))
		for _, prefix := range result {
			for _, option := range options {
				next = append(next, append(slices.Clone(prefix), option))
			}
		}
		result = next
	}
	return result
}

// isCall reports whether a command is a direct call to the named function.
func isCall(command *parse.CommandNode, name string) bool {
	if len(command.Args) == 0 {
		return false
	}
	identifier, ok := command.Args[0].(*parse.IdentifierNode)
	return ok && identifier.Ident == name
}

// collectPipes gathers every pipeline in a parse tree, including the ones
// nested inside if/range/with bodies and inside parentheses.
func collectPipes(node parse.Node, into *[]*parse.PipeNode) {
	switch typed := node.(type) {
	case *parse.ListNode:
		if typed == nil {
			return
		}
		for _, child := range typed.Nodes {
			collectPipes(child, into)
		}
	case *parse.ActionNode:
		collectPipes(typed.Pipe, into)
	case *parse.PipeNode:
		if typed == nil {
			return
		}
		*into = append(*into, typed)
		for _, command := range typed.Cmds {
			collectPipes(command, into)
		}
	case *parse.CommandNode:
		for _, arg := range typed.Args {
			collectPipes(arg, into)
		}
	case *parse.IfNode:
		collectBranch(&typed.BranchNode, into)
	case *parse.RangeNode:
		collectBranch(&typed.BranchNode, into)
	case *parse.WithNode:
		collectBranch(&typed.BranchNode, into)
	case *parse.TemplateNode:
		collectPipes(typed.Pipe, into)
	}
}

func collectBranch(branch *parse.BranchNode, into *[]*parse.PipeNode) {
	collectPipes(branch.Pipe, into)
	if branch.List != nil {
		collectPipes(branch.List, into)
	}
	if branch.ElseList != nil {
		collectPipes(branch.ElseList, into)
	}
}

// lineOf turns a parse position into a line number, so a failure points at the
// line the reader has to open.
func lineOf(source string, position int) int {
	if position > len(source) {
		position = len(source)
	}
	return 1 + strings.Count(source[:position], "\n")
}

func sortedUnique(values []string) []string {
	unique := make([]string, 0, len(values))
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			unique = append(unique, value)
		}
	}
	sort.Strings(unique)
	return unique
}
