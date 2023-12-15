package exporters

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path"
	"runtime"
	"text/template"
	"time"

	scraper "github.com/jon-yin/RecipeScraper"
)

type ctxKey string

const (
	recipeDestFile              = "index.html"
	recipeSrcFile               = "recipe.html"
	recipeKey            ctxKey = "recipe_key"
	recipeScriptTemplate        = `let recipeData = {{.Data}}`
)

type HtmlExporterOption func(*HtmlExporter)

func WithRecipeDir(dir string) HtmlExporterOption {
	return func(h *HtmlExporter) {
		h.RecipeDir = path.Dir(dir)
	}
}

func WithDestDir(dir string) HtmlExporterOption {
	return func(he *HtmlExporter) {
		he.DestDir = path.Dir(dir)
	}
}

func WithParallelism(parallelism int) HtmlExporterOption {
	return func(he *HtmlExporter) {
		he.client.Parallelism = parallelism
	}
}

// HTMLExporter creates HTML file with references to recipe links
type HtmlExporter struct {
	template  *template.Template
	RecipeDir string // Directory name of recipes; this is relative to DestDir, default is "recipes"
	DestDir   string // Where to save index.html to, default is current directory
	client    *MultiHttpClient
}

type recipeTemplateData struct {
	Data string
}

func NewHtmlExporter(opts ...HtmlExporterOption) (*HtmlExporter, error) {
	exporter := &HtmlExporter{
		RecipeDir: "recipes",
		DestDir:   ".",
		client: &MultiHttpClient{
			Parallelism: 10,
			Client: http.Client{
				Timeout: 10 * time.Second,
			},
		},
	}
	for _, v := range opts {
		v(exporter)
	}
	exporter.template = template.Must(template.New("recipeData").Parse(recipeScriptTemplate))
	return exporter, nil
}

func (h *HtmlExporter) saveRecipes(ctx context.Context, recipes []scraper.Recipe) error {
	basePath := h.DestDir
	recipesPath := path.Join(basePath, h.RecipeDir)
	err := os.MkdirAll(basePath, 0666)
	if err != nil {
		return err
	}
	err = os.MkdirAll(recipesPath, 0666)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	for i, v := range recipes {
		ctx = context.WithValue(ctx, recipeKey, i)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.Link, nil)
		if err != nil {
			return err
		}
		h.client.QueueRequest(req)
	}
	h.client.OnResponse(func(res *http.Response) {
		defer res.Body.Close()
		recipeIndex := res.Request.Context().Value(recipeKey).(int)
		recipe := recipes[recipeIndex]
		file, err := os.OpenFile(path.Join(recipesPath, recipe.ID+".html"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
		if err != nil {
			cancel(err)
		}
		defer file.Close()
		_, err = io.Copy(file, res.Body)
		if err != nil {
			cancel(err)
		}
	})
	return h.client.Execute(ctx)
}

func (h *HtmlExporter) makeTemplateData(recipes []scraper.Recipe) (recipeTemplateData, error) {
	templateData := recipeTemplateData{}
	jsonData, err := json.Marshal(recipes)
	if err != nil {
		return recipeTemplateData{}, err
	}
	templateData.Data = string(jsonData)
	return templateData, nil
}

func (h *HtmlExporter) ExportRecipes(ctx context.Context, recipes []scraper.Recipe) error {
	err := h.saveRecipes(ctx, recipes)
	if err != nil {
		return err
	}
	scriptFile, err := os.OpenFile(path.Join(h.DestDir, "data.js"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer scriptFile.Close()
	tmpData, err := h.makeTemplateData(recipes)
	if err != nil {
		return err
	}
	err = h.template.Execute(scriptFile, tmpData)
	if err != nil {
		return err
	}
	_, currentFile, _, _ := runtime.Caller(0)
	recipeFilePath := path.Join(path.Dir(currentFile), recipeSrcFile)
	recipeFile, err := os.Open(recipeFilePath)
	if err != nil {
		return err
	}
	defer recipeFile.Close()
	copiedFile, err := os.OpenFile(path.Join(h.DestDir, recipeDestFile), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer copiedFile.Close()
	_, err = io.Copy(copiedFile, recipeFile)
	return err
}
