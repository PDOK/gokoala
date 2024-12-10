package transform

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_generateCombinations(t *testing.T) {
	type args struct {
		keys     []string
		values   [][]string
		keyDepth int
		current  map[string]any
	}
	tests := []struct {
		name string
		args args
		want []map[string]any
	}{
		{"Single key, single value", args{[]string{"key1"}, [][]string{{"value1"}}, 0, map[string]any{}}, []map[string]any{{"key1": "value1"}}},
		{"Single key, slice of values", args{[]string{"key1"}, [][]string{{"value1", "value2"}}, 0, map[string]any{}}, []map[string]any{{"key1": "value1"}, {"key1": "value2"}}},
		{"Two keys, two single values", args{[]string{"key1", "key2"}, [][]string{{"value1"}, {"value2"}}, 0, map[string]any{}}, []map[string]any{{"key1": "value1", "key2": "value2"}}},
		{"Two keys, slice + single value", args{[]string{"key1", "key2"}, [][]string{{"value1", "value2"}, {"value3"}}, 0, map[string]any{}}, []map[string]any{{"key1": "value1", "key2": "value3"}, {"key1": "value2", "key2": "value3"}}},
		{"Two keys, two slices values", args{[]string{"key1", "key2"}, [][]string{{"value1", "value2"}, {"value3", "value4"}}, 0, map[string]any{}}, []map[string]any{{"key1": "value1", "key2": "value3"}, {"key1": "value1", "key2": "value4"}, {"key1": "value2", "key2": "value3"}, {"key1": "value2", "key2": "value4"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, generateCombinations(tt.args.keys, tt.args.values, tt.args.keyDepth, tt.args.current), "generateCombinations(%v, %v, %v, %v)", tt.args.keys, tt.args.values, tt.args.keyDepth, tt.args.current)
		})
	}
}

func Test_applySubstitution(t *testing.T) {
	type args struct {
		input         string
		substitutions map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr assert.ErrorAssertionFunc
	}{
		{"No substitution", args{input: "foobar", substitutions: map[string]string{"ae": "a", "à": "a"}}, []string{"foobar"}, assert.NoError},
		{"Single substitution", args{input: "achtaertune", substitutions: map[string]string{"ae": "a", "à": "a"}}, []string{"achtaertune", "achtartune"}, assert.NoError},
		{"Multiple but different substitutions", args{input: "klàr achtaertune", substitutions: map[string]string{"ae": "a", "à": "a"}}, []string{"klàr achtaertune", "klàr achtartune", "klar achtartune", "klar achtaertune"}, assert.NoError},
		{"Two from same substitutions", args{input: "naer achtaertune", substitutions: map[string]string{"ae": "a", "à": "a"}}, []string{"naer achtaertune", "naer achtartune", "nar achtartune", "nar achtaertune"}, assert.NoError},
		{"Three from same substitutions", args{input: "naer achtaertune kaerel", substitutions: map[string]string{"ae": "a", "à": "a"}}, []string{"naer achtaertune kaerel", "naer achtartune kaerel", "nar achtartune kaerel", "nar achtaertune kaerel", "naer achtaertune karel", "naer achtartune karel", "nar achtartune karel", "nar achtaertune karel"}, assert.NoError},
		{"Single substitution with capital", args{input: "Aechtartune", substitutions: map[string]string{"ae": "a", "à": "a"}}, []string{"aechtartune", "achtartune"}, assert.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := applySubstitutions(tt.args.input, tt.args.substitutions)
			if !tt.wantErr(t, err, fmt.Sprintf("applySubstitutions(%v, %v)", tt.args.input, tt.args.substitutions)) {
				return
			}
			assert.ElementsMatch(t, tt.want, got, "applySubstitutions(%v, %v)", tt.args.input, tt.args.substitutions)
		})
	}
}

func Test_replaceNth(t *testing.T) {
	type args struct {
		input    string
		oldChar  string
		newChar  string
		nthIndex int
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{"Replace 1st only", args{"naer achtaertune kaerel", "ae", "a", 1}, "nar achtaertune kaerel", assert.NoError},
		{"Replace 2nd only", args{"naer achtaertune kaerel", "ae", "a", 2}, "naer achtartune kaerel", assert.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := replaceNth(tt.args.input, tt.args.oldChar, tt.args.newChar, tt.args.nthIndex)
			if !tt.wantErr(t, err, fmt.Sprintf("replaceNth(%v, %v, %v, %v)", tt.args.input, tt.args.oldChar, tt.args.newChar, tt.args.nthIndex)) {
				return
			}
			assert.Equalf(t, tt.want, got, "replaceNth(%v, %v, %v, %v)", tt.args.input, tt.args.oldChar, tt.args.newChar, tt.args.nthIndex)
		})
	}
}

func Test_uniqueSlice(t *testing.T) {
	type args struct {
		s []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"Already unique slice", args{[]string{"foo", "bar", "baz"}}, []string{"foo", "bar", "baz"}},
		{"Slice with duplicate", args{[]string{"foo", "bar", "foo", "baz"}}, []string{"foo", "bar", "baz"}},
		{"Slice with multiple duplicate", args{[]string{"foo", "bar", "foo", "baz", "baz", "foo"}}, []string{"foo", "bar", "baz"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := uniqueSlice(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("uniqueSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_readSubstitutions(t *testing.T) {
	type args struct {
		filepath string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr assert.ErrorAssertionFunc
	}{
		{"Read substitutions csv", args{"../testdata/substitution.csv"}, map[string]string{"ae": "a", "à": "a"}, assert.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readSubstitutionsFile(tt.args.filepath)
			if !tt.wantErr(t, err, fmt.Sprintf("readSubstitutionsFile(%v)", tt.args.filepath)) {
				return
			}
			assert.Equalf(t, tt.want, got, "readSubstitutionsFile(%v)", tt.args.filepath)
		})
	}
}
