package nlp

import (
	"regexp"
	"strings"
)

const (
	LabelProductimon = "Productimon"
	LabelEducation   = "Education"
	LabelGovernment  = "Government"
)

var (
	eduDomain = regexp.MustCompile(`\.edu(\.[a-z][a-z])?$`)
	govDomain = regexp.MustCompile(`\.gov(\.[a-z][a-z])?$`)
)

func listLabel(app string) string {
	app = strings.ToLower(app)
	if strings.Contains(app, "productimon") {
		return LabelProductimon
	}
	if eduDomain.MatchString(app) {
		return LabelEducation
	}
	if govDomain.MatchString(app) {
		return LabelGovernment
	}
	switch app {
	case "unknown":
		return LABEL_UNKNOWN
	}
	return ""
}
