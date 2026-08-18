package config

import (
	"fmt"

	"github.com/knadh/koanf/providers/confmap"
	envprovider "github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/v2"

	"github.com/Drathveloper/go-web-skeleton/common/config/model"
	"github.com/Drathveloper/go-web-skeleton/common/constants"
)

const loadEnvErrMsg = "load env config failed"

// LoadEnv reads the process-level settings from the environment and returns
// them; there is deliberately no cached global, the instance travels through
// the wire container like every other dependency.
func LoadEnv(buildInfo model.BuildInfo) (*model.EnvConfig, error) {
	// SEED_ADMIN_USERNAME and SEED_ADMIN_PASSWORD deliberately have no
	// default: a template that shipped a known password would put one in
	// every project generated from it, and every deployment that forgot to
	// change it.
	defaults := map[string]any{
		"gin_mode":     "release",
		"port":         "8000",
		"environment":  "dev",
		"service_name": "go-web-skeleton",
		"enable_tls":   false,
	}
	// The variables are mapped explicitly and without a prefix so the current
	// deployment contract (PORT, GIN_MODE, ...) stays intact. Everything else
	// in the process environment is discarded instead of decoded: PATH or
	// HOME must never trip the strict decoder or leak into the config.
	knownVariables := map[string]string{
		"GIN_MODE":            "gin_mode",
		"PORT":                "port",
		"ENVIRONMENT":         "environment",
		"SERVICE_NAME":        "service_name",
		"TLS_CERT_FILE":       "tls_cert_file",
		"TLS_KEY_FILE":        "tls_key_file",
		"SEED_ADMIN_USERNAME": "seed_admin_username",
		"SEED_ADMIN_PASSWORD": "seed_admin_password",
		"ENABLE_TLS":          "enable_tls",
	}
	engine := koanf.New(koanfDelim)
	if err := engine.Load(confmap.Provider(defaults, koanfDelim), nil); err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, loadEnvErrMsg, err)
	}
	variables := envprovider.Provider(koanfDelim, envprovider.Opt{
		// Returning an empty key tells the provider to drop the variable.
		TransformFunc: func(key, value string) (string, any) {
			return knownVariables[key], value
		},
	})
	if err := engine.Load(variables, nil); err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, loadEnvErrMsg, err)
	}
	envConfig := &model.EnvConfig{}
	unmarshalConf := koanf.UnmarshalConf{DecoderConfig: strictDecoderConfig()}
	if err := engine.UnmarshalWithConf("", envConfig, unmarshalConf); err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, loadEnvErrMsg, err)
	}
	envConfig.BuildInfo = buildInfo
	return envConfig, nil
}
