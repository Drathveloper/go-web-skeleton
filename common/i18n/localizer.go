package i18n

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var localizers = make(map[string]*i18n.Localizer) //nolint:gochecknoglobals

func LocalizeMessage(locale, messageID string) string {
	localizer, ok := localizers[locale]
	if !ok {
		return messageID
	}
	message := &i18n.Message{
		ID: messageID,
	}
	localizedMsg, err := localizer.LocalizeMessage(message)
	if err != nil {
		return messageID
	}
	return localizedMsg
}
