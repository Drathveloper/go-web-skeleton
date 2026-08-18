package templates

import (
	"errors"
	"slices"

	"github.com/gin-gonic/gin"

	"github.com/Drathveloper/go-web-skeleton/common/i18n"
)

var ErrInvalidDictCall = errors.New("invalid dict call")
var ErrDictKeysType = errors.New("dict keys must be strings")

const paramsGroupSize = 2

func registerDictFunc(engine *gin.Engine) {
	engine.FuncMap["dict"] = dictFunc
}

func dictFunc(values ...any) (map[string]any, error) {
	if len(values)%2 != 0 {
		return nil, ErrInvalidDictCall
	}

	dict := make(map[string]any, len(values)/paramsGroupSize)
	for i := 0; i < len(values); i += paramsGroupSize {
		key, ok := values[i].(string)
		if !ok {
			return nil, ErrDictKeysType
		}
		dict[key] = values[i+1]
	}

	return dict, nil
}

func registerContainsFunc(engine *gin.Engine) {
	engine.FuncMap["contains"] = slices.Contains[[]string, string]
}

func registerSliceFunc(engine *gin.Engine) {
	engine.FuncMap["slice"] = sliceFunc
}

func sliceFunc(values ...any) []any {
	return values
}

func registerLocalizeFunc(engine *gin.Engine) {
	engine.FuncMap["localize"] = i18n.LocalizeMessage
}
