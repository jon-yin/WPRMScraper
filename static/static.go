package static

import (
	"embed"
	"os"
	"path/filepath"
)

const (
	jsFile  = "recipeViewer.js"
	cssFile = "style.css"
)

// base recipe file html
//
//go:embed recipe.html
var RecipeHtmlFile []byte

// css + js for site
//
//go:embed style.css recipeViewer.js
var cssJs embed.FS

func CopyStaticResources(dir string) error {
	css, err := cssJs.ReadFile(cssFile)
	if err != nil {
		return err
	}
	js, err := cssJs.ReadFile(jsFile)
	if err != nil {
		return err
	}
	resCss := filepath.Join(dir, cssFile)
	if err := os.WriteFile(resCss, css, 0666); err != nil {
		return err
	}
	resJs := filepath.Join(dir, jsFile)
	return os.WriteFile(resJs, js, 0666)
}
