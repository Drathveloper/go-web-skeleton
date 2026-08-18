package i18n

import (
	"golang.org/x/text/language"
)

const defaultLanguage = "en"

func ParseAcceptLanguage(value string) string {
	tags, _, err := language.ParseAcceptLanguage(value)
	if err != nil {
		return defaultLanguage
	}
	for _, tag := range tags {
		b, _ := tag.Base()
		if localizers[b.String()] != nil {
			return b.String()
		}
	}
	return defaultLanguage
}
