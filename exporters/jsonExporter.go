package exporters

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/jon-yin/RecipeScraper/scraper"
)

type JsonExporter struct {
	Filename string // Filename to export json file as
}

func (J JsonExporter) ExportRecipes(ctx context.Context, recipes []scraper.Recipe) error {
	absPath, err := filepath.Abs(J.Filename)
	if err != nil {
		return err
	}
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0644); err != nil {
		return err
	}
	data, err := json.Marshal(recipes)
	if err != nil {
		return err
	}
	return os.WriteFile(absPath, data, 0644)
}
