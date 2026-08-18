package config

import (
	"fmt"
	"sync"

	"github.com/caarlos0/env/v11"

	"github.com/Drathveloper/go-web-skeleton/common/config/model"
	"github.com/Drathveloper/go-web-skeleton/common/constants"
)

const loadEnvErrMsg = "load env config failed"

//nolint:gochecknoglobals
var (
	once sync.Once
	Env  *model.EnvConfig
)

func LoadEnv(buildInfo model.BuildInfo) error {
	var err error
	once.Do(func() {
		instance := &model.EnvConfig{}
		err = env.Parse(instance)
		Env = instance
	})
	if err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, loadEnvErrMsg, err)
	}
	Env.BuildInfo = buildInfo
	return nil
}

func ResetLoadEnv() {
	once = sync.Once{}
}
