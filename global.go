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
		Debug    bool   `json:"debug"`
		BufSize  uint16 `json:"buffer_size"` // max: 65535, 64KB
		TitleMod struct {
			PH         string `json:"placeholder"`
			Format     string `json:"format"`
			MaxLen     int    `json:"max_length"`
			Delay      uint8  `json:"delay"`
			WelcomeMsg string `json:"welcome_msg"` // It is shown until the first window/title event is triggered.
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
func titlef(t string) string { return fmt.Sprintf(Config.TitleMod.Format, t) }

// nonspaceIndexRight returns the index of first nonspace character in rs.
func nonspaceIndexRight(rs []rune) int {
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
