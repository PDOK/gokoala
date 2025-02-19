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
			want: `(spui & 1 & gravenhage)`,
		},
		{
			name: "remove user provided search operators",
			args: args{
				searchQuery: `A & B !C D <-> E`,
			},
			want: `(a & b & c & d & e)`,
		},
		{
			name: "no synonym",
			args: args{
				searchQuery: `just some text`,
			},
			want: `(just & some & text)`,
		},
		{
			name: "one synonym",
			args: args{
				searchQuery: `Foo`,
			},
			want: `(foo) | (foobar) | (foos)`,
		},
		{
			name: "two the same synonyms",
			args: args{
				searchQuery: `Foo FooBar`,
			},
			want: `(foo & foo) | (foo & foobar) | (foo & foos) | (foobar & foo) | (foobar & foobar) | (foobar & foos) | (foos & foo) | (foos & foobar) | (foos & foos)`,
		},
		{
			name: "two-way synonym",
			args: args{
				searchQuery: `eerste 2de`,
			},
			want: `(1ste & 2de) | (1ste & tweede) | (eerste & 2de) | (eerste & tweede)`,
		},
		{
			name: "overlapping synonyms",
			args: args{
				searchQuery: `foosball`,
			},
			want: `(fooball) | (foobarball) | (foosball)`,
		},
		{
			name: "synonym with diacritics",
			args: args{
				searchQuery: `oude fryslân`,
			},
			want: `(oud & friesland) | (oud & fryslân) | (oude & friesland) | (oude & fryslân)`,
		},
		{
			name: "case insensitive",
			args: args{
				searchQuery: `OudE DeN HaAg`,
			},
			want: `(oud & gravenhage) | (oude & gravenhage)`,
		},
		{
			name: "one substring",
			args: args{
				searchQuery: `1e Gouverneurstraat 1800`,
			},
			want: `
(1e & goeverneurstraat & 1800) | 
(1e & goevstraat & 1800) | 
(1e & gouverneurstraat & 1800) | 
(1e & gouvstraat & 1800)
`,
		},
		{
			name: "three substrings",
			args: args{
				searchQuery: `Oude Westelijker-Gouverneurstraat`,
			},
			want: `
(oud & westelijker-goeverneurstraat) | 
(oud & westelijker-goevstraat) | 
(oud & westelijker-gouverneurstraat) | 
(oud & westelijker-gouvstraat) | 
(oud & westerr-goeverneurstraat) | 
(oud & westerr-goevstraat) | 
(oud & westerr-gouverneurstraat) | 
(oud & westerr-gouvstraat) | 
(oud & westr-goeverneurstraat) | 
(oud & westr-goevstraat) | 
(oud & westr-gouverneurstraat) | 
(oud & westr-gouvstraat) | 
(oud & wr-goeverneurstraat) | 
(oud & wr-goevstraat) | 
(oud & wr-gouverneurstraat) | 
(oud & wr-gouvstraat) | 
(oude & westelijker-goeverneurstraat) | 
(oude & westelijker-goevstraat) | 
(oude & westelijker-gouverneurstraat) | 
(oude & westelijker-gouvstraat) | 
(oude & westerr-goeverneurstraat) | 
(oude & westerr-goevstraat) | 
(oude & westerr-gouverneurstraat) | 
(oude & westerr-gouvstraat) | 
(oude & westr-goeverneurstraat) | 
(oude & westr-goevstraat) | 
(oude & westr-gouverneurstraat) | 
(oude & westr-gouvstraat) | 
(oude & wr-goeverneurstraat) | 
(oude & wr-goevstraat) | 
(oude & wr-gouverneurstraat) | 
(oude & wr-gouvstraat)
`,
		},
		{
			name: "one rewrite and multiple synonyms",
			args: args{
				searchQuery: `goev straat 1 in Den Haag niet in Friesland`,
			},
			want: `
(goev & straat & 1 & in & gravenhage & niet & in & friesland) | 
(goev & straat & 1 & in & gravenhage & niet & in & fryslân) | 
(goeverneur & straat & 1 & in & gravenhage & niet & in & friesland) | 
(goeverneur & straat & 1 & in & gravenhage & niet & in & fryslân) | 
(gouv & straat & 1 & in & gravenhage & niet & in & friesland) | 
(gouv & straat & 1 & in & gravenhage & niet & in & fryslân) | 
(gouverneur & straat & 1 & in & gravenhage & niet & in & friesland) | 
(gouverneur & straat & 1 & in & gravenhage & niet & in & fryslân)
`,
		},
		{
			name: "lots of synonyms",
			args: args{
				searchQuery: `Oud Gouv 2DE 's-Gravenhage Fryslân Nederland`,
			},
			want: `
(oud & goev & 2de & gravenhage & friesland & nederland) | 
(oud & goev & 2de & gravenhage & fryslân & nederland) | 
(oud & goev & tweede & gravenhage & friesland & nederland) | 
(oud & goev & tweede & gravenhage & fryslân & nederland) | 
(oud & goeverneur & 2de & gravenhage & friesland & nederland) | 
(oud & goeverneur & 2de & gravenhage & fryslân & nederland) | 
(oud & goeverneur & tweede & gravenhage & friesland & nederland) | 
(oud & goeverneur & tweede & gravenhage & fryslân & nederland) | 
(oud & gouv & 2de & gravenhage & friesland & nederland) | 
(oud & gouv & 2de & gravenhage & fryslân & nederland) | 
(oud & gouv & tweede & gravenhage & friesland & nederland) | 
(oud & gouv & tweede & gravenhage & fryslân & nederland) | 
(oud & gouverneur & 2de & gravenhage & friesland & nederland) | 
(oud & gouverneur & 2de & gravenhage & fryslân & nederland) | 
(oud & gouverneur & tweede & gravenhage & friesland & nederland) | 
(oud & gouverneur & tweede & gravenhage & fryslân & nederland) | 
(oude & goev & 2de & gravenhage & friesland & nederland) | 
(oude & goev & 2de & gravenhage & fryslân & nederland) | 
(oude & goev & tweede & gravenhage & friesland & nederland) | 
(oude & goev & tweede & gravenhage & fryslân & nederland) | 
(oude & goeverneur & 2de & gravenhage & friesland & nederland) | 
(oude & goeverneur & 2de & gravenhage & fryslân & nederland) | 
(oude & goeverneur & tweede & gravenhage & friesland & nederland) | 
(oude & goeverneur & tweede & gravenhage & fryslân & nederland) | 
(oude & gouv & 2de & gravenhage & friesland & nederland) | 
(oude & gouv & 2de & gravenhage & fryslân & nederland) | 
(oude & gouv & tweede & gravenhage & friesland & nederland) | 
(oude & gouv & tweede & gravenhage & fryslân & nederland) | 
(oude & gouverneur & 2de & gravenhage & friesland & nederland) | 
(oude & gouverneur & 2de & gravenhage & fryslân & nederland) | 
(oude & gouverneur & tweede & gravenhage & friesland & nederland) | 
(oude & gouverneur & tweede & gravenhage & fryslân & nederland)
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
