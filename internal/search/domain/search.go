package domain

import (
	"slices"
	"strings"
)

const (
	VersionParam     = "version"
	RelevanceParam   = "relevance"
	DefaultRelevance = 0.5
)

// GeoJSON properties in search response
const (
	PropCollectionID = "collection_id"
	PropGeomType     = "collection_geometry_type"
	PropHref         = "href"
)

// SearchQuery based on parsed search terms/words.
type SearchQuery struct {
	words           []string
	withoutSynonyms map[string]struct{}
	withSynonyms    map[string][]string
}

func NewSearchQuery(words []string, withoutSynonyms map[string]struct{}, withSynonyms map[string][]string) *SearchQuery {
	return &SearchQuery{
		words,
		withoutSynonyms,
		withSynonyms}
}

func (q *SearchQuery) ToWildcardQuery() string {
	return q.toString(true, true)
}

func (q *SearchQuery) ToExactMatchQuery(useSynonyms bool) string {
	return q.toString(false, useSynonyms)
}

func (q *SearchQuery) toString(useWildcard bool, useSynonyms bool) string {
	wildcard := ""
	if useWildcard {
		wildcard = ":*"
	}

	sb := &strings.Builder{}
	for i, word := range q.words {
		if i > 0 {
			sb.WriteString(" & ")
		}
		if _, ok := q.withoutSynonyms[word]; ok {
			sb.WriteString(word)
			sb.WriteString(wildcard)
		} else if synonyms, ok := q.withSynonyms[word]; ok {
			slices.Sort(synonyms)
			sb.WriteByte('(')
			sb.WriteString(word)
			sb.WriteString(wildcard)
			if useSynonyms {
				for _, synonym := range synonyms {
					sb.WriteString(" | ")
					sb.WriteString(synonym)
					sb.WriteString(wildcard)
				}
			}
			sb.WriteByte(')')
		}
	}
	return sb.String()
}
