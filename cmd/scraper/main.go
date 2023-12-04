package main

import (
	"context"
	"fmt"
	"os"

	scraper "github.com/jon-yin/RecipeScraper"
	"github.com/jon-yin/RecipeScraper/exporters"
)

func main() {
	s, err := scraper.New(scraper.MaxRecipes(50))
	if err != nil {
		fmt.Println("Error", err)
		os.Exit(1)
	}
	recipes, err := s.ScrapeRecipe(context.TODO(), "https://www.recipetineats.com/recipes")
	if err != nil {
		fmt.Println("Error", err)
		os.Exit(1)
	}
	exporter, err := exporters.NewHtmlExporter()
	if err != nil {
		fmt.Println("Error", err)
		os.Exit(1)
	}
	err = exporter.ExportRecipes(context.TODO(), recipes)
	if err != nil {
		fmt.Println("Error", err)
		os.Exit(1)
	}
}
