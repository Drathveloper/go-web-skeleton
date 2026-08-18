package model

import "time"

type RedisConfig struct {
	Username        *string                 `validate:"omitempty,notblank" koanf:"username"`
	Password        *string                 `validate:"omitempty,notblank" koanf:"password"`
	TLS             *RedisTLSConfig         `koanf:"tls"`
	RetryPolicy     *RedisRetryPolicyConfig `koanf:"retry_policy"`
	Timeout         *RedisTimeoutConfig     `koanf:"timeout"`
	Pool            *RedisPoolConfig        `koanf:"pool"`
	Hosts           []string                `validate:"required"           koanf:"hosts"`
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
	MaxRetries      *int           `validate:"omitempty,gte=0" koanf:"max_retries"`
	MinRetryBackoff *time.Duration `koanf:"min_retry_backoff"`
	MaxRetryBackoff *time.Duration `koanf:"max_retry_backoff"`
}

type RedisTimeoutConfig struct {
	DialTimeout  *time.Duration `koanf:"dial_timeout"`
	ReadTimeout  *time.Duration `koanf:"read_timeout"`
	WriteTimeout *time.Duration `koanf:"write_timeout"`
}

type RedisPoolConfig struct {
	MaxConnections        *int           `validate:"omitempty,gt=0"        koanf:"max_connections"`
	MinIdleConnections    *int           `validate:"omitempty,gte=0"       koanf:"min_idle_connections"`
	MaxIdleConnections    *int           `validate:"omitempty,gt=0"        koanf:"max_idle_connections"`
	MaxActiveConnections  *int           `validate:"omitempty,gte=0"       koanf:"max_active_connections"`
	Timeout               *time.Duration `koanf:"timeout"`
	MaxConnectionIdleTime *time.Duration `koanf:"max_connection_idle_time"`
	MaxConnectionLifetime *time.Duration `koanf:"max_connection_lifetime"`
}
