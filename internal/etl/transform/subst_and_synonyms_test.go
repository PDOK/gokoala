package transform

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_generate(t *testing.T) {
	type args struct {
		fieldValuesByName map[string]string
		substitutionsFile string
		synonymsFile      string
	}
	tests := []struct {
		name    string
		args    args
		want    []map[string]string
		wantErr assert.ErrorAssertionFunc
	}{
		{"simple record", args{map[string]string{"component_thoroughfarename": "foo", "component_postaldescriptor": "1234AB", "component_addressareaname": "bar"}, "../testdata/substitutions.csv", "../testdata/synonyms.csv"}, []map[string]string{{"component_thoroughfarename": "foo", "component_postaldescriptor": "1234ab", "component_addressareaname": "bar"}}, assert.NoError},
		{"single synonym record", args{map[string]string{"component_thoroughfarename": "eerste", "component_postaldescriptor": "1234AB", "component_addressareaname": "bar"}, "../testdata/substitutions.csv", "../testdata/synonyms.csv"}, []map[string]string{{"component_thoroughfarename": "eerste", "component_postaldescriptor": "1234ab", "component_addressareaname": "bar"}, {"component_thoroughfarename": "1ste", "component_postaldescriptor": "1234ab", "component_addressareaname": "bar"}}, assert.NoError},
		{"single synonym with capital", args{map[string]string{"component_thoroughfarename": "Eerste", "component_postaldescriptor": "1234AB", "component_addressareaname": "bar"}, "../testdata/substitutions.csv", "../testdata/synonyms.csv"}, []map[string]string{{"component_thoroughfarename": "eerste", "component_postaldescriptor": "1234ab", "component_addressareaname": "bar"}, {"component_thoroughfarename": "1ste", "component_postaldescriptor": "1234ab", "component_addressareaname": "bar"}}, assert.NoError},
		{"two-way synonym record", args{map[string]string{"component_thoroughfarename": "eerste 2de", "component_postaldescriptor": "1234AB", "component_addressareaname": "bar"}, "../testdata/substitutions.csv", "../testdata/synonyms.csv"}, []map[string]string{{"component_thoroughfarename": "eerste 2de", "component_postaldescriptor": "1234ab", "component_addressareaname": "bar"}, {"component_thoroughfarename": "1ste 2de", "component_postaldescriptor": "1234ab", "component_addressareaname": "bar"}, {"component_thoroughfarename": "eerste tweede", "component_postaldescriptor": "1234ab", "component_addressareaname": "bar"}, {"component_thoroughfarename": "1ste tweede", "component_postaldescriptor": "1234ab", "component_addressareaname": "bar"}}, assert.NoError},
		{"avoid endless loop for synonyms that contain source value", args{map[string]string{"street": "oude kerkstraat"}, "../testdata/substitutions.csv", "../testdata/synonyms.csv"}, []map[string]string{{"street": "oude kerkstraat"}, {"street": "oud kerkstraat"}}, assert.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ss, err := NewSubstAndSynonyms(tt.args.substitutionsFile, tt.args.synonymsFile)
			assert.NoError(t, err)
			assert.Equalf(t, tt.want, ss.generate(tt.args.fieldValuesByName), "generate(%v, %v, %v)", tt.args.fieldValuesByName, tt.args.substitutionsFile, tt.args.synonymsFile)
		})
	}
}

func Test_generateCombinations(t *testing.T) {
	type args struct {
		keys   []string
		values [][]string
	}
	tests := []struct {
		name string
		args args
		want []map[string]string
	}{
		{"Single key, single value", args{[]string{"key1"}, [][]string{{"value1"}}}, []map[string]string{{"key1": "value1"}}},
		{"Single key, slice of values", args{[]string{"key1"}, [][]string{{"value1", "value2"}}}, []map[string]string{{"key1": "value1"}, {"key1": "value2"}}},
		{"Two keys, two single values", args{[]string{"key1", "key2"}, [][]string{{"value1"}, {"value2"}}}, []map[string]string{{"key1": "value1", "key2": "value2"}}},
		{"Two keys, slice + single value", args{[]string{"key1", "key2"}, [][]string{{"value1", "value2"}, {"value3"}}}, []map[string]string{{"key1": "value1", "key2": "value3"}, {"key1": "value2", "key2": "value3"}}},
		{"Two keys, two slices values", args{[]string{"key1", "key2"}, [][]string{{"value1", "value2"}, {"value3", "value4"}}}, []map[string]string{{"key1": "value1", "key2": "value3"}, {"key1": "value1", "key2": "value4"}, {"key1": "value2", "key2": "value3"}, {"key1": "value2", "key2": "value4"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateCombinations(tt.args.keys, tt.args.values)
			assert.Equalf(t, tt.want, got, "generateCombinations(%v, %v)", tt.args.keys, tt.args.values)
		})
	}
}

func Test_extendValues(t *testing.T) {
	type args struct {
		input   []string
		mapping map[string]string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"No mapping", args{input: []string{"foobar"}, mapping: map[string]string{"eerste": "1ste", "tweede": "2de", "fryslân": "friesland"}}, []string{"foobar"}},
		{"Single mapping", args{input: []string{"foobar eerste"}, mapping: map[string]string{"eerste": "1ste", "tweede": "2de", "fryslân": "friesland"}}, []string{"foobar eerste", "foobar 1ste"}},
		{"Two different mappings", args{input: []string{"foobar eerste tweede"}, mapping: map[string]string{"eerste": "1ste", "tweede": "2de", "fryslân": "friesland"}}, []string{"foobar eerste tweede", "foobar 1ste tweede", "foobar eerste 2de", "foobar 1ste 2de"}},
		{"Two similar mappings", args{input: []string{"foobar eerste eerste"}, mapping: map[string]string{"eerste": "1ste", "tweede": "2de", "fryslân": "friesland"}}, []string{"foobar eerste eerste", "foobar 1ste eerste", "foobar eerste 1ste", "foobar 1ste 1ste"}},
		{"Three from same mapping", args{input: []string{"naer achtaertune kaerel"}, mapping: map[string]string{"ae": "a", "à": "a"}}, []string{"naer achtaertune kaerel", "naer achtartune kaerel", "nar achtartune kaerel", "nar achtaertune kaerel", "naer achtaertune karel", "naer achtartune karel", "nar achtartune karel", "nar achtaertune karel"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extendValues(tt.args.input, tt.args.mapping)
			assert.ElementsMatch(t, tt.want, got, "extendValues(%v, %v)", tt.args.input, tt.args.mapping)
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
		name string
		args args
		want string
	}{
		{"Replace 1st only", args{"naer achtaertune kaerel", "ae", "a", 1}, "nar achtaertune kaerel"},
		{"Replace 2nd only", args{"naer achtaertune kaerel", "ae", "a", 2}, "naer achtartune kaerel"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replaceNth(tt.args.input, tt.args.oldChar, tt.args.newChar, tt.args.nthIndex)
			assert.Equalf(t, tt.want, got, "replaceNth(%v, %v, %v, %v)", tt.args.input, tt.args.oldChar, tt.args.newChar, tt.args.nthIndex)
		})
	}
}

func Test_uniqueSlice1(t *testing.T) {
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
			assert.Equalf(t, tt.want, uniqueSlice(tt.args.s), "uniqueSlice(%v)", tt.args.s)
		})
	}
}

func Test_readCsvFile(t *testing.T) {
	type args struct {
		filepath string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr assert.ErrorAssertionFunc
	}{
		{"Read substitutions csv", args{"../testdata/substitutions.csv"}, map[string]string{"ae": "a", "à": "a"}, assert.NoError},
		{"Read synonyms csv", args{"../testdata/synonyms.csv"}, map[string]string{"eerste": "1ste", "fryslân": "friesland", "oud": "oude", "tweede": "2de"}, assert.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readCsvFile(tt.args.filepath)
			if !tt.wantErr(t, err, fmt.Sprintf("readCsvFile(%v)", tt.args.filepath)) {
				return
			}
			assert.Equalf(t, tt.want, got, "readCsvFile(%v)", tt.args.filepath)
		})
	}
}
