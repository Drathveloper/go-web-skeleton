package model

import "time"

type Configuration struct {
	Databases DatabasesConfig `yaml:"databases"`
	Security  SecurityConfig  `yaml:"security"`
	Logging   LoggingConfig   `yaml:"logging"`
	Server    ServerConfig    `yaml:"server"`
	Events    EventsConfig    `yaml:"events"`
}

type ServerConfig struct {
	ReadTimeout       time.Duration `yaml:"read_timeout"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout"`
	IdleTimeout       time.Duration `yaml:"idle_timeout"`
	MaxHeaderBytes    int           `yaml:"max_header_bytes"`
	ShutdownTimeout   time.Duration `yaml:"shutdown_timeout"`
}

type LoggingConfig struct {
	Level              string   `yaml:"level"               validate:"omitempty,oneof=DEBUG INFO WARN ERROR"`
	ConfidentialFields []string `yaml:"confidential_fields"`
}

type DatabasesConfig struct {
	Postgres *PostgresConfig `yaml:"postgres" validate:"required"`
	Redis    *RedisConfig    `yaml:"redis"    validate:"required"`
}

type SecurityConfig struct {
	Session *SessionConfig `yaml:"session" validate:"required"`
}

type SessionConfig struct {
	TTL          *int64 `yaml:"ttl"           validate:"required,gte=0"`
	CookieDomain string `yaml:"cookie_domain"`
	SecureCookie bool   `yaml:"secure_cookie"`
}

type EventsConfig struct {
	BufferSize        int           `yaml:"buffer_size"        validate:"required"`
	WorkerConcurrency int           `yaml:"worker_concurrency" validate:"required"`
	Timeout           time.Duration `yaml:"timeout"            validate:"required"`
	ShutdownTimeout   time.Duration `yaml:"shutdown_timeout"   validate:"required"`
}
