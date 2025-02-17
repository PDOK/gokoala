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

func TestQueryExpansion_Expand(t *testing.T) {
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
			name: "one synonym",
			args: args{
				searchQuery: `Foo`,
			},
			want: `(foo) | (foobar) | (foos)`,
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
			want: `(foosball)`,
		},
		{
			name: "synonym with diacritics",
			args: args{
				searchQuery: `oude fryslân`,
			},
			want: `(oud & friesland) | (oud & fryslân) | (oude & friesland) | (oude & fryslân)`,
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
