package main

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	MODE_SEP     = " <span color='#666666'>|</span> "
	MODE_SEP_LEN = 3
)

var (
	// cnf is the Config struct.
	cnf = struct {
		PH         string   `json:"placeholder"`
		OldNew     []string `json:"old_new"`
		MaxLen     int      `json:"max_length"`
		BufSize    uint16   `json:"buffer_size"` // max: 65535, 64KB
		StartDelay uint8    `json:"start_delay"` // max: 255
		Debug      bool     `json:"debug"`

		Replacer *strings.Replacer
	}{}
)

// getwd returns the application's working directory.
// Using os.Args[0] is the surest way to get the actual CWD. i3 seems to run
// everything from /home/$USER.
func getwd() string { return filepath.Dir(os.Args[0]) }

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
