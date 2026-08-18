package model

import "fmt"

type BuildInfo struct {
	Commit    string
	BuildTime string
	Version   string
}

type EnvConfig struct {
	BuildInfo

	GinMode         string `env:"GIN_MODE"      envDefault:"release"`
	Port            string `env:"PORT"          envDefault:"8000"`
	Environment     string `env:"ENVIRONMENT"   envDefault:"dev"`
	ServiceName     string `env:"SERVICE_NAME"  envDefault:"go-web-skeleton"`
	TLSCertFilePath string `env:"TLS_CERT_FILE"`
	TLSKeyFilePath  string `env:"TLS_KEY_FILE"`
	// SeedAdminUsername and SeedAdminPassword create the first administrator on
	// an empty database. There is deliberately no default: a template that
	// shipped a known password would put one in every project generated from
	// it, and every deployment that forgot to change it.
	SeedAdminUsername string `env:"SEED_ADMIN_USERNAME"`
	SeedAdminPassword string `env:"SEED_ADMIN_PASSWORD"`
	EnableTLS         bool   `env:"ENABLE_TLS"          envDefault:"false"`
}

func (e EnvConfig) String() string {
	return fmt.Sprintf("service_name=%s environment=%s, port=%s, version=%s, commit=%s, build_time=%s",
		e.ServiceName, e.Environment, e.Port, e.Version, e.Commit, e.BuildTime)
}
