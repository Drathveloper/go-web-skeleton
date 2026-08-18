package config

import (
	"fmt"
	"io"
	"io/fs"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
)

const readConfigFileBaseErrMsg = "read config file failed"

func readFSConfig(fs fs.FS, path string) ([]byte, error) {
	f, err := fs.Open(path)
	if err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, readConfigFileBaseErrMsg, err)
	}
	content, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, readConfigFileBaseErrMsg, err)
	}
	return content, nil
}
