package static

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
)

//go:embed files/**
var staticFS embed.FS

func InitializeStaticFiles(router *gin.Engine) error {
	publicFS, err := fs.Sub(staticFS, "files")
	if err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, "initialize static files failed", err)
	}
	router.StaticFS("assets", http.FS(publicFS))
	return nil
}
