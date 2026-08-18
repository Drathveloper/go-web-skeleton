package config

import (
	"bytes"
	"fmt"
	"io/fs"

	"gopkg.in/yaml.v3"

	"github.com/Drathveloper/go-web-skeleton/common/config/model"
	"github.com/Drathveloper/go-web-skeleton/common/constants"
)

const readYAMLConfigBaseErrMsg = "read yaml config failed"

const yamlConfigFilePath = "config/application.yaml"

func ReadYAMLConfig(fs fs.FS) (*model.Configuration, error) {
	content, err := readFSConfig(fs, yamlConfigFilePath)
	if err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, readYAMLConfigBaseErrMsg, err)
	}
	var config model.Configuration
	// KnownFields makes an unrecognised key a hard error. Non-strict decoding
	// is how a misspelled `max_connection_lifetime` sat in the source's
	// production config being silently ignored: the pool never got the
	// lifetime it was configured with and nothing said so.
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	decoder.KnownFields(true)
	if err = decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf(constants.DefaultWrappedErrorTemplate, readYAMLConfigBaseErrMsg, err)
	}
	return &config, nil
}
