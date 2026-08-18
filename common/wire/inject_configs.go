package wire

import (
	"fmt"

	"github.com/Drathveloper/go-web-skeleton/common/config"
	"github.com/Drathveloper/go-web-skeleton/common/constants"
)

const injectConfigBaseErrMsg = "inject configuration failed"

type RequiredConfigs struct {
	Store *config.Store
}

func injectConfig(container *Container) error {
	configuration, err := config.ReadYAMLConfig(container.fs)
	if err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, injectConfigBaseErrMsg, err)
	}
	if err = container.Validate.Struct(configuration); err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, injectConfigBaseErrMsg, err)
	}
	container.Store = config.NewStore(configuration)

	return nil
}
