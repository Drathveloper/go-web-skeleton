package main

// --field is the whole surface of `scaffold module`: one bad row in the type
// table and the generated code compiles but stores the wrong thing, or worse,
// validates nothing. Every accepted type is pinned here, and so is every
// rejection — a spec the generator silently accepts is a module the author gets
// to debug at runtime.

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseFieldAcceptsEveryDeclaredType(t *testing.T) {
	tests := map[string]struct {
		spec       string
		fieldType  string
		domainType string
		dtoType    string
		gormTag    string
		formType   string
	}{
		"string":   {"name:string", TypeString, "string", "string", "type:text;size:255", "FieldTypeText"},
		"text":     {"notes:text", TypeText, "string", "string", "type:text", "FieldTypeTextarea"},
		"int":      {"stock:int", TypeInt, "int", "int", "type:integer;default:0", "FieldTypeNumber"},
		"uint":     {"units:uint", TypeUint, "uint", "uint", "type:integer;default:0", "FieldTypeNumber"},
		"bool":     {"active:bool", TypeBool, "bool", "bool", "type:boolean;default:false", "FieldTypeCheckbox"},
		"date":     {"released_at:date", TypeDate, goTypeTime, "string", "type:date", "FieldTypeDate"},
		"datetime": {"starts_at:datetime", TypeDateTime, goTypeTime, "string", "type:timestamptz", "FieldTypeDateTime"},
		"money":    {"price:money", TypeMoney, "uint", "string", "type:bigint;default:0", "FieldTypeMoney"},
		"email":    {"contact:email", TypeEmail, "string", "string", "type:text;size:255", "FieldTypeEmail"},
		"ref":      {"category_id:ref=item_category", TypeRef, "uint", "uint", "index", "FieldTypeSelect"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			field, err := ParseField(test.spec)

			require.NoError(t, err)
			require.Equal(t, test.fieldType, field.Type)
			require.Equal(t, test.domainType, field.DomainType)
			require.Equal(t, test.dtoType, field.DTOType)
			require.Equal(t, test.gormTag, field.GormTag)
			require.Equal(t, test.formType, field.FormType)
			require.False(t, field.Required, "no :required suffix was given")
			require.False(t, field.IsPrimaryKey)
		})
	}
}

func TestParseFieldDerivesTheNameSpellings(t *testing.T) {
	field, err := ParseField("ReleasedAt:date")

	require.NoError(t, err)
	require.Equal(t, "released_at", field.Snake)
	require.Equal(t, "ReleasedAt", field.Pascal)
	require.Equal(t, "releasedAt", field.Camel)
	require.Equal(t, "Released at", field.Human)
}

func TestParseFieldRef(t *testing.T) {
	field, err := ParseField("category_id:ref=item_category")

	require.NoError(t, err)
	require.True(t, field.IsRef())
	require.Equal(t, "item_category", field.RefEntity.Snake)
	require.Equal(t, "ItemCategory", field.RefEntity.Pascal)
	require.Equal(t, "item_categories", field.RefEntity.SnakePlural)
	// The struct field holding the referenced entity drops the _id suffix.
	require.Equal(t, "Category", field.RelationPascal())
}

func TestParseFieldRequiredSuffix(t *testing.T) {
	tests := map[string]struct {
		spec    string
		gormTag string
	}{
		"string":   {"name:string:required", "type:text;size:255;not null"},
		"money":    {"price:money:required", "type:bigint;default:0;not null"},
		"ref":      {"category_id:ref=item_category:required", "index;not null"},
		"datetime": {"starts_at:datetime:required", "type:timestamptz;not null"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			field, err := ParseField(test.spec)

			require.NoError(t, err)
			require.True(t, field.Required)
			require.Equal(t, test.gormTag, field.GormTag)
		})
	}
}

// TestParseFieldIgnoresAnUnknownThirdSegment documents the current contract:
// only the exact word "required" turns the flag on, so a typo is not silently
// read as required.
func TestParseFieldIgnoresAnUnknownThirdSegment(t *testing.T) {
	field, err := ParseField("name:string:requird")

	require.NoError(t, err)
	require.False(t, field.Required)
}

func TestParseFieldRejections(t *testing.T) {
	tests := map[string]struct {
		expected error
		spec     string
	}{
		"empty spec":          {ErrEmptyFieldSpec, ""},
		"blank spec":          {ErrEmptyFieldSpec, "   "},
		"no type":             {ErrEmptyFieldSpec, "name"},
		"no name":             {ErrEmptyFieldSpec, ":string"},
		"unknown type":        {ErrUnknownFieldType, "name:blob"},
		"bare ref keyword":    {ErrRefNeedsEntity, "category:ref"},
		"ref without target":  {ErrRefNeedsEntity, "category_id:ref="},
		"reserved id":         {ErrReservedField, "id:uint"},
		"reserved created_at": {ErrReservedField, "created_at:datetime"},
		"reserved updated_at": {ErrReservedField, "updated_at:datetime"},
		"reserved deleted_at": {ErrReservedField, "deleted_at:datetime"},
		"reserved in pascal":  {ErrReservedField, "CreatedAt:datetime"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			field, err := ParseField(test.spec)

			require.ErrorIs(t, err, test.expected)
			require.Equal(t, Field{}, field, "a rejected spec must not yield a half built field")
		})
	}
}

func TestBindingTag(t *testing.T) {
	tests := map[string]struct {
		spec     string
		expected string
	}{
		"optional string":   {"name:string", ""},
		"required string":   {"name:string:required", "required"},
		"optional text":     {"notes:text", ""},
		"optional int":      {"stock:int", ""},
		"required int":      {"stock:int:required", "required"},
		"optional uint":     {"units:uint", "gte=0"},
		"required uint":     {"units:uint:required", "required"},
		"optional bool":     {"active:bool", ""},
		"optional money":    {"price:money", "omitempty,decimal2"},
		"required money":    {"price:money:required", "required,decimal2"},
		"optional email":    {"contact:email", "omitempty,email"},
		"required email":    {"contact:email:required", "required,email"},
		"optional ref":      {"category_id:ref=item_category", ""},
		"required ref":      {"category_id:ref=item_category:required", "required"},
		"optional date":     {"released_at:date", ""},
		"required datetime": {"starts_at:datetime:required", "required"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			field, err := ParseField(test.spec)

			require.NoError(t, err)
			require.Equal(t, test.expected, field.BindingTag())
		})
	}
}

// TestBindingTagIsEmptyForTheSyntheticPrimaryKey: the id is not submitted by
// the form, so a binding rule on it would reject every create.
func TestBindingTagIsEmptyForTheSyntheticPrimaryKey(t *testing.T) {
	id := idField()
	require.True(t, id.IsPrimaryKey)
	require.Empty(t, id.BindingTag())

	// Even if something marks it required, a primary key carries no rule.
	id.Required = true
	require.Empty(t, id.BindingTag())
}

func TestFieldRenderingHelpers(t *testing.T) {
	tests := map[string]struct {
		spec       string
		align      string
		stringy    bool
		needsParse bool
		needsTime  bool
	}{
		"string":   {"name:string", "", true, false, false},
		"text":     {"notes:text", "", true, false, false},
		"email":    {"contact:email", "", true, false, false},
		"int":      {"stock:int", "right", false, false, false},
		"uint":     {"units:uint", "right", false, false, false},
		"money":    {"price:money", "right", false, true, false},
		"bool":     {"active:bool", "center", false, false, false},
		"date":     {"released_at:date", "", false, true, true},
		"datetime": {"starts_at:datetime", "", false, true, true},
		"ref":      {"category_id:ref=item_category", "", false, false, false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			field, err := ParseField(test.spec)

			require.NoError(t, err)
			require.Equal(t, test.align, field.TableAlign())
			require.Equal(t, test.stringy, field.IsStringy())
			require.Equal(t, test.needsParse, field.NeedsParse())
			require.Equal(t, test.needsTime, field.NeedsTime())
		})
	}
}
