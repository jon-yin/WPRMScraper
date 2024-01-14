package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"strings"

	"github.com/jon-yin/RecipeScraper/exporters"
	"github.com/jon-yin/RecipeScraper/scraper"
)

type exporter interface {
	ExportRecipes(ctx context.Context, recipes []scraper.Recipe) error
}

func main() {
	maxRecipes := flag.Int("max-recipes", -1, "maximum number of recipes to scrape")
	maxPages := flag.Int("max-pages", -1, "maximum number of pages to scrape from index")
	exportType := flag.String("exporter", "html", "recipe exporter format (html, json)")
	htmlExportDir := flag.String("out", "./recipes", "directory to export recipe files to")
	link := flag.String("link", "", "link to save recipes to")
	logLev := new(slog.LevelVar)
	logLev.Set(slog.LevelWarn)
	verboseOutput := flag.Bool("v", false, "log verbose output to terminal")
	flag.Parse()
	var opts []scraper.Option
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLev,
	}))
	if *maxRecipes != -1 {
		opts = append(opts, scraper.MaxRecipes(*maxRecipes))
	}
	if *maxPages != -1 {
		opts = append(opts, scraper.MaxPages(*maxPages))
	}
	if *verboseOutput {
		logLev.Set(slog.LevelDebug)
	}
	opts = append(opts, scraper.Logger(logger))
	if len(strings.TrimSpace(*link)) == 0 {
		logger.Error("link cannot be empty")
		os.Exit(1)
	}
	s, err := scraper.New(opts...)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	recipes, err := s.ScrapeRecipeIndex(context.TODO(), *link)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	var e exporter
	switch *exportType {
	case "json":
		e = &exporters.JsonExporter{}
	}
}
