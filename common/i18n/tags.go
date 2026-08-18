package i18n

import (
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
