package search

import (
	"os"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	// change working dir to root
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../../")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func TestExpand(t *testing.T) {
	type args struct {
		searchQuery string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "rewrite",
			args: args{
				searchQuery: `Spui 1 den Haag`,
			},
			want: `spui & 1 & gravenhage`,
		},
		{
			name: "no synonym",
			args: args{
				searchQuery: `just some text`,
			},
			want: `just & some & text`,
		},
		{
			name: "one synonym",
			args: args{
				searchQuery: `Foo`,
			},
			want: `(foo | foobar | foos)`,
		},
		{
			name: "two the same synonyms",
			args: args{
				searchQuery: `Foo FooBar`,
			},
			want: `(foo | foobar | foos) & (foobar | foo | foos)`,
		},
		{
			name: "two-way synonym",
			args: args{
				searchQuery: `eerste 2de`,
			},
			want: `(eerste | 1ste) & (2de | tweede)`,
		},
		{
			name: "nesting",
			args: args{
				searchQuery: `oudwesterlijke-goeverneur`,
			},
			want: `
(oudwesterlijke-goeverneur | oudewestelijkelijke-goev | oudewestelijkelijke-goeverneur | oudewestelijkelijke-gouv | 
oudewestelijkelijke-gouverneur | oudewesterlijke-goev | oudewesterlijke-goeverneur | oudewesterlijke-gouv | 
oudewesterlijke-gouverneur | oudewestlijke-goev | oudewestlijke-goeverneur | oudewestlijke-gouv | oudewestlijke-gouverneur | 
oudwestelijkelijke-goev | oudwestelijkelijke-goeverneur | oudwestelijkelijke-gouv | oudwestelijkelijke-gouverneur | 
oudwesterlijke-goev | oudwesterlijke-gouv | oudwesterlijke-gouverneur | oudwestlijke-goev | 
oudwestlijke-goeverneur | oudwestlijke-gouvoudwestlijke-gouverneur)
`,
		},
		{
			name: "overlapping synonyms",
			args: args{
				searchQuery: `foosball`,
			},
			want: `(foosball | fooball | foobarball)`,
		},
		{
			name: "synonym with diacritics",
			args: args{
				searchQuery: `oude fryslân`,
			},
			want: `(oude | oud) & (fryslân | friesland)`,
		},
		{
			name: "case insensitive",
			args: args{
				searchQuery: `OudE DeN HaAg`,
			},
			want: `(oude | oud) & gravenhage`,
		},
		{
			name: "word delimiters",
			args: args{
				searchQuery: `ok text with spaces ok`,
			},
			want: `ok & text & with & spaces`,
		},
		{
			name: "long",
			args: args{
				searchQuery: `prof dr ir van der 1e noordsteeg`,
			},
			want: `prof & dr & ir & van & der & 1e & noordsteeg`,
		},
		{
			name: "one substring",
			args: args{
				searchQuery: `Piet Gouverneurstraat 1800`,
			},
			want: `
piet & (gouverneurstraat | goeverneurstraat | goevstraat | gouvstraat) & 1800
`,
		},
		{
			name: "two substrings",
			args: args{
				searchQuery: `Oude Piet Gouverneurstraat 1800`,
			},
			want: `
(oude | oud) & piet & (gouverneurstraat | goeverneurstraat | goevstraat | gouvstraat) & 1800
`,
		},
		{
			name: "three substrings",
			args: args{
				searchQuery: `Oude Piet Westgouverneurstraat 1800`,
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
			},
			want: `
(goev | goeverneur | gouv | gouverneur) & straat & 1 & in & gravenhage & niet & (friesland | fryslân)
`,
		},
		{
			name: "four synonyms",
			args: args{
				searchQuery: `Oud Gouv 2DE 's-Gravenhage Fryslân Nederland`,
			},
			want: `
(oud | oude) & (gouv | goev | goeverneur | gouverneur) & (2de | tweede) & gravenhage & (fryslân | friesland) & nederland
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryExpansion, err := NewQueryExpansion("internal/search/testdata/rewrites.csv", "internal/search/testdata/synonyms.csv")
			assert.NoError(t, err)
			actual := queryExpansion.Expand(tt.args.searchQuery)
			assert.Equal(t, strings.ReplaceAll(tt.want, "\n", ""), actual.ToExactMatchQuery(), tt.args.searchQuery)
		})
	}
}
