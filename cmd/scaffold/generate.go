package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// artifact is one file the generator writes: which template renders it, where
// it lands, and whether it is Go (and therefore gofmt'ed) or not.
type artifact struct {
	template string
	path     string
	gofmt    bool
	// perContext files are written once for a bounded context and left alone
	// afterwards, so a second module does not overwrite the first one's.
	perContext bool
}

func loadTemplates() (*template.Template, error) {
	funcs := template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}
	tmpl, err := template.New("scaffold").Funcs(funcs).ParseFS(templateFS, "templates/*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("parse templates failed: %w", err)
	}
	return tmpl, nil
}

func moduleArtifacts(data ModuleData) []artifact {
	ctx := data.Context.Snake
	name := data.Entity.Snake
	return []artifact{
		{"mapper_common.go", filepath.Join(ctx, "http", "mapper", "common.go"), true, true},
		{"context_locale.json", filepath.Join("common", "i18n", "files", ctx+".en.json"), false, true},
		{"context_locale.json", filepath.Join("common", "i18n", "files", ctx+".es.json"), false, true},

		{"domain.go", filepath.Join(ctx, "domain", name+".go"), true, false},
		{"entity.go", filepath.Join(ctx, "repository", "rdbms", "entity", name+".go"), true, false},
		{"rdbms_mapper.go", filepath.Join(ctx, "repository", "rdbms", "mapper", name+".go"), true, false},
		{"repository.go", filepath.Join(ctx, "repository", "rdbms", name+".go"), true, false},
		{"service.go", filepath.Join(ctx, "service", name+".go"), true, false},
		{"dto.go", filepath.Join(ctx, "http", "dto", name+".go"), true, false},
		{"http_mapper.go", filepath.Join(ctx, "http", "mapper", name+".go"), true, false},
		{"handler.go", filepath.Join(ctx, "http", "handler", name+".go"), true, false},
		{"routes.go", filepath.Join("common", "http", "routes", name+".go"), true, false},
		{"list_page.gohtml", filepath.Join("common", "http", "templates", "files", "pages",
			name, "list_"+data.Entity.SnakePlural+".gohtml"), false, false},
		{"locale.json", filepath.Join("common", "i18n", "files", name+".en.json"), false, false},
		{"locale.json", filepath.Join("common", "i18n", "files", name+".es.json"), false, false},
	}
}

func generateModule(root string, data ModuleData) error {
	tmpl, err := loadTemplates()
	if err != nil {
		return err
	}

	written := make([]string, 0, len(moduleArtifacts(data)))
	for _, art := range moduleArtifacts(data) {
		target := filepath.Join(root, art.path)
		if art.perContext {
			if _, statErr := os.Stat(target); statErr == nil {
				continue
			}
		}
		var content []byte
		if art.gofmt {
			content, err = renderGo(tmpl, art.template, data)
		} else {
			content, err = renderRaw(tmpl, art.template, data)
		}
		if err != nil {
			return err
		}
		if err = os.MkdirAll(filepath.Dir(target), dirMode); err != nil {
			return fmt.Errorf("create %s failed: %w", filepath.Dir(target), err)
		}
		if err = os.WriteFile(target, content, fileMode); err != nil {
			return fmt.Errorf("write %s failed: %w", target, err)
		}
		written = append(written, art.path)
	}

	for _, injection := range moduleInjections(root, data) {
		if err = injection.Apply(); err != nil {
			return err
		}
	}

	for _, path := range written {
		fmt.Println("  created", path)
	}
	fmt.Printf("  edited 6 shared files through their scaffold: markers\n")
	return nil
}

// moduleInjections lists every edit to a shared file. These are the five Go
// files plus the sidebar: the only places a new module has to be mentioned by
// hand, and the reason each of them carries a marker comment.
func moduleInjections(root string, data ModuleData) []Injection {
	entity := data.Entity
	ctx := data.Context
	repoAlias := ctx.Camel + "repository"
	serviceAlias := ctx.Camel + "service"
	handlerAlias := ctx.Camel + "handler"
	entityAlias := ctx.Camel + "entity"

	join := func(parts ...string) string { return filepath.Join(append([]string{root}, parts...)...) }

	return []Injection{
		{join("common", "http", "routes", "routes.go"), "scaffold:routes:register",
			[]string{fmt.Sprintf("register%sRoutes(webRouter, container)", entity.Pascal)}},

		{join("common", "wire", "inject_repositories.go"), "scaffold:repositories:imports",
			[]string{fmt.Sprintf("%s %q", repoAlias, data.Module+"/"+ctx.Snake+"/repository/rdbms")}},
		{join("common", "wire", "inject_repositories.go"), "scaffold:repositories:fields",
			[]string{fmt.Sprintf("%sRepository *%s.%s", entity.Pascal, repoAlias, entity.Pascal)}},
		{join("common", "wire", "inject_repositories.go"), "scaffold:repositories:init",
			[]string{fmt.Sprintf("container.%sRepository = %s.New%s(container.PostgresClient)",
				entity.Pascal, repoAlias, entity.Pascal)}},

		{join("common", "wire", "inject_services.go"), "scaffold:services:imports",
			[]string{fmt.Sprintf("%s %q", serviceAlias, data.Module+"/"+ctx.Snake+"/service")}},
		{join("common", "wire", "inject_services.go"), "scaffold:services:fields",
			[]string{fmt.Sprintf("%sService *%s.%s", entity.Pascal, serviceAlias, entity.Pascal)}},
		{join("common", "wire", "inject_services.go"), "scaffold:services:init",
			[]string{fmt.Sprintf("container.%sService = %s.New%s(container.%sRepository)",
				entity.Pascal, serviceAlias, entity.Pascal, entity.Pascal)}},

		{join("common", "wire", "inject_http_handlers.go"), "scaffold:handlers:imports",
			[]string{fmt.Sprintf("%s %q", handlerAlias, data.Module+"/"+ctx.Snake+"/http/handler")}},
		{join("common", "wire", "inject_http_handlers.go"), "scaffold:handlers:fields",
			[]string{fmt.Sprintf("%sHandler *%s.%s", entity.Pascal, handlerAlias, entity.Pascal)}},
		{join("common", "wire", "inject_http_handlers.go"), "scaffold:handlers:init",
			[]string{handlerConstructor(data, handlerAlias)}},

		{join("common", "bootstrap", "db_migration.go"), "scaffold:migrations:imports",
			[]string{fmt.Sprintf("%s %q", entityAlias, data.Module+"/"+ctx.Snake+"/repository/rdbms/entity")}},
		{join("common", "bootstrap", "db_migration.go"), "scaffold:migrations:entities",
			[]string{fmt.Sprintf("&%s.%s{},", entityAlias, entity.Pascal)}},

		{join("common", "http", "templates", "files", "layouts", "sidebar.gohtml"), "scaffold:sidebar:items",
			sidebarLines(entity)},
	}
}

func handlerConstructor(data ModuleData, alias string) string {
	var call strings.Builder
	fmt.Fprintf(&call, "container.%sHandler = %s.New%s(container.%sService",
		data.Entity.Pascal, alias, data.Entity.Pascal, data.Entity.Pascal)
	for _, ref := range data.RefFields {
		fmt.Fprintf(&call, ", container.%sService", ref.RefEntity.Pascal)
	}
	call.WriteString(")")
	return call.String()
}

func sidebarLines(entity Names) []string {
	return []string{
		fmt.Sprintf(`<a href="/%s" class="sidebar-link" title="{{ localize .Language "%s.title" }}">`,
			entity.Kebab, entity.Snake),
		`  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">` +
			`<rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/>` +
			`<rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/></svg>`,
		fmt.Sprintf(`  <span class="sidebar-label">{{ localize .Language "%s.title" }}</span>`, entity.Snake),
		`</a>`,
	}
}
