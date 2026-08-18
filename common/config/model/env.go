package model

import "fmt"

type BuildInfo struct {
	Commit    string
	BuildTime string
	Version   string
}

// EnvConfig holds the process-level settings read from the environment. The
// variable names and defaults live in common/config/env.go: names are mapped
// explicitly there so the deployment contract (PORT, GIN_MODE, ...) is a
// closed list, not whatever happens to be in the process environment.
type EnvConfig struct {
	BuildInfo

	GinMode         string `koanf:"gin_mode"`
	Port            string `koanf:"port"`
	Environment     string `koanf:"environment"`
	ServiceName     string `koanf:"service_name"`
	TLSCertFilePath string `koanf:"tls_cert_file"`
	TLSKeyFilePath  string `koanf:"tls_key_file"`
	// SeedAdminUsername and SeedAdminPassword create the first administrator on
	// an empty database. There is deliberately no default: a template that
	// shipped a known password would put one in every project generated from
	// it, and every deployment that forgot to change it.
	SeedAdminUsername string `koanf:"seed_admin_username"`
	SeedAdminPassword string `koanf:"seed_admin_password"`
	EnableTLS         bool   `koanf:"enable_tls"`
}

func (e EnvConfig) String() string {
	return fmt.Sprintf("service_name=%s environment=%s, port=%s, version=%s, commit=%s, build_time=%s",
		e.ServiceName, e.Environment, e.Port, e.Version, e.Commit, e.BuildTime)
}
