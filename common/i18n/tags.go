package i18n

import (
	"sort"

	"golang.org/x/text/language"
)

// DefaultLanguage is the fallback used when no locale can be resolved.
const DefaultLanguage = "en"

func ParseAcceptLanguage(value string) string {
	tags, _, err := language.ParseAcceptLanguage(value)
	if err != nil {
		return DefaultLanguage
	}
	for _, tag := range tags {
		b, _ := tag.Base()
		if localizers[b.String()] != nil {
			return b.String()
		}
	}
	return DefaultLanguage
}

// AvailableLanguages lists the locales the bundle actually loaded, sorted, so a
// language switcher can only ever offer one that resolves. Deriving it from the
// registered localizers means adding a <module>.<lang>.json is all it takes.
func AvailableLanguages() []string {
	languages := make([]string, 0, len(localizers))
	for lang := range localizers {
		languages = append(languages, lang)
	}
	sort.Strings(languages)
	return languages
}

// IsAvailableLanguage reports whether a locale was loaded.
func IsAvailableLanguage(lang string) bool {
	_, ok := localizers[lang]
	return ok
}
