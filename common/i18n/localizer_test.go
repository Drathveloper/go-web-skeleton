package i18n_test

import (
	"testing"

	i18nlib "github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/language"

	"github.com/Drathveloper/go-web-skeleton/common/i18n"
)

// Nothing in this file calls t.Parallel, on purpose. Every function under test
// reads the package-level catalog, and the tests that install a fixture catalog
// write it. Running them in parallel makes that a data race — which is exactly
// what the version this was rewritten from did, and why it passed without ever
// proving anything.

func newLocalizer(locale language.Tag, messages ...*i18nlib.Message) *i18nlib.Localizer {
	bundle := i18nlib.NewBundle(locale)
	_ = bundle.AddMessages(locale, messages...)

	return i18nlib.NewLocalizer(bundle, locale.String())
}

func TestLocalizeMessage(t *testing.T) {
	enLocalizer := newLocalizer(
		language.English,
		&i18nlib.Message{ID: "greeting", Other: "Hello!"},
		&i18nlib.Message{ID: "farewell", Other: "Goodbye!"},
	)
	esLocalizer := newLocalizer(
		language.Spanish,
		&i18nlib.Message{ID: "greeting", Other: "¡Hola!"},
	)

	tests := []struct {
		localizers map[string]*i18nlib.Localizer
		name       string
		locale     string
		messageID  string
		expected   string
	}{
		{
			name:       "test localize message should resolve an existing message in english",
			localizers: map[string]*i18nlib.Localizer{"en": enLocalizer},
			locale:     "en",
			messageID:  "greeting",
			expected:   "Hello!",
		},
		{
			name:       "test localize message should resolve an existing message in spanish",
			localizers: map[string]*i18nlib.Localizer{"es": esLocalizer},
			locale:     "es",
			messageID:  "greeting",
			expected:   "¡Hola!",
		},
		{
			name:       "test localize message should return the message id for an unregistered locale",
			localizers: map[string]*i18nlib.Localizer{"en": enLocalizer},
			locale:     "fr",
			messageID:  "greeting",
			expected:   "greeting",
		},
		{
			name:       "test localize message should return the message id when no locale is registered",
			localizers: map[string]*i18nlib.Localizer{},
			locale:     "en",
			messageID:  "greeting",
			expected:   "greeting",
		},
		{
			name:       "test localize message should return the message id when the key does not exist",
			localizers: map[string]*i18nlib.Localizer{"en": enLocalizer},
			locale:     "en",
			messageID:  "nonexistent_key",
			expected:   "nonexistent_key",
		},
		{
			name:       "test localize message should pick the localizer of the requested locale",
			localizers: map[string]*i18nlib.Localizer{"en": enLocalizer, "es": esLocalizer},
			locale:     "es",
			messageID:  "greeting",
			expected:   "¡Hola!",
		},
		{
			name:       "test localize message should return an empty string for an empty message id",
			localizers: map[string]*i18nlib.Localizer{"en": enLocalizer},
			locale:     "en",
			messageID:  "",
			expected:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i18n.SetLocalizers(t, tt.localizers)

			result := i18n.LocalizeMessage(tt.locale, tt.messageID)

			require.Equal(t, tt.expected, result)
		})
	}
}

func TestParseAcceptLanguage(t *testing.T) {
	i18n.SetLocalizers(t, map[string]*i18nlib.Localizer{
		"en": newLocalizer(language.English),
		"es": newLocalizer(language.Spanish),
	})

	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{
			name:     "test parse accept language should pick the first available language",
			header:   "es-ES,es;q=0.9,en;q=0.8",
			expected: "es",
		},
		{
			name:     "test parse accept language should skip languages with no catalog",
			header:   "fr-FR,fr;q=0.9,es;q=0.8",
			expected: "es",
		},
		{
			name:     "test parse accept language should fall back to the default when none is available",
			header:   "fr-FR,de;q=0.8",
			expected: i18n.DefaultLanguage,
		},
		{
			name:     "test parse accept language should fall back to the default for an empty header",
			header:   "",
			expected: i18n.DefaultLanguage,
		},
		{
			name:     "test parse accept language should fall back to the default for a malformed header",
			header:   "not a language tag at all;;;",
			expected: i18n.DefaultLanguage,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, i18n.ParseAcceptLanguage(tt.header))
		})
	}
}

// A language switcher may only offer locales that actually resolve: offering one
// with no catalog means every label on the next page falls back to its key.
func TestAvailableLanguages(t *testing.T) {
	i18n.SetLocalizers(t, map[string]*i18nlib.Localizer{
		"es": newLocalizer(language.Spanish),
		"en": newLocalizer(language.English),
		"fr": newLocalizer(language.French),
	})

	require.Equal(t, []string{"en", "es", "fr"}, i18n.AvailableLanguages(), "must be sorted and complete")
	require.True(t, i18n.IsAvailableLanguage("es"))
	require.False(t, i18n.IsAvailableLanguage("de"))
}

func TestAvailableLanguages_EmptyCatalog(t *testing.T) {
	i18n.SetLocalizers(t, map[string]*i18nlib.Localizer{})

	require.Empty(t, i18n.AvailableLanguages())
	require.False(t, i18n.IsAvailableLanguage(i18n.DefaultLanguage))
}

// The embedded catalogs are the ones the application ships: they have to load,
// register exactly the two locales the skeleton documents, and resolve a key.
func TestInitializeI18n(t *testing.T) {
	i18n.SetLocalizers(t, map[string]*i18nlib.Localizer{})

	require.NoError(t, i18n.InitializeI18n())

	require.Equal(t, []string{"en", "es"}, i18n.AvailableLanguages())
	require.Equal(t, "Error", i18n.LocalizeMessage("en", "errors.title"))
	require.Equal(t, "Error interno del servidor.", i18n.LocalizeMessage("es", "errors.internal"))
	require.Equal(t, "en", i18n.ParseAcceptLanguage("es-ES;q=0.1,en;q=0.9"))
}
