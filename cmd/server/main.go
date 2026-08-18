//nolint:gochecknoglobals
package main

import (
	"embed"
	"log"

	"github.com/Drathveloper/go-web-skeleton/common/bootstrap"
	"github.com/Drathveloper/go-web-skeleton/common/config/model"
)

var (
	Commit    string
	BuildTime string
	Version   string

	// ConfigFS carries the configuration into the binary: application.yaml,
	// or application.json as fallback when no YAML exists. Neither real file
	// is committed, only the example variants are. The globs keep the build
	// working from a fresh clone (a go:embed pattern with no match breaks
	// compilation, which is why application.example.json must stay committed
	// too), while startup still fails loudly when the real file is missing,
	// instead of silently running on a checked-in default.
	//go:embed config/*.yaml config/*.json
	ConfigFS embed.FS
)

func main() {
	buildInfo := &model.BuildInfo{
		Commit:    Commit,
		BuildTime: BuildTime,
		Version:   Version,
	}
	if err := bootstrap.Run(ConfigFS, buildInfo); err != nil {
		log.Fatal(err)
	}
}
