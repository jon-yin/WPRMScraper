package scraper

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
)

const (
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36 Edg/119.0.0.0"
)

var (
	NumericalRegexp = regexp.MustCompile(`^\d+$`)
	parallelism     = runtime.GOMAXPROCS(0)
)

func abortRequestHook(ctx context.Context, errPtr *error) func(r *colly.Request) {
	return func(r *colly.Request) {
		select {
		case <-ctx.Done():
			r.Abort()
		default:
			if *errPtr != nil {
				r.Abort()
			}
		}
	}
}

func ScrapeRecipeLinksFromIndex(ctx context.Context, u string, maxPages int) (recipes []string, err error) {
	if err != nil {
		return
	}
	if maxPages == 0 {
		maxPages = math.MaxInt32
	}
	c := colly.NewCollector(colly.Debugger(&debug.LogDebugger{}),
		colly.UserAgent(userAgent),
		colly.Async())
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: parallelism,
	})
	if err != nil {
		return
	}
	var recipeMutex sync.Mutex
	var finished bool
	c.OnRequest(abortRequestHook(ctx, &err))
	c.OnHTML("html body", func(h *colly.HTMLElement) {
		if len(h.DOM.Find("article").Nodes) == 0 {
			finished = true
		}
		h.ForEach("article", func(i int, h *colly.HTMLElement) {
			recipeLink := h.ChildAttr("a", "href")
			recipeMutex.Lock()
			recipes = append(recipes, recipeLink)
			defer recipeMutex.Unlock()
		})
	})
	c.OnError(func(r *colly.Response, err2 error) {
		err = err2
	})
	if !strings.HasSuffix(u, "/") {
		u = strings.TrimSuffix(u, "/")
	}
	for i := 0; i < maxPages; i++ {
		if finished {
			break
		}
		for j := 0; j < parallelism; j++ {
			c.Visit(fmt.Sprintf("%s/page/%d", u, i+1))
			i++
		}
		c.Wait()
	}
	c.Wait()
	if err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	return recipes, err
}

func ScrapePrintLinks(ctx context.Context, urls []string) ([]string, error) {
	var err error
	var mutex sync.Mutex

	recipePrintLinks := make([]string, 0, len(urls))
	c := colly.NewCollector(colly.Debugger(&debug.LogDebugger{}),
		colly.UserAgent(userAgent),
		colly.Async())
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: parallelism,
	})
	c.OnRequest(abortRequestHook(ctx, &err))
	c.OnHTML("a.wprm-recipe-print", func(h *colly.HTMLElement) {
		mutex.Lock()
		defer mutex.Unlock()
		recipePrintLinks = append(recipePrintLinks, h.Attr("href"))
	})
	c.OnError(func(r *colly.Response, err2 error) {
		err = err2
	})
	for i := 0; i < len(urls); i++ {
		c.Wait()
		for j := 0; j < parallelism; j++ {
			if i < len(urls) {
				c.Visit(urls[i])
				i++
			}
		}
	}
	c.Wait()
	if err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	return recipePrintLinks, nil
}
