package scraper

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
)

const (
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36 Edg/119.0.0.0"
)

type ScraperOption func(*Scraper)

func MaxPages(pages int) ScraperOption {
	return func(s *Scraper) {
		s.MaxPages = pages
	}
}

func MaxRecipes(recipes int) ScraperOption {
	return func(s *Scraper) {
		s.MaxRecipes = recipes
	}
}

func Parallelism(parallelism int) ScraperOption {
	return func(s *Scraper) {
		s.Parallelism = parallelism
	}
}

func abortRequestHook(ctx context.Context) func(r *colly.Request) {
	return func(r *colly.Request) {
		select {
		case <-ctx.Done():
			r.Abort()
		default:
		}
	}
}

func splitTags(tagString string) []string {
	tags := strings.Split(tagString, ",")
	for i, v := range tags {
		tags[i] = strings.TrimSpace(v)
	}
	return tags
}

// Scraper represents a scraping job for a wprm site. It starts from the recipe index and fetches links to every recipe
type Scraper struct {
	MaxRecipes  int // Limit on max number of recipes to scrape (default: no limit)
	MaxPages    int // Limit on max number of index pages to scrape through (default: no limit)
	Parallelism int // Number of parallel link scrapes to run at the same time (default 10)
}

type Recipe struct {
	Rating      float64  // User rating
	NumRated    int      // Number of user ratings
	Name        string   // Recipe name
	Link        string   // URL to the print page
	Cuisine     []string // Cuisine tags
	Course      []string // Course tags
	Keywords    []string // Keyword tags
	Ingredients []string // Recipe ingredients
}

func New(options ...ScraperOption) (*Scraper, error) {
	scraper := &Scraper{
		MaxRecipes:  math.MaxInt,
		MaxPages:    math.MaxInt,
		Parallelism: 10,
	}
	for _, option := range options {
		option(scraper)
	}
	if err := scraper.validate(); err != nil {
		return nil, err
	}
	return scraper, nil
}

func (s *Scraper) createScraper(parallel bool) *colly.Collector {
	c := colly.NewCollector(
		colly.UserAgent(userAgent))
	if parallel {
		c.Async = true
		c.Limit(&colly.LimitRule{
			DomainGlob:  "*",
			Parallelism: s.Parallelism,
		})
	}
	return c
}

func (s *Scraper) ScrapeRecipeLink(ctx context.Context, u string) (Recipe, error) {
	ctx, cancel := context.WithCancel(ctx)
	var recipe Recipe
	var err error
	c := s.createScraper(false)
	printScraper := c.Clone()
	c.OnRequest(abortRequestHook(ctx))
	c.OnHTML("a.wprm-recipe-print", func(h *colly.HTMLElement) {
		select {
		case <-ctx.Done():
			return
		default:
		}
		recipeLink := h.Attr("href")
		printScraper.Visit(recipeLink)
	})
	c.OnError(func(r *colly.Response, err2 error) {
		err = err2
		cancel()
	})

	printScraper.OnHTML("html > body", func(h *colly.HTMLElement) {
		recipe.Link = h.Request.URL.String()
		h.ForEach("span.wprm-recipe-rating-average", func(i int, h *colly.HTMLElement) {
			var rating float64
			rating, err = strconv.ParseFloat(h.Text, 64)
			if err != nil {
				fmt.Println("Issue with link", h.Request.URL.String())
				cancel()
			}
			recipe.Rating = rating
		})
		h.ForEach("span.wprm-recipe-rating-count", func(i int, h *colly.HTMLElement) {
			var numRated int64
			numRated, err = strconv.ParseInt(h.Text, 10, 0)
			if err != nil {
				fmt.Println("Issue with link", h.Request.URL.String())
				cancel()
			}
			recipe.NumRated = int(numRated)
		})
		h.ForEach("span.wprm-recipe-course", func(i int, h *colly.HTMLElement) {
			recipe.Course = splitTags(h.Text)
		})
		h.ForEach("span.wprm-recipe-cuisine", func(i int, h *colly.HTMLElement) {
			recipe.Cuisine = splitTags(h.Text)
		})
		h.ForEach("span.wprm-recipe-keyword", func(i int, h *colly.HTMLElement) {
			recipe.Keywords = splitTags(h.Text)
		})
		h.ForEach("h2.wprm-recipe-name", func(i int, h *colly.HTMLElement) {
			recipe.Name = h.Text
		})
		h.ForEach("span.wprm-recipe-ingredient-name", func(i int, h *colly.HTMLElement) {
			recipe.Ingredients = append(recipe.Ingredients, strings.TrimSpace(strings.ToLower(h.Text)))
		})
	})
	printScraper.OnError(func(r *colly.Response, err2 error) {
		err = err2
		cancel()
	})
	c.Visit(u)

	return recipe, err
}

func (s *Scraper) validate() error {
	if s.MaxPages < 1 {
		return errors.New("max pages cannot be less than 1")
	}
	if s.MaxRecipes < 1 {
		return errors.New("max recipes cannot be less than 1")
	}
	if s.Parallelism < 1 {
		return errors.New("parallelism cannot be less than 1")
	}
	return nil
}

func (s *Scraper) ScrapeRecipe(ctx context.Context, u string) ([]Recipe, error) {
	ctx, cancel := context.WithCancel(ctx)
	var recipes []Recipe
	var err error
	defer cancel()
	c := s.createScraper(true)
	var recipeMutex sync.Mutex
	c.OnRequest(abortRequestHook(ctx))
	c.OnHTML("html body", func(h *colly.HTMLElement) {
		if len(h.DOM.Find("article").Nodes) == 0 {
			cancel()
			return
		}
		h.ForEachWithBreak("article", func(i int, h *colly.HTMLElement) bool {
			recipeLink := h.ChildAttr("a", "href")
			var recipe Recipe
			recipe, err = s.ScrapeRecipeLink(ctx, recipeLink)
			if err != nil {
				cancel()
				return false
			}
			if len(strings.TrimSpace(recipe.Name)) == 0 {
				return true
			}
			recipeMutex.Lock()
			defer recipeMutex.Unlock()
			if len(recipes) == s.MaxRecipes {
				cancel()
				return false
			}
			recipes = append(recipes, recipe)
			return true
		})
	})
	c.OnError(func(r *colly.Response, err2 error) {
		err = err2
		cancel()
	})
	u = strings.TrimSuffix(u, "/")
outerLoop:
	for i := 0; i < s.MaxPages; {
		for j := 0; j < s.Parallelism; j++ {
			if i == s.MaxPages-1 {
				break outerLoop
			}
			select {
			case <-ctx.Done():
				break outerLoop
			default:
				c.Visit(fmt.Sprintf("%s/page/%d", u, i+1))
			}
			i++
		}
		c.Wait()
	}
	c.Wait()
	if err != nil {
		return nil, err
	}
	return recipes, nil
}
