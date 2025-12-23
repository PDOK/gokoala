package search

import (
	"context"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	// change working dir to root
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../../../../")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func TestExpand(t *testing.T) {
	type args struct {
		searchQuery string
		useWildcard bool
		useSynonyms bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "rewrite",
			args: args{
				searchQuery: `markt den bosch`,
				useSynonyms: true,
			},
			want: `markt & hertogenbosch`,
		},
		{
			name: "rewrite followed by synonym",
			args: args{
				searchQuery: `Spui 1 den Haag`,
				useSynonyms: true,
			},
			want: `spui & 1 & (gravenhage | den <-> haag | s-gravenhage)`,
		},
		{
			name: "no synonym",
			args: args{
				searchQuery: `just some text`,
				useSynonyms: true,
			},
			want: `just & some & text`,
		},
		{
			name: "wildcard",
			args: args{
				searchQuery: `just some text`,
				useWildcard: true,
			},
			want: `just:* & some:* & text:*`,
		},
		{
			name: "one synonym",
			args: args{
				searchQuery: `Foo`,
				useSynonyms: true,
			},
			want: `(foo | foobar | foos)`,
		},
		{
			name: "two the same synonyms",
			args: args{
				searchQuery: `Foo FooBar`,
				useSynonyms: true,
			},
			want: `(foo | foobar | foos) & (foobar | foo | foos)`,
		},
		{
			name: "two-way synonym",
			args: args{
				searchQuery: `eerste 2de`,
				useSynonyms: true,
			},
			want: `(eerste | 1ste) & (2de | tweede)`,
		},
		{
			name: "nesting",
			args: args{
				searchQuery: `oudwesterlijke-goeverneur`,
				useSynonyms: true,
			},
			want: `
(oudwesterlijke-goeverneur | oudewestelijkelijke-goev | oudewestelijkelijke-goeverneur | oudewestelijkelijke-gouv | 
oudewestelijkelijke-gouverneur | oudewesterlijke-goev | oudewesterlijke-goeverneur | oudewesterlijke-gouv | 
oudewesterlijke-gouverneur | oudewestlijke-goev | oudewestlijke-goeverneur | oudewestlijke-gouv | oudewestlijke-gouverneur | 
oudwestelijkelijke-goev | oudwestelijkelijke-goeverneur | oudwestelijkelijke-gouv | oudwestelijkelijke-gouverneur | 
oudwesterlijke-goev | oudwesterlijke-gouv | oudwesterlijke-gouverneur | oudwestlijke-goev | 
oudwestlijke-goeverneur | oudwestlijke-gouv | oudwestlijke-gouverneur)
`,
		},
		{
			name: "overlapping synonyms",
			args: args{
				searchQuery: `foosball`,
				useSynonyms: true,
			},
			want: `(foosball | fooball | foobarball)`,
		},
		{
			name: "synonym with diacritics",
			args: args{
				searchQuery: `oude fryslân`,
				useSynonyms: true,
			},
			want: `(oude | oud) & (fryslân | friesland)`,
		},
		{
			name: "no synonyms for exact matches",
			args: args{
				searchQuery: `oude fryslân abc`,
				useSynonyms: false,
			},
			want: `(oude) & (fryslân) & abc`,
		},
		{
			name: "case insensitive",
			args: args{
				searchQuery: `OudE DeN HaAg`,
				useSynonyms: true,
			},
			want: `(oude | oud) & (gravenhage | den <-> haag | s-gravenhage)`,
		},
		{
			name: "word delimiters",
			args: args{
				searchQuery: `ok text with spaces ok`,
				useSynonyms: true,
			},
			want: `ok & text & with & spaces`,
		},
		{
			name: "long",
			args: args{
				searchQuery: `prof dr ir van der 1e noordsteeg`,
				useSynonyms: true,
			},
			want: `prof & dr & ir & van & der & 1e & noordsteeg`,
		},
		{
			name: "one substring",
			args: args{
				searchQuery: `Piet Gouverneurstraat 1800`,
				useSynonyms: true,
			},
			want: `
piet & (gouverneurstraat | goeverneurstraat | goevstraat | gouvstraat) & 1800
`,
		},
		{
			name: "two substrings",
			args: args{
				searchQuery: `Oude Piet Gouverneurstraat 1800`,
				useSynonyms: true,
			},
			want: `
(oude | oud) & piet & (gouverneurstraat | goeverneurstraat | goevstraat | gouvstraat) & 1800
`,
		},
		{
			name: "three substrings",
			args: args{
				searchQuery: `Oude Piet Westgouverneurstraat 1800`,
				useSynonyms: true,
			},
			want: `
(oude | oud) & piet & 
(westgouverneurstraat | westelijkegoeverneurstraat | westelijkegoevstraat | westelijkegouverneurstraat | 
westelijkegouvstraat | westergoeverneurstraat | westergoevstraat | westergouverneurstraat | westergouvstraat | 
westgoeverneurstraat | westgoevstraat | westgouvstraat) & 1800
`,
		},
		{
			name: "one rewrite and multiple synonyms",
			args: args{
				searchQuery: `goev straat 1 in Den Haag niet in Friesland`,
				useSynonyms: true,
			},
			want: `
(goev | goeverneur | gouv | gouverneur) & straat & 1 & in & (gravenhage | den <-> haag | s-gravenhage) & niet & (friesland | fryslân)
`,
		},
		{
			name: "five synonyms",
			args: args{
				searchQuery: `Oud Gouv 2DE 's-Gravenhage Fryslân Nederland`,
				useSynonyms: true,
			},
			want: `
(oud | oude) & (gouv | goev | goeverneur | gouverneur) & (2de | tweede) & (gravenhage | den <-> haag | s-gravenhage) & (fryslân | friesland) & nederland
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryExpansion, err := NewQueryExpansion("internal/ogc/features/testdata/search/rewrites.csv", "internal/ogc/features/testdata/search/synonyms.csv")
			require.NoError(t, err)
			actual, err := queryExpansion.Expand(context.Background(), tt.args.searchQuery)
			require.NoError(t, err)
			var query string
			if tt.args.useWildcard {
				query = actual.ToWildcardQuery()
			} else {
				query = actual.ToExactMatchQuery(tt.args.useSynonyms)
			}
			assert.Equal(t, strings.ReplaceAll(tt.want, "\n", ""), query, tt.args.searchQuery)
		})
	}
}
