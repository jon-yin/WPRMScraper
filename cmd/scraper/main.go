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

type CommonFlags struct {
	MaxRecipes int
	MaxPages   int
	Link       string
	Parallel   int
	Verbose    bool
}

func CreateCommonFlagSet() (*flag.FlagSet, *CommonFlags) {
	cFlags := &CommonFlags{}
	flagSet := flag.NewFlagSet("scraper", flag.ExitOnError)
	flagSet.IntVar(&cFlags.MaxRecipes, "max-recipes", -1, "maximum number of recipes to scrape")
	flagSet.IntVar(&cFlags.MaxPages, "max-pages", -1, "maximum number of pages to scrape from index")
	flagSet.StringVar(&cFlags.Link, "link", "", "link to save recipes to")
	flagSet.IntVar(&cFlags.Parallel, "parallel", 10, "number of parallel threads for scraper")
	flagSet.BoolVar(&cFlags.Verbose, "v", false, "log verbose output to terminal")
	return flagSet, cFlags
}

func subcommandUsage(subCommand string, usage func()) func() {
	return func() {
		fmt.Printf("Usage: scraper %s\n", subCommand)
		usage()
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: scraper (json | html)")
		return
	}
	flagSet, cFlags := CreateCommonFlagSet()
	programArgs := os.Args[1:]
	var dest string
	logLev := new(slog.LevelVar)
	logLev.Set(slog.LevelWarn)
	// set up logger
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLev,
	}))
	// set up exporter and parse args
	var e exporter
	switch programArgs[0] {
	case "json":
		flagSet.StringVar(&dest, "dest", "recipes.json", "location of resulting json file")
		flagSet.Usage = subcommandUsage("json", flagSet.Usage)
		flagSet.Parse(programArgs[1:])
		e = &exporters.JsonExporter{
			Filename: dest,
		}
	case "html":
		flagSet.StringVar(&dest, "dest", "./scrapeResults", "directory of result files")
		flagSet.Usage = subcommandUsage("html", flagSet.Usage)
		flagSet.Parse(programArgs[1:])
		var err error
		e, err = exporters.NewHtmlExporter(exporters.WithLogger(logger), exporters.WithParallelism(cFlags.Parallel), exporters.WithDestDir(dest))
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
	default:
		fmt.Println("Usage: scraper (json | html)", "unknown subcommand given")
		os.Exit(1)
	}
	// set up scraper
	var opts []scraper.Option
	if cFlags.MaxRecipes != -1 {
		opts = append(opts, scraper.MaxRecipes(cFlags.MaxRecipes))
	}
	if cFlags.MaxPages != -1 {
		opts = append(opts, scraper.MaxPages(cFlags.MaxPages))
	}
	opts = append(opts, scraper.Parallelism(cFlags.Parallel))
	if cFlags.Verbose {
		logLev.Set(slog.LevelDebug)
	}
	opts = append(opts, scraper.Logger(logger))
	if len(strings.TrimSpace(cFlags.Link)) == 0 {
		logger.Error("link cannot be empty")
		os.Exit(1)
	}
	s, err := scraper.New(opts...)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	recipes, err := s.ScrapeRecipeIndex(context.TODO(), cFlags.Link)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	if err := e.ExportRecipes(context.TODO(), recipes); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
