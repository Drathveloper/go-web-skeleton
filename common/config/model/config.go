package model

import "time"

type Configuration struct {
	Databases DatabasesConfig `koanf:"databases"`
	Security  SecurityConfig  `koanf:"security"`
	Logging   LoggingConfig   `koanf:"logging"`
	Server    ServerConfig    `koanf:"server"`
	Events    EventsConfig    `koanf:"events"`
}

type ServerConfig struct {
	ReadTimeout       time.Duration `koanf:"read_timeout"`
	ReadHeaderTimeout time.Duration `koanf:"read_header_timeout"`
	WriteTimeout      time.Duration `koanf:"write_timeout"`
	IdleTimeout       time.Duration `koanf:"idle_timeout"`
	MaxHeaderBytes    int           `koanf:"max_header_bytes"`
	ShutdownTimeout   time.Duration `koanf:"shutdown_timeout"`
}

type LoggingConfig struct {
	Level              string   `koanf:"level"               validate:"omitempty,oneof=DEBUG INFO WARN ERROR"`
	ConfidentialFields []string `koanf:"confidential_fields"`
}

type DatabasesConfig struct {
	Postgres *PostgresConfig `koanf:"postgres" validate:"required"`
	Redis    *RedisConfig    `koanf:"redis"    validate:"required"`
}

type SecurityConfig struct {
	Session *SessionConfig `koanf:"session" validate:"required"`
}

type SessionConfig struct {
	TTL          *int64 `koanf:"ttl"           validate:"required,gte=0"`
	CookieDomain string `koanf:"cookie_domain"`
	SecureCookie bool   `koanf:"secure_cookie"`
}

type EventsConfig struct {
	BufferSize        int           `koanf:"buffer_size"        validate:"required"`
	WorkerConcurrency int           `koanf:"worker_concurrency" validate:"required"`
	Timeout           time.Duration `koanf:"timeout"            validate:"required"`
	ShutdownTimeout   time.Duration `koanf:"shutdown_timeout"   validate:"required"`
}
