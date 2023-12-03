package exporters

import (
	"errors"
	"html/template"
	"io"
	"os"
	"path"
	"runtime"

	scraper "github.com/jon-yin/RecipeScraper"
)

const TmplPath = "recipe.tmpl"
const secondPath = `C:\Users\Jonathan Yin\Documents\Projects\RecipeScraper\exporters\recipe.tmpl`

// HTMLExporter creates HTML file with references to recipe links
type HtmlExporter struct {
	template *template.Template
}

type recipeTemplateData struct {
}

func NewHtmlExporter() (*HtmlExporter, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return nil, errors.New("could not load template")
	}
	exporter := &HtmlExporter{}
	currentDir := path.Dir(file)
	exporter.template = template.Must(template.ParseFiles(path.Join(currentDir, TmplPath)))
	return exporter, nil
}

func (h *HtmlExporter) ExportRecipesToFile(fileName string, recipes []scraper.Recipe) error {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	return h.ExportRecipes(file, recipes)
}

func (h *HtmlExporter) ExportRecipes(writer io.Writer, recipes []scraper.Recipe) error {
	return h.template.Execute(writer, recipes)
}
