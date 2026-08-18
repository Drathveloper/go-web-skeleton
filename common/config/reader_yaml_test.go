package config_test

import (
	"os"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Drathveloper/go-web-skeleton/common/config"
)

const configPath = "config/application.yaml"

func configFS(t *testing.T, content string) fstest.MapFS {
	t.Helper()

	return fstest.MapFS{
		configPath: &fstest.MapFile{Data: []byte(content)},
	}
}

func TestReadYAMLConfig(t *testing.T) {
	t.Parallel()

	content := `server:
  read_timeout: 10s
  max_header_bytes: 2048
logging:
  level: DEBUG
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

	configuration, err := config.ReadYAMLConfig(configFS(t, content))

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

// The regression this test exists for: with non-strict decoding, a key the model
// does not have was silently dropped. `max_connection_lifetime` (the model calls
// it `max_lifetime`) sat in the source's production config doing nothing — the
// pool never got the lifetime it was configured with, and nothing said so.
// KnownFields(true) makes it a startup failure that names the offending key.
func TestReadYAMLConfig_UnknownKeyIsAHardError(t *testing.T) {
	t.Parallel()

	content := `databases:
  postgres:
    host: localhost
    pool:
      max_connection_lifetime: 30m
`

	configuration, err := config.ReadYAMLConfig(configFS(t, content))

	require.Nil(t, configuration)
	require.Error(t, err)
	require.Equal(t,
		"read yaml config failed: yaml: unmarshal errors:\n"+
			"  line 5: field max_connection_lifetime not found in type model.PostgresPoolConfig",
		err.Error())
}

func TestReadYAMLConfig_Failures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		fs             fstest.MapFS
		name           string
		expectedErrMsg string
	}{
		{
			name: "test read yaml config should fail when the config file is missing",
			fs:   fstest.MapFS{},
			expectedErrMsg: "read yaml config failed: read config file failed: " +
				"open config/application.yaml: file does not exist",
		},
		{
			name:           "test read yaml config should fail when the yaml is malformed",
			fs:             configFS(t, "databases: [1,2\n"),
			expectedErrMsg: "read yaml config failed: yaml: line 1: did not find expected ',' or ']'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			configuration, err := config.ReadYAMLConfig(tt.fs)

			require.Nil(t, configuration)
			require.Error(t, err)
			require.Equal(t, tt.expectedErrMsg, err.Error())
		})
	}
}

// application.example.yaml is what every new project copies to
// application.yaml. Now that an unrecognised key aborts startup, an example that
// drifted from the model would break the first run of every generated project,
// and nothing else in the build would catch it.
func TestReadYAMLConfig_ExampleConfigStillMatchesTheModel(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile("../../cmd/server/config/application.example.yaml")
	require.NoError(t, err)

	configuration, err := config.ReadYAMLConfig(configFS(t, string(content)))

	require.NoError(t, err)
	require.NotNil(t, configuration.Databases.Postgres)
	require.NotNil(t, configuration.Databases.Redis)
	require.NotNil(t, configuration.Security.Session)
	require.NotNil(t, configuration.Security.Session.TTL)
}
