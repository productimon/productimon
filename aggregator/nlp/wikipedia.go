package nlp

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/agnivade/levenshtein"
)

const (
	apiURL = "http://en.wikipedia.org/w/api.php"
)

type WikiPage struct {
	Title         string `json:'title'`
	RedirectTitle string `json:'redirecttitle'`
	PageID        int64  `json:'pageid'`
}

type WikiSearchResult struct {
	Query struct {
		Search []WikiPage `json:'search'`
	} `json:'query'`
}

var pid_regex *regexp.Regexp
var genre_regex *regexp.Regexp

func init() {
	genre_regex = regexp.MustCompile(`genre\s*=\s*([^\[\]]*?(\[\[([^\[\]]*\|)?([^\[\]]*)\]\].*?)?)\s*\\n\s*\|`)
}

func computeDistance(x, y string) float32 {
	x = strings.ToLower(x)
	y = strings.ToLower(y)
	if strings.Contains(y, x) || strings.Contains(x, y) {
		return 1
	}
	return float32(levenshtein.ComputeDistance(x, y))
}

func (page WikiPage) computeDistance(app string) float32 {
	title := page.Title
	if len(page.RedirectTitle) > 0 {
		title = page.RedirectTitle
	}
	return computeDistance(app, title)
}

func wikipediaLabel(app string) string {
	v := url.Values{
		"action":   {"query"},
		"list":     {"search"},
		"format":   {"json"},
		"srsearch": {app + " hastemplate:infobox_software"},
		"srlimit":  {"5"},
		"srprop":   {"redirecttitle"},
	}

	res, err := http.Get(apiURL + "?" + v.Encode())
	if err != nil {
		log.Println(err)
		return LABEL_UNKNOWN
	}
	defer res.Body.Close()
	var sr WikiSearchResult
	if err = json.NewDecoder(res.Body).Decode(&sr); err != nil {
		log.Println(err)
		return LABEL_UNKNOWN
	}

	if len(sr.Query.Search) == 0 {
		return LABEL_UNKNOWN
	}

	// we fetched top 5 results from wikipedia
	// and perform our simple ranking algorithm based on normalized levenshtein distance
	// in matched title, with weightings from Wikipedia's ranking.
	bestMatch := 0
	bestDistance := sr.Query.Search[0].computeDistance(app)
	for idx, page := range sr.Query.Search[1:] {
		if pd := page.computeDistance(app) * (1 + 0.3*float32(idx+1)); pd < bestDistance {
			bestDistance = pd
			bestMatch = idx + 1
		}
	}

	v = url.Values{
		"action":  {"query"},
		"prop":    {"revisions"},
		"rvprop":  {"content"},
		"pageids": {strconv.FormatInt(sr.Query.Search[bestMatch].PageID, 10)},
		"rvslots": {"main"},
		"format":  {"json"},
	}

	res, err = http.Get(apiURL + "?" + v.Encode())
	if err != nil {
		log.Println(err)
		return LABEL_UNKNOWN
	}
	result, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Println(err)
		return LABEL_UNKNOWN
	}

	// json parsing is slower than regex
	// after parsing json, wikipedia uses a weird non-standard format
	// for infobox, that's hard to parse
	genres := genre_regex.FindSubmatch(result)
	if len(genres) < 5 {
		return LABEL_UNKNOWN
	}
	if len(genres[4]) > 0 {
		return string(genres[4])
	}
	if len(genres[1]) > 0 {
		return string(genres[1])
	}
	return LABEL_UNKNOWN
}
