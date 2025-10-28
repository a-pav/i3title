package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"regexp"
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
		Pipe      string   `json:"pipe"` // FIFO pipe for sending messages to overwrite report.
		OldNew    []string `json:"old_new"`

		ModeStyleWidth int
		ModeStyleIndex int
		Replacer       *strings.Replacer
	}{}
)

func loadConfig() {
	// Read config file from current working directory.
	cwd := getwd()
	bs, err := os.ReadFile(cwd + "/config.json")
	if err != nil {
		log.Fatal("opening config file: ", err)
	}
	// Strip comments before decoding.
	bs = regexp.MustCompile(`(?m)^\s*//.*$`).ReplaceAll(bs, nil)

	if err := json.Unmarshal(bs, &cnf); err != nil {
		log.Fatal("reading config file: ", err)
	}
	defer discardConfig()

	// Strip styling tags, attrs and and any char that doesn't add to the width.
	raw := regexp.MustCompile("</?[^>]+>").ReplaceAllString(cnf.ModeStyle, "")
	cnf.ModeStyleWidth = len([]rune(raw)) - len("%s")

	cnf.ModeStyleIndex = strings.Index(cnf.ModeStyle, "%s")

	if len(cnf.OldNew)%2 == 1 {
		log.Println(`cnf.OldNew: odd number of arguments, "old_new" list is ignored.`)
	} else {
		// Initialize strings.Replacer.
		cnf.Replacer = strings.NewReplacer(cnf.OldNew...)
	}
}

// discardConfig discards parts of the config that are no longer needed.
func discardConfig() {
	cnf.OldNew = nil // release reference
}

// getwd returns the application's working directory.
func getwd() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalln("finding executable path:", err)
	}
	// Clean up any symlinks in the executable path.
	exeRealpath, err := filepath.EvalSymlinks(exePath)
	if err != nil {
		log.Fatalln("finding executable real path: ", err)
	}

	return filepath.Dir(exeRealpath)
}

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
