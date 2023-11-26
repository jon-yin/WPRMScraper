package main

import (
	"context"
	"fmt"
	"os"

	scraper "github.com/jon-yin/RecipeScraper"
)

func main() {
	recipes, err := scraper.ScrapeRecipeLinksFromIndex(context.TODO(), "https://www.recipetineats.com/recipes", 200)
	if err != nil {
		fmt.Println("Error", err)
		os.Exit(1)
	}
	fmt.Println("Recipe length", len(recipes))
	// printLinks, err := scraper.ScrapePrintLinks(context.TODO(), recipes)
	// if err != nil {
	// 	fmt.Println("Error", err)
	// 	os.Exit(1)
	// }
	// fmt.Println("Print links", printLinks)
}
