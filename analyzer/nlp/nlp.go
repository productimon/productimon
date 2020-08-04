package nlp

const LABEL_UNKNOWN = "Unknown"

// guess the label for a given app
// we currently rely solely on wikipedia articles with software infobox
// TODO: add google knowledge api, wikidata, and potentially self-trained word embeddings
func GuessLabel(app string) string {
	if label := wikipediaLabel(app); label != "" {
		return label
	}
	return LABEL_UNKNOWN
}
