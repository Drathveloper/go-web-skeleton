package model

import "time"

type RedisConfig struct {
	Username        *string                 `koanf:"username"          validate:"omitempty,notblank"`
	Password        *string                 `koanf:"password"          validate:"omitempty,notblank"`
	TLS             *RedisTLSConfig         `koanf:"tls"`
	RetryPolicy     *RedisRetryPolicyConfig `koanf:"retry_policy"`
	Timeout         *RedisTimeoutConfig     `koanf:"timeout"`
	Pool            *RedisPoolConfig        `koanf:"pool"`
	Hosts           []string                `koanf:"hosts"             validate:"required"`
	IsSingleCluster bool                    `koanf:"is_single_cluster"`
}

type RedisTLSConfig struct {
	CACertificate      *string `koanf:"ca_cert"`
	ClientKey          *string `koanf:"client_key"`
	ClientCertificate  *string `koanf:"client_cert"`
	Enabled            bool    `koanf:"enabled"`
	InsecureSkipVerify bool    `koanf:"insecure_skip_verify"`
}

type RedisRetryPolicyConfig struct {
	MaxRetries      *int           `koanf:"max_retries"       validate:"omitempty,gte=0"`
	MinRetryBackoff *time.Duration `koanf:"min_retry_backoff"`
	MaxRetryBackoff *time.Duration `koanf:"max_retry_backoff"`
}

type RedisTimeoutConfig struct {
	DialTimeout  *time.Duration `koanf:"dial_timeout"`
	ReadTimeout  *time.Duration `koanf:"read_timeout"`
	WriteTimeout *time.Duration `koanf:"write_timeout"`
}

type RedisPoolConfig struct {
	MaxConnections        *int           `koanf:"max_connections"          validate:"omitempty,gt=0"`
	MinIdleConnections    *int           `koanf:"min_idle_connections"     validate:"omitempty,gte=0"`
	MaxIdleConnections    *int           `koanf:"max_idle_connections"     validate:"omitempty,gt=0"`
	MaxActiveConnections  *int           `koanf:"max_active_connections"   validate:"omitempty,gte=0"`
	Timeout               *time.Duration `koanf:"timeout"`
	MaxConnectionIdleTime *time.Duration `koanf:"max_connection_idle_time"`
	MaxConnectionLifetime *time.Duration `koanf:"max_connection_lifetime"`
}
