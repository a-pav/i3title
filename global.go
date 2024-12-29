package main

import (
	"strings"
)

var (
	// cnf is the Config struct.
	cnf = struct {
		PH           string   `json:"placeholder"`
		ModeStyle    string   `json:"mode_style"`
		ModeStyleLen int      `json:"mode_style_length"`
		OldNew       []string `json:"old_new"`
		MaxLen       int      `json:"max_length"`
		BufSize      uint16   `json:"buffer_size"` // max: 65535, 64KB

		Replacer *strings.Replacer
	}{}
)

// lastNonspaceIndex returns the index of last non-space character in s.
func lastNonspaceIndex(s []rune) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] != ' ' {
			return i
		}
	}
	return 0 // All were space.
}

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
