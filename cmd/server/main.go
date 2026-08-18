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

	// ConfigFS carries the YAML configuration into the binary.
	// config/application.yaml is intentionally NOT committed: only
	// application.example.yaml is. The glob keeps the build working from a
	// fresh clone, while startup still fails loudly when the real file is
	// missing, instead of silently running on a checked-in default.
	//go:embed config/*.yaml
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
