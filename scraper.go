package scraper

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
	"github.com/google/uuid"
	"github.com/jon-yin/RecipeScraper/logger"
)

const (
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36 Edg/119.0.0.0"
)

var (
	errComplete = errors.New("finished scraping")
)

type Option func(*Scraper)

func MaxPages(pages int) Option {
	return func(s *Scraper) {
		s.MaxPages = pages
	}
}

func MaxRecipes(recipes int) Option {
	return func(s *Scraper) {
		s.MaxRecipes = recipes
	}
}

func Parallelism(parallelism int) Option {
	return func(s *Scraper) {
		s.Parallelism = parallelism
	}
}

func WithLogger(logger *logger.Logger) Option {
	return func(s *Scraper) {
		s.Logger = logger
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

func logHook(logger *logger.Logger) func(r *colly.Request) {
	return func(r *colly.Request) {
		logger.Info("Visited site %s", r.URL)
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
	MaxRecipes  int            // Limit on max number of recipes to scrape (default: no limit)
	MaxPages    int            // Limit on max number of index pages to scrape through (default: no limit)
	Parallelism int            // Number of parallel link scrapes to run at the same time (default 10)
	Logger      *logger.Logger // If set, will log scraping events
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
	ID          string   // a unique id for this recipe
}

func New(options ...Option) (*Scraper, error) {
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
		colly.UserAgent(userAgent),
	)
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
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	var recipe Recipe
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
	c.OnError(func(r *colly.Response, err error) {
		if r.StatusCode == http.StatusNotFound {
			cancel(errComplete)
			return
		}
		cancel(err)
	})

	printScraper.OnHTML("html > body", func(h *colly.HTMLElement) {
		recipe.ID = uuid.NewString()
		recipe.Link = h.Request.URL.String()
		h.ForEach("span.wprm-recipe-rating-average", func(i int, h *colly.HTMLElement) {
			rating, err := strconv.ParseFloat(h.Text, 64)
			if err != nil {
				cancel(err)
			}
			recipe.Rating = rating
		})
		h.ForEach("span.wprm-recipe-rating-count", func(i int, h *colly.HTMLElement) {
			numRated, err := strconv.ParseInt(h.Text, 10, 0)
			if err != nil {
				fmt.Println("Issue with link", h.Request.URL.String())
				cancel(err)
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
	printScraper.OnError(func(r *colly.Response, err error) {
		cancel(err)
	})
	c.Visit(u)
	err := context.Cause(ctx)
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
	if s.Logger == nil {
		s.Logger = logger.NewLogger(nil)
	}
	return nil
}

func (s *Scraper) validateIndex(u string) error {
	hasArticles := false
	c := s.createScraper(false)
	c.OnHTML("html body", func(h *colly.HTMLElement) {
		if len(h.DOM.Find("article").Nodes) != 0 {
			hasArticles = true
		}
	})
	c.Visit(u)
	if !hasArticles {
		return errors.New("invalid index no articles found")
	}
	return nil
}

func (s *Scraper) ScrapeRecipeIndex(ctx context.Context, u string) ([]Recipe, error) {
	if err := s.validateIndex(u); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancelCause(ctx)
	var recipes []Recipe
	defer cancel(nil)
	c := s.createScraper(true)
	var recipeMutex sync.Mutex
	c.OnRequest(abortRequestHook(ctx))
	c.OnRequest(logHook(s.Logger))
	c.OnHTML("html body", func(h *colly.HTMLElement) {
		h.ForEachWithBreak("article", func(i int, h *colly.HTMLElement) bool {
			recipeLink := h.ChildAttr("a", "href")
			recipe, err := s.ScrapeRecipeLink(ctx, recipeLink)
			if err != nil {
				cancel(err)
				return false
			}
			if len(strings.TrimSpace(recipe.Name)) == 0 {
				return true
			}
			recipeMutex.Lock()
			defer recipeMutex.Unlock()
			if len(recipes) == s.MaxRecipes {
				cancel(errComplete)
				return false
			}
			s.Logger.Info("Scraped recipe %s", recipe.Name)
			recipes = append(recipes, recipe)
			return true
		})
	})
	c.OnError(func(r *colly.Response, err error) {
		if r.StatusCode == http.StatusNotFound {
			cancel(errComplete)
		}
		cancel(err)
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
	err := context.Cause(ctx)
	if errors.Is(err, errComplete) {
		err = nil
	}
	if err != nil {
		return nil, err
	}
	return recipes, nil
}
