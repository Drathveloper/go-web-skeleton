package model

import "time"

type RedisConfig struct {
	Username        *string                 `yaml:"username"          validate:"omitempty,notblank"`
	Password        *string                 `yaml:"password"          validate:"omitempty,notblank"`
	TLS             *RedisTLSConfig         `yaml:"tls"`
	RetryPolicy     *RedisRetryPolicyConfig `yaml:"retry_policy"`
	Timeout         *RedisTimeoutConfig     `yaml:"timeout"`
	Pool            *RedisPoolConfig        `yaml:"pool"`
	Hosts           []string                `yaml:"hosts"             validate:"required"`
	IsSingleCluster bool                    `yaml:"is_single_cluster"`
}

type RedisTLSConfig struct {
	CACertificate      *string `yaml:"ca_cert"`
	ClientKey          *string `yaml:"client_key"`
	ClientCertificate  *string `yaml:"client_cert"`
	Enabled            bool    `yaml:"enabled"`
	InsecureSkipVerify bool    `yaml:"insecure_skip_verify"`
}

type RedisRetryPolicyConfig struct {
	MaxRetries      *int           `yaml:"max_retries"       validate:"omitempty,gte=0"`
	MinRetryBackoff *time.Duration `yaml:"min_retry_backoff"`
	MaxRetryBackoff *time.Duration `yaml:"max_retry_backoff"`
}

type RedisTimeoutConfig struct {
	DialTimeout  *time.Duration `yaml:"dial_timeout"`
	ReadTimeout  *time.Duration `yaml:"read_timeout"`
	WriteTimeout *time.Duration `yaml:"write_timeout"`
}

type RedisPoolConfig struct {
	MaxConnections        *int           `yaml:"max_connections"          validate:"omitempty,gt=0"`
	MinIdleConnections    *int           `yaml:"min_idle_connections"     validate:"omitempty,gte=0"`
	MaxIdleConnections    *int           `yaml:"max_idle_connections"     validate:"omitempty,gt=0"`
	MaxActiveConnections  *int           `yaml:"max_active_connections"   validate:"omitempty,gte=0"`
	Timeout               *time.Duration `yaml:"timeout"`
	MaxConnectionIdleTime *time.Duration `yaml:"max_connection_idle_time"`
	MaxConnectionLifetime *time.Duration `yaml:"max_connection_lifetime"`
}
