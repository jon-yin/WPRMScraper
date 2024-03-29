package exporters

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"text/template"
	"time"

	"github.com/jon-yin/RecipeScraper/scraper"
	"github.com/jon-yin/RecipeScraper/static"
)

type ctxKey string

const (
	recipeSubDirectory          = "recipes"
	recipeDestFile              = "index.html"
	recipeKey            ctxKey = "recipe_key"
	recipeScriptTemplate        = `let recipeData = {{.Data}}`
)

type HtmlExporterOption func(*HtmlExporter)

func WithDestDir(dir string) HtmlExporterOption {
	return func(he *HtmlExporter) {
		he.DestDir = dir
	}
}

func WithParallelism(parallelism int) HtmlExporterOption {
	return func(he *HtmlExporter) {
		he.client.Parallelism = parallelism
	}
}

func WithLogger(logger *slog.Logger) HtmlExporterOption {
	return func(he *HtmlExporter) {
		he.Logger = logger
	}
}

// HTMLExporter creates HTML file with references to recipe links
type HtmlExporter struct {
	template *template.Template
	DestDir  string // Where to save index.html to, default is ./scrapeResults
	client   *MultiHttpClient
	Logger   *slog.Logger // Event logger
}

type recipeTemplateData struct {
	Data string
}

func NewHtmlExporter(opts ...HtmlExporterOption) (*HtmlExporter, error) {
	exporter := &HtmlExporter{
		DestDir: ".",
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
	recipesPath := filepath.Join(h.DestDir, recipeSubDirectory)
	err := os.MkdirAll(h.DestDir, 0666)
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
		h.Logger.Info("writing file", "link", res.Request.URL, "filename", recipe.ID+".html")
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
	staticDir := filepath.Join(h.DestDir, "static")
	if err := os.MkdirAll(staticDir, 0666); err != nil {
		return err
	}
	// Write json data
	err := h.saveRecipes(ctx, recipes)
	if err != nil {
		return err
	}
	scriptFile, err := os.OpenFile(path.Join(staticDir, "data.js"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
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
	// Write static data to destination
	if err = static.CopyStaticResources(staticDir); err != nil {
		return err
	}
	// Write index html file
	return os.WriteFile(path.Join(h.DestDir, recipeDestFile), static.RecipeHtmlFile, 0666)
}
