package config

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/yaml"
	envprovider "github.com/knadh/koanf/providers/env/v2"
	fsprovider "github.com/knadh/koanf/providers/fs"
	"github.com/knadh/koanf/v2"

	"github.com/Drathveloper/go-web-skeleton/common/config/model"
	"github.com/Drathveloper/go-web-skeleton/common/constants"
)

const (
	readConfigBaseErrMsg = "read config failed"
	loadConfigBaseErrMsg = "load config failed"
)

const (
	yamlConfigFilePath = "config/application.yaml"
	jsonConfigFilePath = "config/application.json"
)

// EnvOverridePrefix and EnvOverrideNestingDelim define the environment
// override syntax: APP_SERVER__READ_TIMEOUT=30s becomes server.read_timeout.
// The nesting separator is a double underscore because the configuration keys
// themselves contain single ones.
const (
	EnvOverridePrefix       = "APP_"
	EnvOverrideNestingDelim = "__"
)

const koanfDelim = "."

var errConfigFileNotFound = errors.New("configuration file not found")

// ReadConfig loads the application configuration: the file first, then any
// APP_-prefixed environment variables layered on top of it. The environment
// only overrides, it never replaces the file — without a file there is no
// startup.
func ReadConfig(fsys fs.FS) (*model.Configuration, error) {
	engine := koanf.New(koanfDelim)
	if err := loadConfigFile(engine, fsys); err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, readConfigBaseErrMsg, err)
	}
	overrides := envprovider.Provider(koanfDelim, envprovider.Opt{
		Prefix:        EnvOverridePrefix,
		TransformFunc: transformOverrideVariable,
	})
	if err := engine.Load(overrides, nil); err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, readConfigBaseErrMsg, err)
	}
	var configuration model.Configuration
	unmarshalConf := koanf.UnmarshalConf{DecoderConfig: strictDecoderConfig()}
	if err := engine.UnmarshalWithConf("", &configuration, unmarshalConf); err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, readConfigBaseErrMsg, err)
	}
	return &configuration, nil
}

// loadConfigFile reads config/application.yaml, falling back to
// config/application.json only when the YAML file does not exist: a malformed
// YAML file must fail here, or a syntax error would silently switch the
// deployment to whatever stale application.json sits next to it.
func loadConfigFile(engine *koanf.Koanf, fsys fs.FS) error {
	yamlErr := engine.Load(fsprovider.Provider(fsys, yamlConfigFilePath), yaml.Parser())
	if yamlErr == nil {
		return nil
	}
	if !errors.Is(yamlErr, fs.ErrNotExist) {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, loadConfigBaseErrMsg, yamlErr)
	}
	jsonErr := engine.Load(fsprovider.Provider(fsys, jsonConfigFilePath), json.Parser())
	if jsonErr == nil {
		return nil
	}
	if !errors.Is(jsonErr, fs.ErrNotExist) {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, loadConfigBaseErrMsg, jsonErr)
	}
	return fmt.Errorf("%s: %w: tried %s and %s",
		loadConfigBaseErrMsg, errConfigFileNotFound, yamlConfigFilePath, jsonConfigFilePath)
}

// transformOverrideVariable maps APP_SERVER__READ_TIMEOUT to
// server.read_timeout: strip the prefix, lowercase, and turn the double
// underscore into the nesting separator. There is no allowlist here on
// purpose — a misspelled APP_ variable maps to a key the model does not
// have, and the strict decoder aborts startup naming it, instead of the
// override being silently ignored.
func transformOverrideVariable(key, value string) (string, any) {
	key = strings.TrimPrefix(key, EnvOverridePrefix)
	key = strings.ToLower(key)
	key = strings.ReplaceAll(key, EnvOverrideNestingDelim, koanfDelim)
	return key, value
}

// strictDecoderConfig is shared by the file and environment loaders.
//
// ErrorUnused makes an unrecognised key a hard error. Non-strict decoding is
// how a misspelled `max_connection_lifetime` sat in the source's production
// config being silently ignored: the pool never got the lifetime it was
// configured with and nothing said so.
//
// Passing a custom DecoderConfig discards the hooks koanf would otherwise
// install, so the useful ones are recomposed here. StringToBasicTypeHookFunc
// exists because environment values always arrive as strings: without it an
// override like APP_SERVER__MAX_HEADER_BYTES=4096 could not decode into an
// int field.
func strictDecoderConfig() *mapstructure.DecoderConfig {
	return &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			mapstructure.TextUnmarshallerHookFunc(),
			mapstructure.StringToBasicTypeHookFunc(),
		),
		ErrorUnused: true,
	}
}
