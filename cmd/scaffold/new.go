package main

import (
	"errors"
	"flag"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// dirMode and fileMode are the permissions generated files are created with:
// readable by the owner's group, never world writable.
const (
	dirMode  = 0o750
	fileMode = 0o600
)

var (
	ErrProjectNameRequired = errors.New("--name is required")
	ErrModulePathRequired  = errors.New("--module is required")
	ErrOutRequired         = errors.New("--out is required")
	ErrOutNotEmpty         = errors.New("--out already exists and is not empty")
)

// skippedDirs never travel into a new project: build output, tool downloads and
// the template's own git history.
var skippedDirs = map[string]bool{ //nolint:gochecknoglobals
	".git": true, "bin": true, "reports": true, "tools": true, "node_modules": true,
}

// skippedFiles are the ones a new project must provide itself. The real
// configuration files (yaml, and json — the fallback format) are deliberately
// absent so startup fails loudly rather than running on a configuration — or
// worse, credentials — inherited from the template.
var skippedFiles = map[string]bool{ //nolint:gochecknoglobals
	filepath.Join("cmd", "server", "config", "application.yaml"): true,
	filepath.Join("cmd", "server", "config", "application.json"): true,
}

type newFlags struct {
	name      string
	module    string
	out       string
	roles     string
	root      string
	noExample bool
}

func runNew(args []string) error {
	var opts newFlags
	set := flag.NewFlagSet("new", flag.ContinueOnError)
	set.StringVar(&opts.name, "name", "", "service name, e.g. mi-erp")
	set.StringVar(&opts.module, "module", "", "Go module path, e.g. github.com/acme/mi-erp")
	set.StringVar(&opts.out, "out", "", "directory to create the project in")
	set.StringVar(&opts.roles, "roles", "admin,user", "comma separated roles for the project")
	set.BoolVar(&opts.noExample, "no-example", false, "drop the example module")
	set.StringVar(&opts.root, "root", ".", "template root to copy from")
	if err := set.Parse(args); err != nil {
		return fmt.Errorf("parse new flags failed: %w", err)
	}

	switch {
	case opts.name == "":
		return ErrProjectNameRequired
	case opts.module == "":
		return ErrModulePathRequired
	case opts.out == "":
		return ErrOutRequired
	}

	if entries, err := os.ReadDir(opts.out); err == nil && len(entries) > 0 {
		return fmt.Errorf("%w: %s", ErrOutNotEmpty, opts.out)
	}

	sourceModule, err := readModulePath(opts.root)
	if err != nil {
		return err
	}

	if err = copyTree(opts, sourceModule); err != nil {
		return err
	}
	if err = writeRoles(opts); err != nil {
		return err
	}
	if opts.noExample {
		if err = dropExample(opts); err != nil {
			return err
		}
	}

	fmt.Printf("  created %s (module %s)\n", opts.out, opts.module)
	fmt.Printf("  next: cd %s && cp cmd/server/config/application.example.yaml "+
		"cmd/server/config/application.yaml && make build\n", opts.out)
	return nil
}

func copyTree(opts newFlags, sourceModule string) error {
	if err := walkTemplate(opts, sourceModule); err != nil {
		return fmt.Errorf("copy template failed: %w", err)
	}
	return nil
}

func walkTemplate(opts newFlags, sourceModule string) error {
	//nolint:wrapcheck // the callback below already wraps everything it returns
	return filepath.WalkDir(opts.root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(opts.root, path)
		if relErr != nil {
			return fmt.Errorf("relative path of %s failed: %w", path, relErr)
		}
		if rel == "." {
			return nil
		}
		if entry.IsDir() {
			if skippedDirs[entry.Name()] {
				return filepath.SkipDir
			}
			if mkdirErr := os.MkdirAll(filepath.Join(opts.out, rel), dirMode); mkdirErr != nil {
				return fmt.Errorf("create %s failed: %w", rel, mkdirErr)
			}
			return nil
		}
		if skippedFiles[rel] {
			return nil
		}
		return copyFile(path, filepath.Join(opts.out, rel), opts, sourceModule)
	})
}

func copyFile(from, target string, opts newFlags, sourceModule string) error {
	content, err := os.ReadFile(from)
	if err != nil {
		return fmt.Errorf("read %s failed: %w", from, err)
	}
	rewritten := strings.ReplaceAll(string(content), sourceModule, opts.module)
	if filepath.Base(target) == "Makefile" || strings.HasSuffix(target, "env.go") {
		rewritten = strings.ReplaceAll(rewritten, "go-web-skeleton", opts.name)
	}
	if err = os.MkdirAll(filepath.Dir(target), dirMode); err != nil {
		return fmt.Errorf("create %s failed: %w", filepath.Dir(target), err)
	}
	if err = os.WriteFile(target, []byte(rewritten), fileMode); err != nil {
		return fmt.Errorf("write %s failed: %w", target, err)
	}
	return nil
}

// writeRoles regenerates the single source of truth for the project's roles.
func writeRoles(opts newFlags) error {
	roles := make([]string, 0, 2)
	for role := range strings.SplitSeq(opts.roles, ",") {
		if trimmed := strings.TrimSpace(role); trimmed != "" {
			roles = append(roles, toSnake(trimmed))
		}
	}
	if len(roles) == 0 {
		roles = []string{"admin"}
	}

	var builder strings.Builder
	builder.WriteString("package domain\n\ntype Role string\n\n")
	builder.WriteString("// The block below is the single source of truth for the roles of the project:\n")
	builder.WriteString("// `scaffold new --roles a,b,c` rewrites this file wholesale. Declare roles here\n")
	builder.WriteString("// and nowhere else — the middleware.Authorize arguments and the \"isvalidrole\"\n")
	builder.WriteString("// validator, fed by GetAllowedRoles, both derive from it.\nconst (\n")
	for _, role := range roles {
		fmt.Fprintf(&builder, "\t%sRole Role = %q\n", toPascal(role), role)
	}
	builder.WriteString(")\n\nvar allowedRoles = []Role{")
	for index, role := range roles {
		if index > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(toPascal(role) + "Role")
	}
	builder.WriteString("} //nolint:gochecknoglobals\n\n")
	builder.WriteString("func GetAllowedRoles() []string {\n")
	builder.WriteString("\tallowedRolesStr := make([]string, 0, len(allowedRoles))\n")
	builder.WriteString("\tfor _, r := range allowedRoles {\n")
	builder.WriteString("\t\tallowedRolesStr = append(allowedRolesStr, string(r))\n\t}\n")
	builder.WriteString("\treturn allowedRolesStr\n}\n")

	formatted, err := format.Source([]byte(builder.String()))
	if err != nil {
		return fmt.Errorf("gofmt roles failed: %w", err)
	}
	target := filepath.Join(opts.out, "common", "domain", "roles.go")
	if err = os.WriteFile(target, formatted, 0o600); err != nil {
		return fmt.Errorf("write roles failed: %w", err)
	}
	return nil
}

// dropExample removes the demonstration module and everything that referenced
// it, so --no-example leaves a project that still builds.
func dropExample(opts newFlags) error {
	paths := []string{
		filepath.Join(opts.out, "example"),
		filepath.Join(opts.out, "common", "http", "routes", "item.go"),
		filepath.Join(opts.out, "common", "http", "routes", "item_category.go"),
		filepath.Join(opts.out, "common", "http", "templates", "files", "pages", "item"),
		filepath.Join(opts.out, "common", "http", "templates", "files", "pages", "item_category"),
	}
	for _, glob := range []string{"example.*.json", "item.*.json", "item_category.*.json"} {
		matches, err := filepath.Glob(filepath.Join(opts.out, "common", "i18n", "files", glob))
		if err != nil {
			return fmt.Errorf("drop example failed: %w", err)
		}
		paths = append(paths, matches...)
	}
	for _, path := range paths {
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("drop example failed: %w", err)
		}
	}
	return removeExampleReferences(opts.out)
}

// removeExampleReferences strips the generated lines out of the shared marker
// files. They were inserted by `scaffold module`, so they are removed the same
// way: by line, not by AST.
func removeExampleReferences(out string) error {
	targets := []string{
		filepath.Join(out, "common", "http", "routes", "routes.go"),
		filepath.Join(out, "common", "wire", "inject_repositories.go"),
		filepath.Join(out, "common", "wire", "inject_services.go"),
		filepath.Join(out, "common", "wire", "inject_http_handlers.go"),
		filepath.Join(out, "common", "bootstrap", "db_migration.go"),
		filepath.Join(out, "common", "http", "templates", "files", "layouts", "sidebar.gohtml"),
	}
	for _, target := range targets {
		content, err := os.ReadFile(target)
		if err != nil {
			return fmt.Errorf("drop example failed: %w", err)
		}
		kept := stripExampleLines(strings.Split(string(content), "\n"))
		if err = os.WriteFile(target, []byte(strings.Join(kept, "\n")), 0o600); err != nil {
			return fmt.Errorf("drop example failed: %w", err)
		}
	}
	return nil
}

// exampleIdentifiers are the exact identifiers `scaffold module` inserted for
// the demonstration module.
var exampleIdentifiers = []string{ //nolint:gochecknoglobals
	"Item", "ItemCategory", "examplerepository", "exampleservice",
	"examplehandler", "exampleentity", `"item.title"`, `"item_category.title"`,
	`"example.title"`, "/item", "/item-category",
}

// stripExampleLines removes the generated lines, and in a template removes the
// whole element rather than the line that named it.
//
// A sidebar entry is an <a> spanning four lines and only the first and third
// mention the module; dropping those by themselves leaves a dangling <svg> and
// </a> behind — markup that still compiles and still renders, as two icons
// pointing nowhere.
func stripExampleLines(lines []string) []string {
	kept := make([]string, 0, len(lines))
	skippingElement := false
	for _, line := range lines {
		if skippingElement {
			if strings.Contains(line, "</a>") {
				skippingElement = false
			}
			continue
		}
		if !isExampleReference(line) {
			kept = append(kept, line)
			continue
		}
		// An opening tag that is not closed on the same line takes its whole
		// element with it.
		if strings.Contains(line, "<a ") && !strings.Contains(line, "</a>") {
			skippingElement = true
		}
	}
	return kept
}

// isExampleReference decides whether a line was generated for the example
// module. A marker line is never one, whatever it contains: scaffold:sidebar:items
// has "item" inside it, and removing it would leave the project unable to
// generate anything ever again.
func isExampleReference(line string) bool {
	if strings.Contains(line, "scaffold:") {
		return false
	}
	for _, identifier := range exampleIdentifiers {
		if strings.Contains(line, identifier) {
			return true
		}
	}
	return false
}
