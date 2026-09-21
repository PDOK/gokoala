package domain

import (
	"slices"
	"strings"
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

// returns the full search term with SQL wildcard at the end
func (q *SearchQuery) ToUntokenizedQuery() string {
	return strings.Join(q.words, " ")
}

func (q *SearchQuery) toString(useWildcard bool, useSynonyms bool) string {
	wildcard := ""
	if useWildcard {
		wildcard = ":*"
	}

	replacer := strings.NewReplacer("&", "", "'", "")
	sb := &strings.Builder{}
	for _, word := range q.words {
		// remove & from search input since it's an AND operator (in some datastores, like Postgres FTS);
		// remove ' from search input since it can break Postgres to_tsquery;
		// keep the original word for synonym lookup.
		renderedWord := replacer.Replace(word)
		if renderedWord == "" {
			continue
		}
		// add AND operator only between terms that were actually rendered.
		if sb.Len() > 0 {
			sb.WriteString(" & ")
		}
		if _, ok := q.withoutSynonyms[word]; ok {
			sb.WriteString(renderedWord)
			sb.WriteString(wildcard)
		} else if synonyms, ok := q.withSynonyms[word]; ok {
			slices.Sort(synonyms)
			sb.WriteByte('(')
			sb.WriteString(renderedWord)
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
