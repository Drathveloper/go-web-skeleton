package main

import (
	"strings"
	"unicode"
)

// Names holds every spelling of an entity name the generator needs. They are
// derived once, from the snake_case name the user typed, so a template never
// has to do string surgery and two templates can never disagree about how a
// name is spelled.
type Names struct {
	Snake        string // item_category
	SnakePlural  string // item_categories
	Pascal       string // ItemCategory
	PascalPlural string // ItemCategories
	Camel        string // itemCategory
	CamelPlural  string // itemCategories
	Kebab        string // item-category
	KebabPlural  string // item-categories
	Words        string // item category
	WordsPlural  string // item categories
	Human        string // Item category
	HumanPlural  string // Item categories
	Receiver     string // i
}

// NewNames derives the spellings. plural may be empty, in which case it is
// guessed; the --plural flag exists because English is not derivable.
func NewNames(name, plural string) Names {
	snake := toSnake(name)
	if plural == "" {
		plural = pluralize(snake)
	}
	snakePlural := toSnake(plural)

	return Names{
		Snake:        snake,
		SnakePlural:  snakePlural,
		Pascal:       toPascal(snake),
		PascalPlural: toPascal(snakePlural),
		Camel:        toCamel(snake),
		CamelPlural:  toCamel(snakePlural),
		Kebab:        strings.ReplaceAll(snake, "_", "-"),
		KebabPlural:  strings.ReplaceAll(snakePlural, "_", "-"),
		Words:        strings.ReplaceAll(snake, "_", " "),
		WordsPlural:  strings.ReplaceAll(snakePlural, "_", " "),
		Human:        capitalizeFirst(strings.ReplaceAll(snake, "_", " ")),
		HumanPlural:  capitalizeFirst(strings.ReplaceAll(snakePlural, "_", " ")),
		Receiver:     receiverOf(snake),
	}
}

// toSnake accepts whatever the user typed — ItemCategory, itemCategory,
// item-category, item category — and normalises it.
func toSnake(value string) string {
	var builder strings.Builder
	runes := []rune(strings.TrimSpace(value))
	for index, char := range runes {
		switch {
		case char == '-' || char == ' ' || char == '.':
			builder.WriteRune('_')
		case unicode.IsUpper(char):
			// Only break before an uppercase run that starts a new word, so
			// "HTTPServer" becomes http_server rather than h_t_t_p_server.
			if index > 0 && (unicode.IsLower(runes[index-1]) ||
				(index+1 < len(runes) && unicode.IsLower(runes[index+1]))) {
				builder.WriteRune('_')
			}
			builder.WriteRune(unicode.ToLower(char))
		default:
			builder.WriteRune(char)
		}
	}
	return strings.Trim(collapseUnderscores(builder.String()), "_")
}

func collapseUnderscores(value string) string {
	for strings.Contains(value, "__") {
		value = strings.ReplaceAll(value, "__", "_")
	}
	return value
}

func toPascal(snake string) string {
	parts := strings.Split(snake, "_")
	var builder strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		if initialism, ok := commonInitialisms[part]; ok {
			builder.WriteString(initialism)
			continue
		}
		builder.WriteString(capitalizeFirst(part))
	}
	return builder.String()
}

func toCamel(snake string) string {
	pascal := toPascal(snake)
	if pascal == "" {
		return ""
	}
	// An initialism that starts the name lowercases whole: ID -> id, not iD.
	first := strings.Split(snake, "_")[0]
	if initialism, ok := commonInitialisms[first]; ok {
		return strings.ToLower(initialism) + pascal[len(initialism):]
	}
	runes := []rune(pascal)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func capitalizeFirst(value string) string {
	if value == "" {
		return ""
	}
	runes := []rune(value)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// commonInitialisms keeps generated identifiers idiomatic. Without it a field
// named customer_id would produce CustomerId, which every Go linter flags.
var commonInitialisms = map[string]string{ //nolint:gochecknoglobals
	"id": "ID", "url": "URL", "uri": "URI", "api": "API", "http": "HTTP",
	"https": "HTTPS", "html": "HTML", "json": "JSON", "sql": "SQL", "db": "DB",
	"uuid": "UUID", "ip": "IP", "csv": "CSV", "pdf": "PDF", "vat": "VAT",
}

// pluralize covers the regular English cases and is deliberately not clever:
// anything irregular is what --plural is for.
func pluralize(snake string) string {
	parts := strings.Split(snake, "_")
	last := parts[len(parts)-1]
	parts[len(parts)-1] = pluralizeWord(last)
	return strings.Join(parts, "_")
}

func pluralizeWord(word string) string {
	switch {
	case word == "":
		return word
	// The length guard is not hypothetical: a one letter name ending in y
	// would index before the start of the string and panic.
	case len(word) > 1 && strings.HasSuffix(word, "y") && !isVowel(word[len(word)-2]):
		return word[:len(word)-1] + "ies"
	case strings.HasSuffix(word, "s"), strings.HasSuffix(word, "x"),
		strings.HasSuffix(word, "z"), strings.HasSuffix(word, "ch"),
		strings.HasSuffix(word, "sh"):
		return word + "es"
	default:
		return word + "s"
	}
}

func isVowel(char byte) bool {
	return strings.IndexByte("aeiou", char) >= 0
}

// receiverOf is the single-letter method receiver. Go style wants it short and
// derived from the type, and every hand-written module in this project uses the
// first letter of the entity.
func receiverOf(snake string) string {
	if snake == "" {
		return "e"
	}
	return string([]rune(snake)[0])
}
