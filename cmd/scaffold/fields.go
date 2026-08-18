package main

import (
	"errors"
	"fmt"
	"strings"
)

// Field type identifiers accepted by --field. Each one decides, together, the
// Go type in three layers, the GORM column, the binding rule, the form control,
// the table cell and the i18n key — which is why they are a closed set and not
// free text.
const (
	TypeString   = "string"
	TypeText     = "text"
	TypeInt      = "int"
	TypeUint     = "uint"
	TypeBool     = "bool"
	TypeDate     = "date"
	TypeDateTime = "datetime"
	TypeMoney    = "money"
	TypeEmail    = "email"
	TypeRef      = "ref"

	// goTypeTime is the Go type a date or datetime lands on.
	goTypeTime = "time.Time"
)

var (
	ErrEmptyFieldSpec   = errors.New("field spec is empty")
	ErrUnknownFieldType = errors.New("unknown field type")
	ErrRefNeedsEntity   = errors.New("a ref field needs a target: name:ref=other_entity")
	ErrReservedField    = errors.New("field name is reserved by the generated struct")
)

// reservedFieldNames are declared by every generated entity, so a module field
// of the same name would collide silently at compile time or, worse, shadow a
// gorm.Model column at runtime.
var reservedFieldNames = map[string]bool{ //nolint:gochecknoglobals
	"id": true, "created_at": true, "updated_at": true, "deleted_at": true,
}

// Field is one declared column, with every spelling and type the templates
// need already derived. Templates branch on Type rather than receiving code
// strings, so the Go they emit stays visible in the .tmpl files where it can be
// diffed against the hand-written reference module.
type Field struct {
	Names

	RefEntity Names

	Type string
	// TagPrefix is the binding tag padded so the form tags line up, which is
	// what tagalign wants. Emitting it aligned means generated code passes
	// make lint without a fix-up pass.
	TagPrefix  string
	DomainType string
	DTOType    string
	GormTag    string
	FormType   string

	Required bool
	// IsPrimaryKey marks the synthetic id. It is not a declared field: it
	// carries no binding rule and is never rendered as a form control.
	IsPrimaryKey bool
}

// IsRef reports whether the field is a foreign key to another entity.
func (f Field) IsRef() bool { return f.Type == TypeRef }

// RelationPascal is the name of the struct field holding the referenced entity,
// which is the foreign key without its _id suffix: category_id -> Category.
func (f Field) RelationPascal() string {
	return toPascal(strings.TrimSuffix(f.Snake, "_id"))
}

// TableAlign is the column alignment for this field's cell. Numbers read as a
// column when they are right aligned; a badge reads as one when it is centred.
func (f Field) TableAlign() string {
	switch f.Type {
	case TypeInt, TypeUint, TypeMoney:
		return "right"
	case TypeBool:
		return "center"
	default:
		return ""
	}
}

// IsStringy reports whether the field is stored as text, which is what makes it
// a candidate for the lookup label.
func (f Field) IsStringy() bool {
	return f.Type == TypeString || f.Type == TypeText || f.Type == TypeEmail
}

// NeedsParse reports whether converting the submitted DTO value to the domain
// can fail. Those fields make the DTO->domain mapper return an error.
func (f Field) NeedsParse() bool {
	return f.Type == TypeMoney || f.Type == TypeDate || f.Type == TypeDateTime
}

// NeedsTime reports whether the field puts time.Time in the domain and entity.
func (f Field) NeedsTime() bool {
	return f.Type == TypeDate || f.Type == TypeDateTime
}

// BindingTag is the gin binding rule. It is `binding:`, never `validate:` —
// gin's binder ignores the latter, so a validate tag looks like validation
// while silently never running.
func (f Field) BindingTag() string {
	if f.IsPrimaryKey {
		return ""
	}
	rules := make([]string, 0, 2)
	if f.Required {
		rules = append(rules, "required")
	}
	switch f.Type {
	case TypeMoney:
		if !f.Required {
			rules = append(rules, "omitempty")
		}
		rules = append(rules, "decimal2")
	case TypeEmail:
		if !f.Required {
			rules = append(rules, "omitempty")
		}
		rules = append(rules, "email")
	case TypeUint:
		if !f.Required {
			rules = append(rules, "gte=0")
		}
	}
	return strings.Join(rules, ",")
}

// ParseField reads one --field value: name:type, name:type:required, or
// name:ref=other[:required].
func ParseField(spec string) (Field, error) {
	parts := strings.Split(strings.TrimSpace(spec), ":")
	if len(parts) < 2 || parts[0] == "" {
		return Field{}, fmt.Errorf("%w: %q", ErrEmptyFieldSpec, spec)
	}

	name := toSnake(parts[0])
	if reservedFieldNames[name] {
		return Field{}, fmt.Errorf("%w: %q", ErrReservedField, name)
	}

	declared := parts[1]
	required := len(parts) > 2 && parts[2] == "required"

	field := Field{
		Names:    NewNames(name, ""),
		Required: required,
	}

	if target, isRef := strings.CutPrefix(declared, TypeRef+"="); isRef {
		if target == "" {
			return Field{}, fmt.Errorf("%w: %q", ErrRefNeedsEntity, spec)
		}
		field.Type = TypeRef
		field.RefEntity = NewNames(toSnake(target), "")
	} else {
		field.Type = declared
	}

	derived, err := deriveTypes(field)
	if err != nil {
		return Field{}, err
	}
	return derived, nil
}

// deriveTypes fills in the type table: one row per accepted field type,
// deciding the domain type, the DTO type, the GORM column and the form control
// together. It is a function rather than a pointer method so Field keeps a
// single receiver kind.
//
//nolint:funlen // a table with one branch per accepted type; splitting hides it
func deriveTypes(field Field) (Field, error) {
	switch field.Type {
	case TypeString:
		field.DomainType, field.DTOType = TypeString, TypeString
		field.GormTag = "type:text;size:255"
		field.FormType = "FieldTypeText"
	case TypeText:
		field.DomainType, field.DTOType = TypeString, TypeString
		field.GormTag = "type:text"
		field.FormType = "FieldTypeTextarea"
	case TypeInt:
		field.DomainType, field.DTOType = TypeInt, TypeInt
		field.GormTag = "type:integer;default:0"
		field.FormType = "FieldTypeNumber"
	case TypeUint:
		field.DomainType, field.DTOType = TypeUint, TypeUint
		field.GormTag = "type:integer;default:0"
		field.FormType = "FieldTypeNumber"
	case TypeBool:
		field.DomainType, field.DTOType = TypeBool, TypeBool
		field.GormTag = "type:boolean;default:false"
		field.FormType = "FieldTypeCheckbox"
	case TypeDate:
		// The DTO stays a string: it holds the <input type="date"> wire format
		// and is parsed by the mapper, so a malformed date is a validation
		// error and not a silently zeroed column.
		field.DomainType, field.DTOType = goTypeTime, TypeString
		field.GormTag = "type:date"
		field.FormType = "FieldTypeDate"
	case TypeDateTime:
		field.DomainType, field.DTOType = goTypeTime, TypeString
		field.GormTag = "type:timestamptz"
		field.FormType = "FieldTypeDateTime"
	case TypeMoney:
		// Stored in cents as an integer. Money in a float is a rounding bug
		// waiting for its first invoice.
		field.DomainType, field.DTOType = TypeUint, TypeString
		field.GormTag = "type:bigint;default:0"
		field.FormType = "FieldTypeMoney"
	case TypeEmail:
		field.DomainType, field.DTOType = TypeString, TypeString
		field.GormTag = "type:text;size:255"
		field.FormType = "FieldTypeEmail"
	case TypeRef:
		field.DomainType, field.DTOType = TypeUint, TypeUint
		field.GormTag = "index"
		field.FormType = "FieldTypeSelect"
	default:
		return Field{}, fmt.Errorf("%w: %q", ErrUnknownFieldType, field.Type)
	}

	if field.Required {
		field.GormTag += ";not null"
	}
	return field, nil
}
