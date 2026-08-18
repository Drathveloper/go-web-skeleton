package model

import (
	"strconv"
	"strings"
	"time"
)

type PostgresConfig struct {
	Host     string             `yaml:"host"     validate:"required,notblank"`
	Database string             `yaml:"database" validate:"required,notblank"`
	User     string             `yaml:"user"     validate:"required,notblank"`
	Password string             `yaml:"password" validate:"required,notblank"`
	SSL      PostgresSSLConfig  `yaml:"ssl"`
	Pool     PostgresPoolConfig `yaml:"pool"`
	Port     int                `yaml:"port"     validate:"required,notblank"`
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
	Mode string `yaml:"mode" validate:"omitempty,oneof=disable allow prefer require verify-ca verify-full"`
}

func (c PostgresSSLConfig) String() string {
	return "sslmode=" + c.Mode
}

type PostgresPoolConfig struct {
	MaxConnections         int32         `yaml:"max_connections"          validate:"omitempty,gt=0"`
	MaxIdleConnections     int32         `yaml:"max_idle_connections"     validate:"omitempty,gt=0"`
	MinConnections         int32         `yaml:"min_connections"          validate:"omitempty,gte=0"`
	MaxConnectionsLifetime time.Duration `yaml:"max_lifetime"`
	MaxIdleTime            time.Duration `yaml:"max_idle_time"`
	HealthCheckPeriod      time.Duration `yaml:"health_check_period"`
	MaxConnLifetimeJitter  time.Duration `yaml:"max_conn_lifetime_jitter"`
}
