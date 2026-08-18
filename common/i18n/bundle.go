package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

const bundleExtension = ".json"

// LocaleFS holds one catalog per module, named <module>.<lang>.json. go-i18n
// derives the language from the file name, so adding a module only means
// dropping its two files here.
//
//go:embed files/*.json
var LocaleFS embed.FS

func InitializeI18n() error {
	localePaths := make([]string, 0)
	if err := fs.WalkDir(LocaleFS, ".", appendI18nPaths(&localePaths)); err != nil {
		return fmt.Errorf("initialize i18n bundle failed: %w", err)
	}
	bundle, err := initializeBundle(localePaths...)
	if err != nil {
		return fmt.Errorf("initialize i18n bundle failed: %w", err)
	}
	for _, tag := range bundle.LanguageTags() {
		localizers[tag.String()] = i18n.NewLocalizer(bundle, tag.String())
	}
	return nil
}

func appendI18nPaths(localePaths *[]string) fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != bundleExtension {
			return nil
		}
		*localePaths = append(*localePaths, path)
		return nil
	}
}

func initializeBundle(localePaths ...string) (*i18n.Bundle, error) {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	for _, path := range localePaths {
		if _, err := bundle.LoadMessageFileFS(LocaleFS, path); err != nil {
			return nil, fmt.Errorf("initialize i18n bundle failed: %w", err)
		}
	}

	return bundle, nil
}
