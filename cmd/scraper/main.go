package main

import (
	"context"
	"fmt"
	"os"

	scraper "github.com/jon-yin/RecipeScraper"
)

func main() {
	s, err := scraper.New(scraper.MaxRecipes(50))
	if err != nil {
		fmt.Println("Error", err)
		os.Exit(1)
	}
	recipesLinks, err := s.ScrapeRecipeLinksFromIndex(context.TODO(), "https://www.recipetineats.com/recipes")
	if err != nil {
		fmt.Println("Error", err)
		os.Exit(1)
	}
	printLinks, err := s.ScrapeRecipePrintLinks(context.TODO(), recipesLinks)
	if err != nil {
		fmt.Println("Error", err)
		os.Exit(1)
	}
	recipes, err := s.ScrapeRecipeInfoFromPrintLinks(context.TODO(), printLinks)
	if err != nil {
		fmt.Println("Error", err)
		os.Exit(1)
	}
	fmt.Println("Number of recipes", len(recipes))
	fmt.Println("First recipe", recipes[0])
	// printLinks, err := scraper.ScrapePrintLinks(context.TODO(), recipes)
	// if err != nil {
	// 	fmt.Println("Error", err)
	// 	os.Exit(1)
	// }
	// fmt.Println("Print links", printLinks)
}
