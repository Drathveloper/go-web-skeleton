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
	Level              string   `validate:"omitempty,oneof=DEBUG INFO WARN ERROR" koanf:"level"`
	ConfidentialFields []string `koanf:"confidential_fields"`
}

type DatabasesConfig struct {
	Postgres *PostgresConfig `validate:"required" koanf:"postgres"`
	Redis    *RedisConfig    `validate:"required" koanf:"redis"`
}

type SecurityConfig struct {
	Session *SessionConfig `validate:"required" koanf:"session"`
}

type SessionConfig struct {
	TTL          *int64 `validate:"required,gte=0" koanf:"ttl"`
	CookieDomain string `koanf:"cookie_domain"`
	SecureCookie bool   `koanf:"secure_cookie"`
}

type EventsConfig struct {
	BufferSize        int           `validate:"required" koanf:"buffer_size"`
	WorkerConcurrency int           `validate:"required" koanf:"worker_concurrency"`
	Timeout           time.Duration `validate:"required" koanf:"timeout"`
	ShutdownTimeout   time.Duration `validate:"required" koanf:"shutdown_timeout"`
}
