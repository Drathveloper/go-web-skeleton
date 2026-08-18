package config_test

import (
	"os"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Drathveloper/go-web-skeleton/common/config"
)

const (
	yamlConfigPath = "config/application.yaml"
	jsonConfigPath = "config/application.json"
)

const baseYAMLContent = `server:
  read_timeout: 10s
  max_header_bytes: 2048
logging:
  level: DEBUG
  confidential_fields:
    - password
security:
  session:
    ttl: 3600
    secure_cookie: true
databases:
  postgres:
    host: localhost
    port: 5432
    database: skeleton
    user: skeleton
    password: skeleton
    ssl:
      mode: disable
    pool:
      max_connections: 5
      max_lifetime: 30m
  redis:
    hosts:
      - localhost:6379
`

const baseJSONContent = `{
  "server": {"read_timeout": "10s", "max_header_bytes": 2048},
  "logging": {"level": "ERROR"},
  "security": {"session": {"ttl": 3600, "secure_cookie": true}},
  "databases": {
    "postgres": {
      "host": "localhost",
      "port": 5432,
      "database": "skeleton",
      "user": "skeleton",
      "password": "skeleton",
      "ssl": {"mode": "disable"},
      "pool": {"max_connections": 5, "max_lifetime": "30m"}
    },
    "redis": {"hosts": ["localhost:6379"]}
  }
}`

// clearOverrideVariables removes any APP_-prefixed variable leaking from the
// host environment, so the tests only see what they set themselves. Without
// this, a stray APP_ variable on a developer or CI machine would override the
// fixture values or trip the strict decoder in every test below. t.Setenv
// records the original value for restoration (which also rules out
// t.Parallel); the Unsetenv makes the variable truly absent rather than
// empty, because an empty override still reaches the decoder.
func clearOverrideVariables(t *testing.T) {
	t.Helper()

	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, config.EnvOverridePrefix) {
			continue
		}
		key := strings.SplitN(entry, "=", 2)[0]
		t.Setenv(key, "")
		require.NoError(t, os.Unsetenv(key))
	}
}

func yamlConfigFS(t *testing.T, content string) fstest.MapFS {
	t.Helper()

	return fstest.MapFS{
		yamlConfigPath: &fstest.MapFile{Data: []byte(content)},
	}
}

func jsonConfigFS(t *testing.T, content string) fstest.MapFS {
	t.Helper()

	return fstest.MapFS{
		jsonConfigPath: &fstest.MapFile{Data: []byte(content)},
	}
}

func TestReadConfig(t *testing.T) {
	clearOverrideVariables(t)

	configuration, err := config.ReadConfig(yamlConfigFS(t, baseYAMLContent))

	require.NoError(t, err)
	require.Equal(t, 10*time.Second, configuration.Server.ReadTimeout)
	require.Equal(t, "DEBUG", configuration.Logging.Level)
	require.Equal(t, int64(3600), *configuration.Security.Session.TTL)
	require.True(t, configuration.Security.Session.SecureCookie)
	require.Equal(t, "localhost", configuration.Databases.Postgres.Host)
	require.Equal(t, int32(5), configuration.Databases.Postgres.Pool.MaxConnections)
	require.Equal(t, 30*time.Minute, configuration.Databases.Postgres.Pool.MaxConnectionsLifetime)
	require.Equal(t, []string{"localhost:6379"}, configuration.Databases.Redis.Hosts)
}

// JSON is only a fallback: it must carry the exact same model, durations
// included, so a deployment that renders application.json instead of the YAML
// behaves identically.
func TestReadConfig_JSONFallbackWhenYAMLIsMissing(t *testing.T) {
	clearOverrideVariables(t)

	configuration, err := config.ReadConfig(jsonConfigFS(t, baseJSONContent))

	require.NoError(t, err)
	require.Equal(t, 10*time.Second, configuration.Server.ReadTimeout)
	require.Equal(t, "ERROR", configuration.Logging.Level)
	require.Equal(t, int64(3600), *configuration.Security.Session.TTL)
	require.Equal(t, 2048, configuration.Server.MaxHeaderBytes)
	require.Equal(t, 30*time.Minute, configuration.Databases.Postgres.Pool.MaxConnectionsLifetime)
	require.Equal(t, []string{"localhost:6379"}, configuration.Databases.Redis.Hosts)
}

func TestReadConfig_YAMLWinsWhenBothFilesArePresent(t *testing.T) {
	clearOverrideVariables(t)

	fsys := fstest.MapFS{
		yamlConfigPath: &fstest.MapFile{Data: []byte(baseYAMLContent)},
		jsonConfigPath: &fstest.MapFile{Data: []byte(baseJSONContent)},
	}

	configuration, err := config.ReadConfig(fsys)

	require.NoError(t, err)
	// The JSON file says ERROR; it must not even have been read.
	require.Equal(t, "DEBUG", configuration.Logging.Level)
}

// The regression this test exists for: with non-strict decoding, a key the model
// does not have was silently dropped. `max_connection_lifetime` (the model calls
// it `max_lifetime`) sat in the source's production config doing nothing — the
// pool never got the lifetime it was configured with, and nothing said so.
// ErrorUnused makes it a startup failure that names the offending key, in both
// file formats.
func TestReadConfig_UnknownKeyIsAHardError(t *testing.T) {
	clearOverrideVariables(t)

	yamlContent := `databases:
  postgres:
    host: localhost
    pool:
      max_connection_lifetime: 30m
`
	jsonContent := `{"databases":{"postgres":{"host":"localhost","pool":{"max_connection_lifetime":"30m"}}}}`

	tests := []struct {
		fs   fstest.MapFS
		name string
	}{
		{
			name: "test read config should fail when the yaml has an unknown key",
			fs:   yamlConfigFS(t, yamlContent),
		},
		{
			name: "test read config should fail when the json has an unknown key",
			fs:   jsonConfigFS(t, jsonContent),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configuration, err := config.ReadConfig(tt.fs)

			require.Nil(t, configuration)
			require.Error(t, err)
			require.Equal(t,
				"read config failed: decoding failed due to the following error(s):\n\n"+
					"'databases.postgres.pool' has invalid keys: max_connection_lifetime",
				err.Error())
		})
	}
}

func TestReadConfig_Failures(t *testing.T) {
	clearOverrideVariables(t)

	tests := []struct {
		fs             fstest.MapFS
		name           string
		expectedErrMsg string
	}{
		{
			name: "test read config should fail when both config files are missing",
			fs:   fstest.MapFS{},
			expectedErrMsg: "read config failed: load config failed: configuration file not found: " +
				"tried config/application.yaml and config/application.json",
		},
		{
			// A malformed YAML must fail, not silently switch the deployment
			// to whatever application.json sits next to it — hence the valid
			// JSON planted alongside.
			name: "test read config should fail when the yaml is malformed even if a valid json exists",
			fs: fstest.MapFS{
				yamlConfigPath: &fstest.MapFile{Data: []byte("databases: [1,2\n")},
				jsonConfigPath: &fstest.MapFile{Data: []byte(baseJSONContent)},
			},
			expectedErrMsg: "read config failed: load config failed: yaml: line 1: did not find expected ',' or ']'",
		},
		{
			name:           "test read config should fail when the json fallback is malformed",
			fs:             jsonConfigFS(t, "{oops"),
			expectedErrMsg: "read config failed: load config failed: invalid character 'o' looking for beginning of object key string",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configuration, err := config.ReadConfig(tt.fs)

			require.Nil(t, configuration)
			require.Error(t, err)
			require.Equal(t, tt.expectedErrMsg, err.Error())
		})
	}
}

func TestReadConfig_EnvironmentOverridesFile(t *testing.T) {
	clearOverrideVariables(t)
	t.Setenv("APP_LOGGING__LEVEL", "ERROR")
	t.Setenv("APP_LOGGING__CONFIDENTIAL_FIELDS", "password,token")
	t.Setenv("APP_SERVER__READ_TIMEOUT", "1s")
	t.Setenv("APP_SERVER__MAX_HEADER_BYTES", "4096")
	t.Setenv("APP_SECURITY__SESSION__SECURE_COOKIE", "false")

	configuration, err := config.ReadConfig(yamlConfigFS(t, baseYAMLContent))

	require.NoError(t, err)
	require.Equal(t, "ERROR", configuration.Logging.Level)
	require.Equal(t, []string{"password", "token"}, configuration.Logging.ConfidentialFields)
	require.Equal(t, 1*time.Second, configuration.Server.ReadTimeout)
	require.Equal(t, 4096, configuration.Server.MaxHeaderBytes)
	require.False(t, configuration.Security.Session.SecureCookie)
	// Values without an override keep what the file says.
	require.Equal(t, "localhost", configuration.Databases.Postgres.Host)
	require.Equal(t, int64(3600), *configuration.Security.Session.TTL)
}

// A misspelled APP_ variable maps to a key the model does not have. It must
// abort startup naming the variable, for the same reason an unknown file key
// does: an override that is silently ignored is worse than no override.
func TestReadConfig_UnknownEnvironmentOverrideIsAHardError(t *testing.T) {
	clearOverrideVariables(t)
	t.Setenv("APP_SERVER__READ_TIMEOUTT", "5s")

	configuration, err := config.ReadConfig(yamlConfigFS(t, baseYAMLContent))

	require.Nil(t, configuration)
	require.Error(t, err)
	require.Equal(t,
		"read config failed: decoding failed due to the following error(s):\n\n"+
			"'server' has invalid keys: read_timeoutt",
		err.Error())
}

// The example files are what every new project copies to application.yaml or
// application.json. Since an unrecognised key aborts startup, an example that
// drifted from the model would break the first run of every generated project,
// and nothing else in the build would catch it.
func TestReadConfig_ExampleConfigsStillMatchTheModel(t *testing.T) {
	clearOverrideVariables(t)

	tests := []struct {
		name        string
		examplePath string
		mountPath   string
	}{
		{
			name:        "test yaml example should match the model",
			examplePath: "../../cmd/server/config/application.example.yaml",
			mountPath:   yamlConfigPath,
		},
		{
			name:        "test json example should match the model",
			examplePath: "../../cmd/server/config/application.example.json",
			mountPath:   jsonConfigPath,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := os.ReadFile(tt.examplePath)
			require.NoError(t, err)

			configuration, err := config.ReadConfig(fstest.MapFS{
				tt.mountPath: &fstest.MapFile{Data: content},
			})

			require.NoError(t, err)
			require.NotNil(t, configuration.Databases.Postgres)
			require.NotNil(t, configuration.Databases.Redis)
			require.NotNil(t, configuration.Security.Session)
			require.NotNil(t, configuration.Security.Session.TTL)
			require.Equal(t, 30*time.Minute, configuration.Databases.Postgres.Pool.MaxConnectionsLifetime)
		})
	}
}
