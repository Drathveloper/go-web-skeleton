package main

// Naming is where a generator regression is cheapest to make and most
// expensive to notice: every spelling of the entity is derived here, and a
// wrong one lands in hundreds of generated lines that still compile. The cases
// below are the ones that produce wrong Go rather than merely ugly Go.

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToSnake(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected string
	}{
		"already snake":            {"item_category", "item_category"},
		"pascal":                   {"ItemCategory", "item_category"},
		"camel":                    {"itemCategory", "item_category"},
		"kebab":                    {"item-category", "item_category"},
		"words":                    {"item category", "item_category"},
		"dotted":                   {"item.category", "item_category"},
		"surrounding whitespace":   {"  Item  Category  ", "item_category"},
		"single word":              {"item", "item"},
		"trailing initialism":      {"UserID", "user_id"},
		"leading initialism run":   {"HTTPServer", "http_server"},
		"initialism mid name":      {"APIKeyValue", "api_key_value"},
		"all caps":                 {"HTTP", "http"},
		"already snake with digit": {"address_line_2", "address_line_2"},
		"empty":                    {"", ""},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, test.expected, toSnake(test.input))
		})
	}
}

// TestToPascalKeepsInitialismsIdiomatic pins the case the linter would reject:
// CustomerId is what a naive Title() produces and what golangci-lint flags in
// every generated file at once.
func TestToPascalKeepsInitialismsIdiomatic(t *testing.T) {
	tests := map[string]string{
		"customer_id":   "CustomerID",
		"id":            "ID",
		"item_category": "ItemCategory",
		"http_server":   "HTTPServer",
		"api_url":       "APIURL",
		"invoice_pdf":   "InvoicePDF",
		"":              "",
	}

	for input, expected := range tests {
		t.Run(input, func(t *testing.T) {
			require.Equal(t, expected, toPascal(input))
		})
	}
}

// TestToCamelLowercasesALeadingInitialismWhole covers the other half of the
// same rule: iD is not a Go identifier anybody writes, id is.
func TestToCamelLowercasesALeadingInitialismWhole(t *testing.T) {
	tests := map[string]string{
		"id":            "id",
		"id_value":      "idValue",
		"customer_id":   "customerID",
		"http_server":   "httpServer",
		"item_category": "itemCategory",
		"item":          "item",
		"":              "",
	}

	for input, expected := range tests {
		t.Run(input, func(t *testing.T) {
			require.Equal(t, expected, toCamel(input))
		})
	}
}

func TestPluralize(t *testing.T) {
	tests := map[string]string{
		"item":          "items",
		"category":      "categories",
		"box":           "boxes",
		"day":           "days",
		"key":           "keys",
		"address":       "addresses",
		"batch":         "batches",
		"dish":          "dishes",
		"item_category": "item_categories",
		"invoice_line":  "invoice_lines",
	}

	for input, expected := range tests {
		t.Run(input, func(t *testing.T) {
			require.Equal(t, expected, pluralize(input))
		})
	}
}

func TestNewNames(t *testing.T) {
	names := NewNames("ItemCategory", "")

	require.Equal(t, Names{
		Snake:        "item_category",
		SnakePlural:  "item_categories",
		Pascal:       "ItemCategory",
		PascalPlural: "ItemCategories",
		Camel:        "itemCategory",
		CamelPlural:  "itemCategories",
		Kebab:        "item-category",
		KebabPlural:  "item-categories",
		Words:        "item category",
		WordsPlural:  "item categories",
		Human:        "Item category",
		HumanPlural:  "Item categories",
		Receiver:     "i",
	}, names)
}

func TestNewNamesDerivesIdiomaticGoForForeignKeys(t *testing.T) {
	names := NewNames("customer_id", "")

	require.Equal(t, "CustomerID", names.Pascal)
	require.Equal(t, "customerID", names.Camel)
	require.Equal(t, "customer-id", names.Kebab)
	require.Equal(t, "c", names.Receiver)
}

// TestNewNamesPluralOverridesTheGuess is why --plural exists: English plurals
// are not derivable, and the guess is deliberately not clever.
func TestNewNamesPluralOverridesTheGuess(t *testing.T) {
	guessed := NewNames("person", "")
	require.Equal(t, "persons", guessed.SnakePlural)

	overridden := NewNames("person", "people")
	require.Equal(t, "people", overridden.SnakePlural)
	require.Equal(t, "People", overridden.PascalPlural)
	require.Equal(t, "people", overridden.CamelPlural)
	require.Equal(t, "people", overridden.KebabPlural)
	require.Equal(t, "People", overridden.HumanPlural)
	// The singular spellings are untouched by --plural.
	require.Equal(t, "Person", overridden.Pascal)
}

// TestNewNamesNormalisesThePluralFlagToo: --plural ItemCategories has to end up
// as item_categories, not as a literal that leaks into file names.
func TestNewNamesNormalisesThePluralFlag(t *testing.T) {
	names := NewNames("item_category", "ItemCategories")

	require.Equal(t, "item_categories", names.SnakePlural)
	require.Equal(t, "ItemCategories", names.PascalPlural)
	require.Equal(t, "item-categories", names.KebabPlural)
}

func TestReceiverOf(t *testing.T) {
	require.Equal(t, "i", receiverOf("item"))
	require.Equal(t, "u", receiverOf("user_role"))
	// An empty name would otherwise index into an empty slice.
	require.Equal(t, "e", receiverOf(""))
}
