package config_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log/slog"
	"math/big"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Drathveloper/go-web-skeleton/common/config"
	"github.com/Drathveloper/go-web-skeleton/common/config/model"
)

// newKeyPair mints a throwaway self-signed certificate. Generating it here keeps
// the suite free of checked-in PEM fixtures that expire.
func newKeyPair(t *testing.T, commonName string) (string, string) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: commonName},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	require.NoError(t, err)

	keyDER, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)

	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})),
		string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}))
}

func newStore(redisConfig *model.RedisConfig) *config.Store {
	return config.NewStore(&model.Configuration{
		Databases: model.DatabasesConfig{
			Postgres: &model.PostgresConfig{},
			Redis:    redisConfig,
		},
	})
}

func TestStore_GetRedisOptions(t *testing.T) {
	t.Parallel()

	username := "someUser"
	password := "somePassword"
	maxConnections := 20
	dialTimeout := 3 * time.Second
	maxRetries := 4

	store := newStore(&model.RedisConfig{
		Hosts:           []string{"redis-1:6379", "redis-2:6379"},
		IsSingleCluster: true,
		Username:        &username,
		Password:        &password,
		Pool:            &model.RedisPoolConfig{MaxConnections: &maxConnections},
		Timeout:         &model.RedisTimeoutConfig{DialTimeout: &dialTimeout},
		RetryPolicy:     &model.RedisRetryPolicyConfig{MaxRetries: &maxRetries},
	})

	opts, err := store.GetRedisOptions()

	require.NoError(t, err)
	require.Equal(t, []string{"redis-1:6379", "redis-2:6379"}, opts.Addrs)
	require.True(t, opts.IsClusterMode)
	require.Equal(t, "someUser", opts.Username)
	require.Equal(t, "somePassword", opts.Password)
	require.Equal(t, 20, opts.PoolSize)
	require.Equal(t, 3*time.Second, opts.DialTimeout)
	require.Equal(t, 4, opts.MaxRetries)
	require.Nil(t, opts.TLSConfig)
}

func TestStore_GetRedisOptionsTLS(t *testing.T) {
	t.Parallel()

	caCert, _ := newKeyPair(t, "ca")
	clientCert, clientKey := newKeyPair(t, "client")
	_, otherKey := newKeyPair(t, "other")

	tests := []struct {
		tlsConfig *model.RedisTLSConfig
		assert    func(t *testing.T, store *config.Store)
		name      string
	}{
		{
			name:      "test get redis options should not build a TLS config when TLS is disabled",
			tlsConfig: &model.RedisTLSConfig{Enabled: false, CACertificate: &caCert},
			assert: func(t *testing.T, store *config.Store) {
				t.Helper()
				opts, err := store.GetRedisOptions()
				require.NoError(t, err)
				require.Nil(t, opts.TLSConfig)
			},
		},
		{
			name:      "test get redis options should load the CA bundle",
			tlsConfig: &model.RedisTLSConfig{Enabled: true, CACertificate: &caCert, InsecureSkipVerify: true},
			assert: func(t *testing.T, store *config.Store) {
				t.Helper()
				opts, err := store.GetRedisOptions()
				require.NoError(t, err)
				require.NotNil(t, opts.TLSConfig.RootCAs)
				require.True(t, opts.TLSConfig.InsecureSkipVerify)
			},
		},
		{
			name: "test get redis options should load the client keypair",
			tlsConfig: &model.RedisTLSConfig{
				Enabled: true, ClientCertificate: &clientCert, ClientKey: &clientKey,
			},
			assert: func(t *testing.T, store *config.Store) {
				t.Helper()
				opts, err := store.GetRedisOptions()
				require.NoError(t, err)
				require.Len(t, opts.TLSConfig.Certificates, 1)
			},
		},
		// The three cases below used to abort the process from inside a getter:
		// a malformed bundle panicked outright, and a certificate without its key
		// dereferenced a nil pointer. Startup misconfiguration has to be an error
		// the injector can report, not a stack trace.
		{
			name:      "test get redis options should fail when the CA bundle is not valid PEM",
			tlsConfig: &model.RedisTLSConfig{Enabled: true, CACertificate: new("-----BEGIN NONSENSE-----")},
			assert: func(t *testing.T, store *config.Store) {
				t.Helper()
				opts, err := store.GetRedisOptions()
				require.Nil(t, opts)
				require.ErrorIs(t, err, config.ErrInvalidRedisCACertificate)
				require.Equal(t, "failed to append provided redis CA certificates", err.Error())
			},
		},
		{
			name:      "test get redis options should fail when the client certificate has no key",
			tlsConfig: &model.RedisTLSConfig{Enabled: true, ClientCertificate: &clientCert},
			assert: func(t *testing.T, store *config.Store) {
				t.Helper()
				opts, err := store.GetRedisOptions()
				require.Nil(t, opts)
				require.ErrorIs(t, err, config.ErrIncompleteRedisClientKeypair)
			},
		},
		{
			name:      "test get redis options should fail when the client key has no certificate",
			tlsConfig: &model.RedisTLSConfig{Enabled: true, ClientKey: &clientKey},
			assert: func(t *testing.T, store *config.Store) {
				t.Helper()
				opts, err := store.GetRedisOptions()
				require.Nil(t, opts)
				require.ErrorIs(t, err, config.ErrIncompleteRedisClientKeypair)
			},
		},
		{
			name: "test get redis options should fail when the certificate and the key do not match",
			tlsConfig: &model.RedisTLSConfig{
				Enabled: true, ClientCertificate: &clientCert, ClientKey: &otherKey,
			},
			assert: func(t *testing.T, store *config.Store) {
				t.Helper()
				opts, err := store.GetRedisOptions()
				require.Nil(t, opts)
				require.ErrorContains(t, err, "build redis TLS options failed: ")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.assert(t, newStore(&model.RedisConfig{
				Hosts: []string{"localhost:6379"},
				TLS:   tt.tlsConfig,
			}))
		})
	}
}

func TestStore_Defaults(t *testing.T) {
	t.Parallel()

	store := config.NewStore(&model.Configuration{
		Databases: model.DatabasesConfig{Postgres: &model.PostgresConfig{}, Redis: &model.RedisConfig{}},
	})

	require.Equal(t, 15*time.Second, store.GetServerReadTimeout())
	require.Equal(t, 5*time.Second, store.GetServerReadHeaderTimeout())
	require.Equal(t, 15*time.Second, store.GetServerWriteTimeout())
	require.Equal(t, 120*time.Second, store.GetServerIdleTimeout())
	require.Equal(t, http.DefaultMaxHeaderBytes, store.GetServerMaxHeaderBytes())
	require.Equal(t, 30*time.Second, store.GetHTTPServerShutdownTimeout())
	require.Equal(t, 30*time.Second, store.GetEventShutdownTimeout())
	require.Equal(t, slog.LevelInfo, store.GetLoggingLevel())

	busOptions := store.GetEventBusConfig()
	require.Equal(t, 100, busOptions.ChannelBuffer)
	require.Equal(t, 10, busOptions.WorkerConcurrency)
	require.Equal(t, 30*time.Second, busOptions.DefaultTimeout)
}

func TestStore_ConfiguredValues(t *testing.T) {
	t.Parallel()

	ttl := int64(3600)
	store := config.NewStore(&model.Configuration{
		Server: model.ServerConfig{
			ReadTimeout:       time.Second,
			ReadHeaderTimeout: 2 * time.Second,
			WriteTimeout:      3 * time.Second,
			IdleTimeout:       4 * time.Second,
			MaxHeaderBytes:    2048,
			ShutdownTimeout:   5 * time.Second,
		},
		Logging: model.LoggingConfig{Level: "DEBUG", ConfidentialFields: []string{"password"}},
		Security: model.SecurityConfig{Session: &model.SessionConfig{
			TTL: &ttl, SecureCookie: true, CookieDomain: "localhost",
		}},
		Events: model.EventsConfig{
			BufferSize: 7, WorkerConcurrency: 3, Timeout: 8 * time.Second, ShutdownTimeout: 9 * time.Second,
		},
		Databases: model.DatabasesConfig{
			Postgres: &model.PostgresConfig{
				Host: "localhost", Port: 5432, Database: "skeleton", User: "skeleton", Password: "secret",
				SSL: model.PostgresSSLConfig{Mode: "disable"},
			},
			Redis: &model.RedisConfig{},
		},
	})

	require.Equal(t, time.Second, store.GetServerReadTimeout())
	require.Equal(t, 2*time.Second, store.GetServerReadHeaderTimeout())
	require.Equal(t, 3*time.Second, store.GetServerWriteTimeout())
	require.Equal(t, 4*time.Second, store.GetServerIdleTimeout())
	require.Equal(t, 2048, store.GetServerMaxHeaderBytes())
	require.Equal(t, 5*time.Second, store.GetHTTPServerShutdownTimeout())
	require.Equal(t, 9*time.Second, store.GetEventShutdownTimeout())
	require.Equal(t, slog.LevelDebug, store.GetLoggingLevel())
	require.Equal(t, []string{"password"}, store.GetLoggingConfidentialFields())
	require.Equal(t, 3600*time.Second, store.GetSessionTTL())
	require.True(t, store.IsSecureCookie())
	require.Equal(t, "localhost", store.GetCookieDomain())
	require.Equal(t,
		"host=localhost port=5432 user=skeleton password=secret dbname=skeleton sslmode=disable",
		store.GetPostgresConnectionString())

	busOptions := store.GetEventBusConfig()
	require.Equal(t, 7, busOptions.ChannelBuffer)
	require.Equal(t, 3, busOptions.WorkerConcurrency)
	require.Equal(t, 8*time.Second, busOptions.DefaultTimeout)
}

func TestStore_GetLoggingLevel(t *testing.T) {
	t.Parallel()

	tests := map[string]slog.Level{
		"DEBUG":    slog.LevelDebug,
		"INFO":     slog.LevelInfo,
		"WARN":     slog.LevelWarn,
		"ERROR":    slog.LevelError,
		"":         slog.LevelInfo,
		"NONSENSE": slog.LevelInfo,
		"debug":    slog.LevelInfo,
	}
	for level, expected := range tests {
		t.Run("test logging level "+level, func(t *testing.T) {
			t.Parallel()

			store := config.NewStore(&model.Configuration{Logging: model.LoggingConfig{Level: level}})

			require.Equal(t, expected, store.GetLoggingLevel())
		})
	}
}
