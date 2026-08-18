package main

import (
	"errors"
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

var (
	ErrContextRequired = errors.New("--context is required")
	ErrNameRequired    = errors.New("--name is required")
)

// ModuleData is everything the templates see. Nothing is computed inside a
// template: if a spelling or an ordering is needed, it is derived here, once,
// so two templates cannot disagree.
type ModuleData struct {
	Context Names
	Entity  Names

	Module string
	// LookupColumn is the label a relation <select> shows. Lookups select it
	// alongside the id instead of loading whole rows.
	LookupColumn string
	// LookupPascal is LookupColumn as a Go field name.
	LookupPascal string
	// UpdateColumns is the explicit column list of the update statement.
	// Select("*") would work but hides what an update touches, and the
	// generator knows the fields, so it can name them.
	UpdateColumns string

	Fields []Field
	Roles  []string
	// RoleConstants are the Go identifiers of the roles allowed on the routes.
	RoleConstants []string
	// BoolFields get a label/badge pair of helpers.
	BoolFields   []Field
	DomainFields []Field
	EntityFields []Field
	DTOFields    []Field
	RefFields    []Field
	TableFields  []Field

	HasTime  bool
	HasRef   bool
	HasMoney bool
	HasParse bool
}

type moduleFlags struct {
	context string
	name    string
	plural  string
	root    string
	roles   stringList
	fields  stringList
}

type stringList []string

func (s *stringList) String() string { return strings.Join(*s, ",") }

func (s *stringList) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func runModule(args []string) error {
	var opts moduleFlags
	set := flag.NewFlagSet("module", flag.ContinueOnError)
	set.StringVar(&opts.context, "context", "", "bounded context the module belongs to, e.g. billing")
	set.StringVar(&opts.name, "name", "", "entity name in snake_case, e.g. invoice")
	set.StringVar(&opts.plural, "plural", "", "override the guessed plural, e.g. invoices")
	set.Var(&opts.roles, "roles", "role allowed to reach the routes; repeatable")
	set.Var(&opts.fields, "field", "field spec name:type[:required]; repeatable")
	set.StringVar(&opts.root, "root", ".", "project root to generate into")
	if err := set.Parse(args); err != nil {
		return fmt.Errorf("parse module flags failed: %w", err)
	}

	if opts.context == "" {
		return ErrContextRequired
	}
	if opts.name == "" {
		return ErrNameRequired
	}

	data, err := buildModuleData(opts)
	if err != nil {
		return err
	}
	return generateModule(opts.root, data)
}

func buildModuleData(opts moduleFlags) (ModuleData, error) {
	modulePath, err := readModulePath(opts.root)
	if err != nil {
		return ModuleData{}, err
	}

	fields := make([]Field, 0, len(opts.fields))
	for _, spec := range opts.fields {
		field, parseErr := ParseField(spec)
		if parseErr != nil {
			return ModuleData{}, parseErr
		}
		fields = append(fields, field)
	}

	roles := opts.roles
	if len(roles) == 0 {
		roles = stringList{"admin"}
	}

	data := ModuleData{
		Module:  modulePath,
		Context: NewNames(opts.context, ""),
		Entity:  NewNames(opts.name, opts.plural),
		Fields:  fields,
		Roles:   roles,
	}
	classifyFields(&data)
	return data, nil
}

// classifyFields precomputes the orderings the templates need. Struct field
// order is not cosmetic here: govet runs with enable-all, so fieldalignment
// rejects a struct whose pointer-bearing fields are not grouped at the front.
func classifyFields(data *ModuleData) {
	for _, field := range data.Fields {
		switch {
		case field.NeedsTime():
			data.HasTime = true
		case field.Type == TypeMoney:
			data.HasMoney = true
		}
		if field.NeedsParse() {
			data.HasParse = true
		}
		if field.IsRef() {
			data.HasRef = true
			data.RefFields = append(data.RefFields, field)
		}
	}

	for _, role := range data.Roles {
		data.RoleConstants = append(data.RoleConstants, toPascal(toSnake(role))+"Role")
	}
	data.LookupColumn = lookupColumn(data.Fields)
	data.LookupPascal = toPascal(data.LookupColumn)
	for _, field := range data.Fields {
		if field.Type == TypeBool {
			data.BoolFields = append(data.BoolFields, field)
		}
	}
	data.UpdateColumns = updateColumns(data.Fields)
	data.DomainFields = orderForAlignment(data.Fields, true)
	data.EntityFields = orderForAlignment(data.Fields, false)
	data.DTOFields = orderDTOFields(data.Fields)
	data.TableFields = tableFields(data.Fields)
}

// orderForAlignment groups fields the way fieldalignment wants them: the widest
// pointer-bearing types first, scalars last. withID adds the primary key, which
// the entity inherits from gorm.Model and therefore does not declare.
func orderForAlignment(fields []Field, withID bool) []Field {
	var times, refs, strs, nums, bools []Field
	for _, field := range fields {
		switch field.Type {
		case TypeDate, TypeDateTime:
			times = append(times, field)
		case TypeString, TypeText, TypeEmail:
			strs = append(strs, field)
		case TypeBool:
			bools = append(bools, field)
		case TypeRef:
			refs = append(refs, field)
			nums = append(nums, field)
		default:
			nums = append(nums, field)
		}
	}

	ordered := make([]Field, 0, len(fields)+1)
	ordered = append(ordered, times...)
	for _, ref := range refs {
		ordered = append(ordered, relationField(ref))
	}
	ordered = append(ordered, strs...)
	if withID {
		ordered = append(ordered, idField())
	}
	ordered = append(ordered, nums...)
	return append(ordered, bools...)
}

// relationField is the pointer to the referenced entity that sits beside the
// foreign key, so a listing can show the related name without a query per row.
func relationField(ref Field) Field {
	relation := ref
	relation.Names = NewNames(strings.TrimSuffix(ref.Snake, "_id"), "")
	relation.Type = "relation"
	relation.DomainType = "*" + ref.RefEntity.Pascal
	relation.GormTag = fmt.Sprintf(
		"foreignKey:%s;constraint:OnDelete:RESTRICT", ref.Pascal)
	return relation
}

func idField() Field {
	return Field{
		Names:        NewNames("id", ""),
		Type:         TypeUint,
		DomainType:   "uint",
		DTOType:      "uint",
		IsPrimaryKey: true,
	}
}

// orderDTOFields keeps the DTO aligned too, with ID between the strings and the
// numbers exactly as in the domain.
func orderDTOFields(fields []Field) []Field {
	var strs, nums, bools []Field
	for _, field := range fields {
		switch field.DTOType {
		case TypeString:
			strs = append(strs, field)
		case TypeBool:
			bools = append(bools, field)
		default:
			nums = append(nums, field)
		}
	}
	ordered := make([]Field, 0, len(fields)+1)
	ordered = append(ordered, strs...)
	ordered = append(ordered, idField())
	ordered = append(ordered, nums...)
	ordered = append(ordered, bools...)
	return alignTags(ordered)
}

// alignTags pads every binding tag to the widest one so the form tags line up
// in a column. tagalign enforces this, so a generator that skipped it would
// emit code that fails the project's own lint on the first run.
func alignTags(fields []Field) []Field {
	widest := 0
	for _, field := range fields {
		if tag := field.BindingTag(); tag != "" {
			if width := len(tag) + len(`binding:""`); width > widest {
				widest = width
			}
		}
	}
	for index, field := range fields {
		tag := field.BindingTag()
		if tag == "" {
			continue
		}
		rendered := fmt.Sprintf("binding:%q", tag)
		fields[index].TagPrefix = rendered + strings.Repeat(" ", widest-len(rendered)+1)
	}
	return fields
}

// tableFields picks the columns a listing shows. Free text is left out: a
// textarea does not belong in a table cell, and showing it makes every row a
// different height.
func tableFields(fields []Field) []Field {
	columns := make([]Field, 0, len(fields))
	for _, field := range fields {
		if field.Type == TypeText || field.Type == TypeDateTime {
			continue
		}
		columns = append(columns, field)
	}
	return columns
}

// readModulePath takes the module path from go.mod rather than a flag: it is
// already declared there, and a mismatch would produce imports that do not
// resolve.
func readModulePath(root string) (string, error) {
	content, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", fmt.Errorf("read go.mod failed: %w", err)
	}
	for line := range strings.SplitSeq(string(content), "\n") {
		if after, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			return strings.TrimSpace(after), nil
		}
	}
	return "", fmt.Errorf("read go.mod failed: %w", ErrNoModuleDirective)
}

var ErrNoModuleDirective = errors.New("go.mod has no module directive")

// renderGo renders a template and gofmts the result. Formatting the output
// rather than aligning inside the template is what lets the templates stay
// readable next to the code they reproduce.
func renderGo(tmpl *template.Template, name string, data ModuleData) ([]byte, error) {
	var buf strings.Builder
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return nil, fmt.Errorf("render %s failed: %w", name, err)
	}
	formatted, err := format.Source([]byte(pruneUnusedNolint(buf.String())))
	if err != nil {
		return nil, fmt.Errorf("gofmt %s failed: %w\n%s", name, err, buf.String())
	}
	return formatted, nil
}

func renderRaw(tmpl *template.Template, name string, data ModuleData) ([]byte, error) {
	var buf strings.Builder
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return nil, fmt.Errorf("render %s failed: %w", name, err)
	}
	return []byte(buf.String()), nil
}

// lookupColumn is the first text field, which is what a human recognises the
// record by. With no text field at all there is nothing better than the id.
func lookupColumn(fields []Field) string {
	for _, field := range fields {
		if field.IsStringy() {
			return field.Snake
		}
	}
	return "id"
}

// updateColumns lists the columns in declared order, plus updated_at, which
// GORM will not refresh unless it is selected. It wraps at a width that keeps
// the rendered line inside the project's 120 column limit however many fields
// the module has.
func updateColumns(fields []Field) string {
	columns := make([]string, 0, len(fields)+1)
	for _, field := range fields {
		columns = append(columns, strconv.Quote(field.Snake))
	}
	columns = append(columns, strconv.Quote("updated_at"))

	const (
		indent   = "\t\t\t"
		maxWidth = 100
	)
	var (
		builder strings.Builder
		width   int
	)
	for index, column := range columns {
		if index > 0 {
			builder.WriteString(",")
			width++
			if width+len(column) > maxWidth {
				builder.WriteString("\n" + indent)
				width = len(indent)
			} else {
				builder.WriteString(" ")
				width++
			}
		}
		builder.WriteString(column)
		width += len(column)
	}
	return builder.String()
}
