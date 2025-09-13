package main

import (
	"strings"
)

var (
	// cnf is the Config struct.
	cnf = struct {
		BufSize   uint16   `json:"buffer_size"` // max: 65535, 64KB
		PH        string   `json:"placeholder"`
		PHIndex   int      `json:"placeholder_index"`
		MaxWidth  int      `json:"max_width"`
		ModeStyle string   `json:"mode_style"`
		OldNew    []string `json:"old_new"`

		ModeStyleWidth int
		ModeStyleIndex int
		Replacer       *strings.Replacer
	}{}
)

// lastIndexNonSpace returns the index of last non-space character in s.
func lastIndexNonSpace(s []rune) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] != ' ' {
			return i
		}
	}
	return -1 // All were space.
}

// Read-only `[]byte(string)` convertions are optimized by compiler:
// https://github.com/golang/go/issues/2205 (commits=c8adb30,925d2fb,d63c88d).
// func string2Bytes(s string) []byte {
// 	if len(s) == 0 {
// 		return nil
// 	}
// 	return unsafe.Slice(unsafe.StringData(s), len(s))
// }

// func bytes2String(b []byte) string {
// 	if len(b) == 0 {
// 		return ""
// 	}
// 	return unsafe.String(unsafe.SliceData(b), len(b))
// }
