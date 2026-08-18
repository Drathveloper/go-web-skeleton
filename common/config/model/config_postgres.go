package model

import (
	"strconv"
	"strings"
	"time"
)

type PostgresConfig struct {
	Host     string             `koanf:"host"     validate:"required,notblank"`
	Database string             `koanf:"database" validate:"required,notblank"`
	User     string             `koanf:"user"     validate:"required,notblank"`
	Password string             `koanf:"password" validate:"required,notblank"`
	SSL      PostgresSSLConfig  `koanf:"ssl"`
	Pool     PostgresPoolConfig `koanf:"pool"`
	Port     int                `koanf:"port"     validate:"required,notblank"`
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
	Mode string `koanf:"mode" validate:"omitempty,oneof=disable allow prefer require verify-ca verify-full"`
}

func (c PostgresSSLConfig) String() string {
	return "sslmode=" + c.Mode
}

type PostgresPoolConfig struct {
	MaxConnections         int32         `koanf:"max_connections"          validate:"omitempty,gt=0"`
	MaxIdleConnections     int32         `koanf:"max_idle_connections"     validate:"omitempty,gt=0"`
	MinConnections         int32         `koanf:"min_connections"          validate:"omitempty,gte=0"`
	MaxConnectionsLifetime time.Duration `koanf:"max_lifetime"`
	MaxIdleTime            time.Duration `koanf:"max_idle_time"`
	HealthCheckPeriod      time.Duration `koanf:"health_check_period"`
	MaxConnLifetimeJitter  time.Duration `koanf:"max_conn_lifetime_jitter"`
}
