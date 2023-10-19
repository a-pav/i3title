package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	// Globally shared variables.
	// For the sake of this program, these should work fine with no mutex mechanism in place.
	TITLE string
	LINE  []byte

	// cnf is the Config struct.
	cnf = struct {
		PH         string   `json:"placeholder"`
		Format     string   `json:"format"`
		OldNew     []string `json:"old_new"`
		MaxLen     int      `json:"max_length"`
		BufSize    uint16   `json:"buffer_size"` // max: 65535, 64KB
		StartDelay uint8    `json:"start_delay"` // max: 255
		Debug      bool     `json:"debug"`

		Replacer  *strings.Replacer
		MaxEscLen int
	}{}
)

// getwd returns the application's working directory.
// Using os.Args[0] is the surest way to get the actual CWD. i3 seems to run
// everything from /home/$USER.
func getwd() string { return filepath.Dir(os.Args[0]) }

// titlef formats title as defined in config.
func titlef(t string) string { return fmt.Sprintf(cnf.Format, t) }

// lastNonEscapeIndex checks if rs ends within a escape sequence. If so, it
// returns the index of escape sequece's beginning. Otherwise len(rs) is returned.
// width determines how far from rs's end is examined.
func lastNonEscapeIndex(rs []rune, width int) int {
	var (
		start      = len(rs) - width
		end        = len(rs)
		iSemicolon = -1
		iAmpersand = -1
	)

	for i := end - 1; i >= start; i-- {
		switch {
		case iAmpersand < 0 && rs[i] == '&':
			iAmpersand = i
		case iSemicolon < 0 && rs[i] == ';':
			iSemicolon = i
		}
	}

	switch {
	case iAmpersand < iSemicolon:
		// There's a fully formed escape sequence inside width with both `&` and
		// `;` (like: `&escape;`) Or, there's a `;` but no `&`. (like: `cape;`).
		return end
	case iAmpersand > 0:
		// There's a half-formed escape sequence inside width with only `&` and
		// no `;` (like: `&esca`)
		return iAmpersand
	}

	// No escape sequence was detected inside width.
	return end
}

// lastNonSpaceIndex returns the index of last non-space character in rs.
func lastNonSpaceIndex(rs []rune) int {
	for i := len(rs) - 1; i > 0; i-- {
		if rs[i] != ' ' {
			return i
		}
	}

	// All space!
	return 0
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
