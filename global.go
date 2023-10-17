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

	Config = struct {
		Debug   bool   `json:"debug"`
		BufSize uint16 `json:"buffer_size"` // max: 65535, 64KB
		TiMod   struct {
			PH     string `json:"placeholder"`
			Format string `json:"format"`
			MaxLen int    `json:"max_length"`
			Delay  uint8  `json:"delay"`

			MaxEscLen int
		} `json:"title_module"`
		OldNew []string `json:"old_new"`

		Replacer *strings.Replacer
	}{}
)

// getwd returns the application's working directory.
// Using os.Args[0] is the surest way to get the actual CWD. i3 seems to run
// everything from /home/$USER.
func getwd() string { return filepath.Dir(os.Args[0]) }

// titlef formats title as defined in config.
func titlef(t string) string { return fmt.Sprintf(Config.TiMod.Format, t) }

// lastSafeIndex checks rs[len(rs)-width : len(rs)] for a half-formed escape
// sequence towards the end and returns an index at its beginning, making it easy
// to drop such escape sequence in the caller. If no half-formed escape sequence
// is detected, it returns len(rs).
func lastSafeIndex(rs []rune, width int) int {
	var (
		len0  = len(rs)
		start = len0 - width
		end   = len0 - 1

		iSemicolon = -1
		iAmpersand = -1
	)

	for i := end; i >= start; i-- {
		switch {
		case iAmpersand < 0 && rs[i] == '&':
			iAmpersand = i
		case iSemicolon < 0 && rs[i] == ';':
			iSemicolon = i
		}
	}

	switch {
	case iAmpersand < iSemicolon:
		// There's a fully formed escape sequence inside width before end. e.g. `&escape;`
		return len0
	case iAmpersand > 0:
		// There's a half-formed escape sequence inside width before end. e.g. `&esca`
		return iAmpersand
	}

	// There was no escape sequence inside the width.
	return len0
}

// lastNonspaceIndex returns the index of last non-space character in rs.
func lastNonspaceIndex(rs []rune) int {
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
