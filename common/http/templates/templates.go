package templates

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"

	"github.com/Drathveloper/go-web-skeleton/common/constants"
)

const (
	baseTemplateFile      = "files/base.gohtml"
	baseErrorTemplateFile = "files/base_error.gohtml"

	layoutsDir   = "files/layouts"
	pagesDir     = "files/pages"
	fragmentsDir = "files/fragments"

	templateExtension = ".gohtml"

	initializeTemplateRendererErrMsg = "initialize template renderer failed"
)

//go:embed all:files
var templatesFS embed.FS

// InitializeTemplateRenderer registers every embedded template by convention,
// never by name:
//
//	files/pages/<group>/<page>.gohtml     -> "<group>/<page>", composed with
//	                                         base.gohtml plus every layout
//	files/fragments/<group>/<frag>.gohtml -> "fragments/<group>/<frag>", standalone
//
// Dropping a file in the right directory is all it takes to register it, which
// is what makes the scaffold generator possible.
func InitializeTemplateRenderer(engine *gin.Engine) error {
	registerFuncs(engine)
	renderer := multitemplate.NewRenderer()
	renderer.AddFromFSFuncs("error", engine.FuncMap, templatesFS, baseErrorTemplateFile)
	baseFiles := make([]string, 0, 1)
	baseFiles = append(baseFiles, baseTemplateFile)
	if err := walkTemplateDir(layoutsDir, appendFoundTemplatesToList(&baseFiles)); err != nil {
		return err
	}
	if err := walkTemplateDir(pagesDir, appendPagesToRenderer(renderer, engine.FuncMap, baseFiles)); err != nil {
		return err
	}
	if err := walkTemplateDir(fragmentsDir, appendFragmentsToRenderer(renderer, engine.FuncMap)); err != nil {
		return err
	}
	engine.HTMLRender = renderer
	return nil
}

// walkTemplateDir walks dir inside the embedded filesystem. A missing directory
// is not an error: layouts, pages and fragments are filled in by the generator,
// and a freshly scaffolded project legitimately has none of them yet.
func walkTemplateDir(dir string, walkFunc fs.WalkDirFunc) error {
	if _, err := fs.Stat(templatesFS, dir); errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err := fs.WalkDir(templatesFS, dir, walkFunc); err != nil {
		return fmt.Errorf(constants.DefaultWrappedErrorTemplate, initializeTemplateRendererErrMsg, err)
	}
	return nil
}

func registerFuncs(engine *gin.Engine) {
	registerDictFunc(engine)
	registerContainsFunc(engine)
	registerSliceFunc(engine)
	registerLocalizeFunc(engine)
}

func appendFoundTemplatesToList(baseFiles *[]string) fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != templateExtension {
			return nil
		}
		*baseFiles = append(*baseFiles, path)
		return nil
	}
}

func appendPagesToRenderer(renderer multitemplate.Renderer, funcs template.FuncMap, templates []string) fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != templateExtension {
			return nil
		}

		pageFiles := slices.Clone(templates)
		pageFiles = append(pageFiles, path)

		relativePath := strings.TrimPrefix(path, pagesDir+"/")
		templateName := strings.TrimSuffix(relativePath, templateExtension)

		renderer.AddFromFSFuncs(templateName, funcs, templatesFS, pageFiles...)
		return nil
	}
}

func appendFragmentsToRenderer(renderer multitemplate.Renderer, funcs template.FuncMap) fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != templateExtension {
			return nil
		}

		relativePath := strings.TrimPrefix(path, "files/")
		templateName := strings.TrimSuffix(relativePath, templateExtension)

		renderer.AddFromFSFuncs(templateName, funcs, templatesFS, path)
		return nil
	}
}
