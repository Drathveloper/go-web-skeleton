package wire

import (
	"fmt"

	"github.com/Drathveloper/go-web-skeleton/common/config"
	"github.com/Drathveloper/go-web-skeleton/common/config/model"
	"github.com/Drathveloper/go-web-skeleton/common/constants"
)

const injectConfigBaseErrMsg = "inject configuration failed"

type RequiredConfigs struct {
	Store *config.Store
	Env   *model.EnvConfig
}

func injectConfig(container *Container) error {
	envConfig, err := config.LoadEnv(container.buildInfo)
	if err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, injectConfigBaseErrMsg, err)
	}
	container.Env = envConfig
	configuration, err := config.ReadConfig(container.fs)
	if err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, injectConfigBaseErrMsg, err)
	}
	if err = container.Validate.Struct(configuration); err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, injectConfigBaseErrMsg, err)
	}
	container.Store = config.NewStore(configuration)

	return nil
}
