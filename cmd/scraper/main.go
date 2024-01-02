package main

import (
	"context"
	"flag"
	"strings"

	scraper "github.com/jon-yin/RecipeScraper"
	"github.com/jon-yin/RecipeScraper/exporters"
	"github.com/jon-yin/RecipeScraper/logger"
)

func main() {
	maxRecipes := flag.Int("max-recipes", -1, "maximum number of recipes to scrape")
	maxPages := flag.Int("max-pages", -1, "maximum number of pages to scrape from index")
	htmlExportDir := flag.String("out", "./recipes", "directory to export recipe files to")
	verboseOutput := flag.Bool("v", false, "log verbose output to terminal")
	flag.Parse()
	link := flag.Arg(0)
	var opts []scraper.Option
	var scLog logger.Logger
	if *verboseOutput {
		scLog = *logger.NewLogger(logger.VerboseLoggerOpts)
	} else {
		scLog = *logger.NewLogger(logger.DefaultLoggerOpts)
	}
	if *maxRecipes != -1 {
		opts = append(opts, scraper.MaxRecipes(*maxRecipes))
	}
	if *maxPages != -1 {
		opts = append(opts, scraper.MaxPages(*maxPages))
	}
	if len(strings.TrimSpace(link)) == 0 {
		scLog.Fatal("link cannot be empty")
	}
	s, err := scraper.New(opts...)
	if err != nil {
		scLog.Fatal(err)
	}
	recipes, err := s.ScrapeRecipeIndex(context.TODO(), link)
	if err != nil {
		scLog.Fatal(err)
	}
	exporter, err := exporters.NewHtmlExporter(exporters.WithDestDir(*htmlExportDir))
	if err != nil {
		scLog.Fatal(err)
	}
	err = exporter.ExportRecipes(context.TODO(), recipes)
	if err != nil {
		scLog.Fatal(err)
	}
}
