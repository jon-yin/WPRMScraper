package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/jon-yin/RecipeScraper/exporters"
	"github.com/jon-yin/RecipeScraper/scraper"
)

type exporter interface {
	ExportRecipes(ctx context.Context, recipes []scraper.Recipe) error
}

func usage(subUsages ...func()) func() {
	return func() {
		fmt.Println("Usage: main (json | html) -link [-max-recipes] [-max-pages] [-parallel] [-dest]")
		for _, usage := range subUsages {
			usage()
		}
	}
}

func main() {
	maxRecipes := flag.Int("max-recipes", -1, "maximum number of recipes to scrape")
	maxPages := flag.Int("max-pages", -1, "maximum number of pages to scrape from index")
	link := flag.String("link", "", "link to save recipes to")
	parallel := flag.Int("parallel", 10, "number of parallel threads for scraper")
	logLev := new(slog.LevelVar)
	logLev.Set(slog.LevelWarn)
	verboseOutput := flag.Bool("v", false, "log verbose output to terminal")
	jsonCmd := flag.NewFlagSet("json", flag.ExitOnError)
	jDFile := jsonCmd.String("dest", "recipes.json", "destination file for json file")
	htmlCmd := flag.NewFlagSet("html", flag.ExitOnError)
	hDDir := htmlCmd.String("dest", "recipes", "destination directory for recipes")
	flag.Parse()
	flag.Usage = usage(flag.PrintDefaults, htmlCmd.Usage, jsonCmd.Usage)
	flag.Usage()

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
	opts = append(opts, scraper.Parallelism(*parallel))
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
	switch os.Args[1] {
	case "json":
		jsonCmd.Parse(os.Args[1:])
		e = &exporters.JsonExporter{
			Filename: *jDFile,
		}
	case "html":
		htmlCmd.Parse(os.Args[1:])
		e, err = exporters.NewHtmlExporter(exporters.WithLogger(logger), exporters.WithParallelism(*parallel), exporters.WithDestDir(*hDDir))
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}

	default:
		logger.Error(fmt.Sprintf("invalid exporter specified, valid ones are %q, %q", "json", "html"))
		os.Exit(1)
	}
	if err := e.ExportRecipes(context.TODO(), recipes); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
