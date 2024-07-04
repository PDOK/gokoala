package util

import (
	"testing"
)

// Based upon os/env_test.go but modified to only allow ${VAR} and not $VAR expansion
func TestExpand(t *testing.T) {
	var expandTests = []struct {
		in, out string
		wantErr bool
	}{
		{"", "", false},
		{"$*", "all the args", true},
		{"$$", "PID", true},
		{"${*}", "all the args", false},
		{"$1", "ARGUMENT1", true},
		{"${1}", "ARGUMENT1", false},
		{"now is the time", "now is the time", false},
		{"$HOME", "/usr/gopher", true},
		{"$home_1", "/usr/foo", true},
		{"${HOME}", "/usr/gopher", false},
		{"${H}OME", "(Value of H)OME", false},
		{"A$$$#$1$H$home_1*B", "APIDNARGSARGUMENT1(Value of H)/usr/foo*B", true},
		{"start$+middle$^end$", "start$+middle$^end$", false},
		{"mixed$|bag$$$", "mixed$|bagPID$", true},
		{"$", "$", false},
		{"$}", "$}", false},
		{"${", "", false},  // invalid syntax; eat up the characters
		{"${}", "", false}, // invalid syntax; eat up the characters
	}
	for _, test := range expandTests {
		result := customExpand(test.in, getenvTestdata)
		if result != test.out {
			if !test.wantErr {
				t.Errorf("%q: got %q, want %q", test.in, result, test.out)
			}
		}
	}
}

func getenvTestdata(s string) string {
	switch s {
	case "*":
		return "all the args"
	case "#":
		return "NARGS"
	case "$":
		return "PID"
	case "1":
		return "ARGUMENT1"
	case "HOME":
		return "/usr/gopher"
	case "H":
		return "(Value of H)"
	case "home_1":
		return "/usr/foo"
	case "_":
		return "underscore"
	}
	return ""
}
