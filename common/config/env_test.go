package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Drathveloper/go-web-skeleton/common/config"
	"github.com/Drathveloper/go-web-skeleton/common/config/model"
)

// clearMappedVariables removes every variable of the explicit mapping from
// the host environment, so the tests only see what they set themselves. A CI
// machine routinely defines PORT or ENVIRONMENT, which would otherwise
// override the defaults these tests assert. t.Setenv records the original
// value for restoration (which also rules out t.Parallel); the Unsetenv makes
// the variable truly absent rather than empty.
func clearMappedVariables(t *testing.T) {
	t.Helper()

	keys := []string{
		"GIN_MODE", "PORT", "ENVIRONMENT", "SERVICE_NAME", "TLS_CERT_FILE",
		"TLS_KEY_FILE", "SEED_ADMIN_USERNAME", "SEED_ADMIN_PASSWORD", "ENABLE_TLS",
	}
	for _, key := range keys {
		t.Setenv(key, "")
		require.NoError(t, os.Unsetenv(key))
	}
}

func TestLoadEnv(t *testing.T) {
	clearMappedVariables(t)

	envConfig, err := config.LoadEnv(model.BuildInfo{
		Commit:    "someCommit",
		BuildTime: "someBuildTime",
		Version:   "someVersion",
	})

	require.NoError(t, err)
	require.Equal(t, "someCommit", envConfig.Commit)
	require.Equal(t, "someBuildTime", envConfig.BuildTime)
	require.Equal(t, "someVersion", envConfig.Version)
	require.Equal(t, "8000", envConfig.Port)
	require.Equal(t, "release", envConfig.GinMode)
	require.Equal(t, "dev", envConfig.Environment)
	require.Equal(t, "go-web-skeleton", envConfig.ServiceName)
	require.False(t, envConfig.EnableTLS)
}

func TestLoadEnv_ExplicitMappingOverridesDefaults(t *testing.T) {
	clearMappedVariables(t)
	t.Setenv("PORT", "9999")
	t.Setenv("GIN_MODE", "debug")
	t.Setenv("SERVICE_NAME", "custom-service")
	t.Setenv("ENVIRONMENT", "prod")
	t.Setenv("ENABLE_TLS", "true")
	t.Setenv("TLS_CERT_FILE", "/tmp/cert.pem")
	t.Setenv("TLS_KEY_FILE", "/tmp/key.pem")
	t.Setenv("SEED_ADMIN_USERNAME", "admin")
	t.Setenv("SEED_ADMIN_PASSWORD", "someSecret")

	envConfig, err := config.LoadEnv(model.BuildInfo{})

	require.NoError(t, err)
	require.Equal(t, "9999", envConfig.Port)
	require.Equal(t, "debug", envConfig.GinMode)
	require.Equal(t, "custom-service", envConfig.ServiceName)
	require.Equal(t, "prod", envConfig.Environment)
	require.True(t, envConfig.EnableTLS)
	require.Equal(t, "/tmp/cert.pem", envConfig.TLSCertFilePath)
	require.Equal(t, "/tmp/key.pem", envConfig.TLSKeyFilePath)
	require.Equal(t, "admin", envConfig.SeedAdminUsername)
	require.Equal(t, "someSecret", envConfig.SeedAdminPassword)
}

// The mapping is a closed list: anything else living in the process
// environment (PATH, HOME, or an unrelated variable a deployment happens to
// set) must be discarded before the strict decoder sees it, or every foreign
// variable would abort startup.
func TestLoadEnv_UnmappedVariablesAreDiscarded(t *testing.T) {
	clearMappedVariables(t)
	t.Setenv("SOME_UNRELATED_VARIABLE", "whatever")

	envConfig, err := config.LoadEnv(model.BuildInfo{})

	require.NoError(t, err)
	require.Equal(t, "8000", envConfig.Port)
	require.Equal(t, "go-web-skeleton", envConfig.ServiceName)
}

func TestLoadEnv_InvalidValueIsAHardError(t *testing.T) {
	clearMappedVariables(t)
	t.Setenv("ENABLE_TLS", "notabool")

	envConfig, err := config.LoadEnv(model.BuildInfo{})

	require.Nil(t, envConfig)
	require.Error(t, err)
	require.Equal(t,
		"load env config failed: decoding failed due to the following error(s):\n\n"+
			"'enable_tls' strconv.ParseBool: invalid syntax",
		err.Error())
}

// A template that shipped a default administrator password would put the same
// one in every project generated from it, and in every deployment that forgot
// to change it. There must be no default.
func TestLoadEnv_SeedAdminCredentialsHaveNoDefault(t *testing.T) {
	clearMappedVariables(t)

	envConfig, err := config.LoadEnv(model.BuildInfo{})

	require.NoError(t, err)
	require.Empty(t, envConfig.SeedAdminUsername)
	require.Empty(t, envConfig.SeedAdminPassword)
}
