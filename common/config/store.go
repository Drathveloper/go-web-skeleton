package config

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Drathveloper/go-web-skeleton/common/config/model"
	"github.com/Drathveloper/go-web-skeleton/common/constants"
	"github.com/Drathveloper/go-web-skeleton/pkg/event"
)

const (
	defaultReadTimeout       = 15 * time.Second
	defaultReadHeaderTimeout = 5 * time.Second
	defaultWriteTimeout      = 15 * time.Second
	defaultIdleTimeout       = 120 * time.Second
	defaultShutdownTimeout   = 30 * time.Second
)

type Store struct {
	configuration model.Configuration
}

func NewStore(configuration *model.Configuration) *Store {
	return &Store{
		configuration: *configuration,
	}
}

func (s *Store) GetServerReadTimeout() time.Duration {
	if s.configuration.Server.ReadTimeout == 0 {
		return defaultReadTimeout
	}
	return s.configuration.Server.ReadTimeout
}

func (s *Store) GetServerReadHeaderTimeout() time.Duration {
	if s.configuration.Server.ReadHeaderTimeout == 0 {
		return defaultReadHeaderTimeout
	}
	return s.configuration.Server.ReadHeaderTimeout
}

func (s *Store) GetServerWriteTimeout() time.Duration {
	if s.configuration.Server.WriteTimeout == 0 {
		return defaultWriteTimeout
	}
	return s.configuration.Server.WriteTimeout
}

func (s *Store) GetServerIdleTimeout() time.Duration {
	if s.configuration.Server.IdleTimeout == 0 {
		return defaultIdleTimeout
	}
	return s.configuration.Server.IdleTimeout
}

func (s *Store) GetServerMaxHeaderBytes() int {
	if s.configuration.Server.MaxHeaderBytes == 0 {
		return http.DefaultMaxHeaderBytes
	}
	return s.configuration.Server.MaxHeaderBytes
}

func (s *Store) GetLoggingLevel() slog.Level {
	switch s.configuration.Logging.Level {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func (s *Store) GetLoggingConfidentialFields() []string {
	return s.configuration.Logging.ConfidentialFields
}

func (s *Store) GetPostgresConnectionString() string {
	return s.configuration.Databases.Postgres.String()
}

func (s *Store) GetPostgresPoolConfig() model.PostgresPoolConfig {
	return s.configuration.Databases.Postgres.Pool
}

func (s *Store) GetRedisOptions() (*redis.UniversalOptions, error) {
	opts := &redis.UniversalOptions{
		Addrs:         s.configuration.Databases.Redis.Hosts,
		IsClusterMode: s.configuration.Databases.Redis.IsSingleCluster,
	}
	if s.configuration.Databases.Redis.Username != nil {
		opts.Username = *s.configuration.Databases.Redis.Username
	}
	if s.configuration.Databases.Redis.Password != nil {
		opts.Password = *s.configuration.Databases.Redis.Password
	}
	if err := s.getRedisTLSOptions(opts); err != nil {
		return nil, err
	}
	s.getRedisPoolOptions(opts)
	s.getRedisRetryPolicyOptions(opts)
	s.getRedisTimeoutOptions(opts)
	return opts, nil
}

func (s *Store) GetSessionTTL() time.Duration {
	return time.Duration(*s.configuration.Security.Session.TTL) * time.Second
}

func (s *Store) IsSecureCookie() bool {
	return s.configuration.Security.Session.SecureCookie
}

func (s *Store) GetCookieDomain() string {
	return s.configuration.Security.Session.CookieDomain
}

func (s *Store) GetEventBusConfig() event.Options {
	opts := event.NewDefaultOptions()
	if s.configuration.Events.BufferSize != 0 {
		opts.ChannelBuffer = s.configuration.Events.BufferSize
	}
	if s.configuration.Events.WorkerConcurrency != 0 {
		opts.WorkerConcurrency = s.configuration.Events.WorkerConcurrency
	}
	if s.configuration.Events.Timeout != 0 {
		opts.DefaultTimeout = s.configuration.Events.Timeout
	}
	return opts
}

func (s *Store) GetEventShutdownTimeout() time.Duration {
	if s.configuration.Events.ShutdownTimeout == 0 {
		return defaultShutdownTimeout
	}
	return s.configuration.Events.ShutdownTimeout
}

func (s *Store) GetHTTPServerShutdownTimeout() time.Duration {
	if s.configuration.Server.ShutdownTimeout == 0 {
		return defaultShutdownTimeout
	}
	return s.configuration.Server.ShutdownTimeout
}

func (s *Store) getRedisTLSOptions(opts *redis.UniversalOptions) error {
	tlsConfig := s.configuration.Databases.Redis.TLS
	if tlsConfig == nil || !tlsConfig.Enabled {
		return nil
	}
	opts.TLSConfig = &tls.Config{
		InsecureSkipVerify: tlsConfig.InsecureSkipVerify, //nolint:gosec
	}
	if tlsConfig.CACertificate != nil {
		caPool := x509.NewCertPool()
		if !caPool.AppendCertsFromPEM([]byte(*tlsConfig.CACertificate)) {
			return ErrInvalidRedisCACertificate
		}
		opts.TLSConfig.RootCAs = caPool
	}
	// A certificate without its key is a half-filled config, not a usable
	// keypair: dereferencing ClientKey here is what made the source panic.
	if tlsConfig.ClientCertificate != nil || tlsConfig.ClientKey != nil {
		if tlsConfig.ClientCertificate == nil || tlsConfig.ClientKey == nil {
			return ErrIncompleteRedisClientKeypair
		}
		clientCert, err := tls.X509KeyPair(
			[]byte(*tlsConfig.ClientCertificate),
			[]byte(*tlsConfig.ClientKey),
		)
		if err != nil {
			return fmt.Errorf(constants.DefaultWrappedErrorTemplate, redisTLSOptionsBaseErrMsg, err)
		}
		opts.TLSConfig.Certificates = []tls.Certificate{clientCert}
	}
	return nil
}

//nolint:nestif
func (s *Store) getRedisPoolOptions(opts *redis.UniversalOptions) {
	if s.configuration.Databases.Redis.Pool != nil {
		if s.configuration.Databases.Redis.Pool.MaxConnections != nil {
			opts.PoolSize = *s.configuration.Databases.Redis.Pool.MaxConnections
		}
		if s.configuration.Databases.Redis.Pool.MinIdleConnections != nil {
			opts.MinIdleConns = *s.configuration.Databases.Redis.Pool.MinIdleConnections
		}
		if s.configuration.Databases.Redis.Pool.MaxIdleConnections != nil {
			opts.MaxIdleConns = *s.configuration.Databases.Redis.Pool.MaxIdleConnections
		}
		if s.configuration.Databases.Redis.Pool.MaxActiveConnections != nil {
			opts.MaxActiveConns = *s.configuration.Databases.Redis.Pool.MaxActiveConnections
		}
		if s.configuration.Databases.Redis.Pool.Timeout != nil {
			opts.PoolTimeout = *s.configuration.Databases.Redis.Pool.Timeout
		}
		if s.configuration.Databases.Redis.Pool.MaxConnectionIdleTime != nil {
			opts.ConnMaxIdleTime = *s.configuration.Databases.Redis.Pool.MaxConnectionIdleTime
		}
		if s.configuration.Databases.Redis.Pool.MaxConnectionLifetime != nil {
			opts.ConnMaxLifetime = *s.configuration.Databases.Redis.Pool.MaxConnectionLifetime
		}
	}
}

func (s *Store) getRedisRetryPolicyOptions(opts *redis.UniversalOptions) {
	if s.configuration.Databases.Redis.RetryPolicy != nil {
		if s.configuration.Databases.Redis.RetryPolicy.MaxRetries != nil {
			opts.MaxRetries = *s.configuration.Databases.Redis.RetryPolicy.MaxRetries
		}
		if s.configuration.Databases.Redis.RetryPolicy.MinRetryBackoff != nil {
			opts.MinRetryBackoff = *s.configuration.Databases.Redis.RetryPolicy.MinRetryBackoff
		}
		if s.configuration.Databases.Redis.RetryPolicy.MaxRetryBackoff != nil {
			opts.MaxRetryBackoff = *s.configuration.Databases.Redis.RetryPolicy.MaxRetryBackoff
		}
	}
}

func (s *Store) getRedisTimeoutOptions(opts *redis.UniversalOptions) {
	if s.configuration.Databases.Redis.Timeout != nil {
		if s.configuration.Databases.Redis.Timeout.ReadTimeout != nil {
			opts.ReadTimeout = *s.configuration.Databases.Redis.Timeout.ReadTimeout
		}
		if s.configuration.Databases.Redis.Timeout.WriteTimeout != nil {
			opts.WriteTimeout = *s.configuration.Databases.Redis.Timeout.WriteTimeout
		}
		if s.configuration.Databases.Redis.Timeout.DialTimeout != nil {
			opts.DialTimeout = *s.configuration.Databases.Redis.Timeout.DialTimeout
		}
	}
}

const redisTLSOptionsBaseErrMsg = "build redis TLS options failed"

// Misconfigured TLS used to panic at startup deep inside a getter. These make
// it an ordinary error the injector can report with context.
var (
	ErrInvalidRedisCACertificate    = errors.New("failed to append provided redis CA certificates")
	ErrIncompleteRedisClientKeypair = errors.New("redis client certificate and client key must be set together")
)
