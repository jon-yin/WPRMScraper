package scraper

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var nonAlphanumericalRegexp = regexp.MustCompile("[^a-zA-Z0-9]")

func canonizeTags(tags []string) []string {
	if len(tags) == 0 {
		return nil
	}
	cTags := make([]string, len(tags))
	for i, v := range tags {
		t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
		noDiacritics, _, _ := transform.String(t, v)
		alphaNumStr := nonAlphanumericalRegexp.ReplaceAllString(noDiacritics, "")
		cTags[i] = strings.ToLower(alphaNumStr)
	}
	return cTags
}

func hasTags(allTags []string, desiredTags []string) bool {
	if len(desiredTags) == 0 {
		return true
	}
	if len(allTags) == 0 {
		return false
	}
	for _, desTag := range desiredTags {
		hasTag := false
		for _, tag := range allTags {
			if strings.Contains(tag, desTag) {
				hasTag = true
				break
			}
		}
		if !hasTag {
			return false
		}
	}
	return true
}

// RecipeFilter filters recipes based on certain criteria. A default filter allows everything.
type RecipeFilter struct {
	MinRating   float64  // Recipe's minimum rating
	MinNumRated int      // Minimum number of users to rate recipe
	Ingredients []string // Recipe must contain these ingredients
	Keywords    []string // Recipe must contain these tags
	Course      []string // Recipe must contain these course tags
}

func (f RecipeFilter) Filter(r Recipe) bool {
	if r.Rating < f.MinRating {
		return false
	}
	if r.NumRated < f.MinNumRated {
		return false
	}
	if len(f.Ingredients) > 0 {
		wantedIngredients := canonizeTags(f.Ingredients)
		gotIngredients := canonizeTags(r.Ingredients)
		if !hasTags(gotIngredients, wantedIngredients) {
			return false
		}
	}
	if len(f.Keywords) > 0 {
		wantedKeywords := canonizeTags(f.Keywords)
		gotKeywords := canonizeTags(r.Keywords)
		if !hasTags(gotKeywords, wantedKeywords) {
			return false
		}
	}
	if len(f.Course) > 0 {
		wantedCourses := canonizeTags(f.Course)
		gotCourses := canonizeTags(r.Course)
		if !hasTags(gotCourses, wantedCourses) {
			return false
		}
	}
	return true
}
