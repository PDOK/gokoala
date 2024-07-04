package util

import (
	"os"
	_ "unsafe" // required to access the private getShellName() func.
)

// Trick to access the private getShellName func
//
//go:linkname getShellName os.getShellName
func getShellName(s string) (string, int)

// ExpandEnv replaces ${var} in the string according to the values
// of the current environment variables. References to undefined
// variables are replaced by the empty string.
func ExpandEnv(s string) string {
	return customExpand(s, os.Getenv)
}

// CustomExpand is a slightly customized version of os.Expand that replaces ${var} BUT NOT $var
// in the string based on the mapping function.
//
// This code is copied from os.Expand and modified.
// Original copyright note:
//
// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
func customExpand(s string, mapping func(string) string) string {
	var buf []byte
	// ${} is all ASCII, so bytes are fine for this operation.
	i := 0
	for j := 0; j < len(s); j++ {
		if s[j] == '$' && j+1 < len(s) && s[j+1] == '{' {
			if buf == nil {
				buf = make([]byte, 0, 2*len(s))
			}
			buf = append(buf, s[i:j]...)
			name, w := getShellName(s[j+1:])
			switch {
			case name == "" && w > 0:
			case name == "":
				buf = append(buf, s[j])
			default:
				buf = append(buf, mapping(name)...)
			}
			j += w
			i = j + 1
		}
	}
	if buf == nil {
		return s
	}
	return string(buf) + s[i:]
}
