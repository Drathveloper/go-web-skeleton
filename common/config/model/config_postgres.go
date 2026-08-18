package model

import (
	"strconv"
	"strings"
	"time"
)

type PostgresConfig struct {
	Host     string             `validate:"required,notblank" koanf:"host"`
	Database string             `validate:"required,notblank" koanf:"database"`
	User     string             `validate:"required,notblank" koanf:"user"`
	Password string             `validate:"required,notblank" koanf:"password"`
	SSL      PostgresSSLConfig  `koanf:"ssl"`
	Pool     PostgresPoolConfig `koanf:"pool"`
	Port     int                `validate:"required,notblank" koanf:"port"`
}

func (c PostgresConfig) String() string {
	builder := strings.Builder{}

	builder.WriteString("host=")
	builder.WriteString(c.Host)
	builder.WriteString(" ")

	builder.WriteString("port=")
	builder.WriteString(strconv.Itoa(c.Port))
	builder.WriteString(" ")

	builder.WriteString("user=")
	builder.WriteString(c.User)
	builder.WriteString(" ")

	builder.WriteString("password=")
	builder.WriteString(c.Password)
	builder.WriteString(" ")

	builder.WriteString("dbname=")
	builder.WriteString(c.Database)
	builder.WriteString(" ")

	builder.WriteString("sslmode=")
	builder.WriteString(c.SSL.Mode)

	return builder.String()
}

type PostgresSSLConfig struct {
	Mode string `validate:"omitempty,oneof=disable allow prefer require verify-ca verify-full" koanf:"mode"`
}

func (c PostgresSSLConfig) String() string {
	return "sslmode=" + c.Mode
}

type PostgresPoolConfig struct {
	MaxConnections         int32         `validate:"omitempty,gt=0"        koanf:"max_connections"`
	MaxIdleConnections     int32         `validate:"omitempty,gt=0"        koanf:"max_idle_connections"`
	MinConnections         int32         `validate:"omitempty,gte=0"       koanf:"min_connections"`
	MaxConnectionsLifetime time.Duration `koanf:"max_lifetime"`
	MaxIdleTime            time.Duration `koanf:"max_idle_time"`
	HealthCheckPeriod      time.Duration `koanf:"health_check_period"`
	MaxConnLifetimeJitter  time.Duration `koanf:"max_conn_lifetime_jitter"`
}
