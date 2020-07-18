package nlp

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	apiURL = "http://en.wikipedia.org/w/api.php"
)

var pid_regex *regexp.Regexp
var genre_regex *regexp.Regexp

func init() {
	pid_regex = regexp.MustCompile(`"pageid":(\d+),`)
	genre_regex = regexp.MustCompile(`genre\s*=[^\[\]]*\[\[([^\[\]]*\|)?([^\[\]]*)\]\]`)
}

func wikipediaLabel(app string) string {
	v := url.Values{
		"action":   {"query"},
		"list":     {"search"},
		"format":   {"json"},
		"srsearch": {strings.ReplaceAll(app, "\"", "\\\"") + " hastemplate:infobox_software"},
		"srlimit":  {"1"},
	}

	res, err := http.Get(apiURL + "?" + v.Encode())
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
	pids := pid_regex.FindSubmatch(result)
	if len(pids) < 2 {
		return LABEL_UNKNOWN
	}

	v = url.Values{
		"action":  {"query"},
		"prop":    {"revisions"},
		"rvprop":  {"content"},
		"pageids": {string(pids[1])},
		"rvslots": {"main"},
		"format":  {"json"},
	}

	res, err = http.Get(apiURL + "?" + v.Encode())
	if err != nil {
		log.Println(err)
		return LABEL_UNKNOWN
	}
	result, err = ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Println(err)
		return LABEL_UNKNOWN
	}

	// json parsing is slower than regex
	genres := genre_regex.FindSubmatch(result)
	if len(genres) < 3 {
		return LABEL_UNKNOWN
	}
	return string(genres[2])
}
