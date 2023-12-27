package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	scraper "github.com/jon-yin/RecipeScraper"
	"github.com/jon-yin/RecipeScraper/exporters"
)

func main() {
	maxRecipes := flag.Int("max-recipes", -1, "maximum number of recipes to scrape")
	maxPages := flag.Int("max-pages", -1, "maximum number of pages to scrape from index")
	htmlExportDir := flag.String("out", "./recipes", "directory to export recipe files to")
	flag.Parse()
	link := flag.Arg(0)
	var opts []scraper.Option
	if *maxRecipes != -1 {
		opts = append(opts, scraper.MaxRecipes(*maxRecipes))
	}
	if *maxPages != -1 {
		opts = append(opts, scraper.MaxPages(*maxPages))
	}
	if len(strings.TrimSpace(link)) == 0 {
		fmt.Fprintln(os.Stderr, "link cannot be empty")
		os.Exit(1)
	}
	s, err := scraper.New(opts...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	recipes, err := s.ScrapeRecipeIndex(context.TODO(), link)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	exporter, err := exporters.NewHtmlExporter(exporters.WithDestDir(*htmlExportDir))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	err = exporter.ExportRecipes(context.TODO(), recipes)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
