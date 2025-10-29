package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// cnf is the Config struct.
var cnf = struct {
	BufSize   uint16   `json:"buffer_size"`       // Buffer size of both stdin scanner and stdout printer.
	PH        string   `json:"placeholder"`       // Placerhoder that is defined in i3status config file.
	PHIndex   int      `json:"placeholder_index"` // Index of placeholder in i3status output. Leaving it out causes re-calculation on each print.
	MaxWidth  int      `json:"max_width"`         // Maximum width of printed report in characters.
	ModeStyle string   `json:"mode_style"`        // Pango styling to be used for i3 modes.
	Pipe      string   `json:"pipe"`              // FIFO named pipe for sending messages to overwrite the report.
	OldNew    []string `json:"old_new"`           // List of old-new string pairs that will be used for the replacer.

	ModeStyleWidth int // Width of characters that will be added to report as the result of wrapping i3 mode with ModeStyle.
	ModeStyleIndex int // Index of string `%s` inside ModeStyle.
	Replacer       *strings.Replacer
}{}

func loadConfig() error {
	// Read config file from current working directory.
	cwd, err := getwd()
	if err != nil {
		return err
	}
	bs, err := os.ReadFile(cwd + "/config.json")
	if err != nil {
		return fmt.Errorf("opening config: %v", err)
	}
	// Strip comments before decoding.
	bs = regexp.MustCompile(`(?m)^\s*//.*$`).ReplaceAll(bs, nil)

	if err := json.Unmarshal(bs, &cnf); err != nil {
		return fmt.Errorf("reading config: %v", err)
	}
	defer discardConfig()

	// Strip styling tags, attrs and and any char that doesn't add to the width.
	raw := regexp.MustCompile("</?[^>]+>").ReplaceAllString(cnf.ModeStyle, "")
	cnf.ModeStyleWidth = len([]rune(raw)) - len("%s")

	cnf.ModeStyleIndex = strings.Index(cnf.ModeStyle, "%s")

	if len(cnf.OldNew)%2 == 1 {
		log.Println(`loadConfig: "old_new" list is ignored. odd number of arguments.`)
	} else {
		cnf.Replacer = strings.NewReplacer(cnf.OldNew...)
	}

	return nil
}

// discardConfig discards parts of the config that are no longer needed.
func discardConfig() {
	cnf.OldNew = nil // release reference
}

// getwd returns the executable working directory.
func getwd() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("getwd: executable path: %v", err)
	}
	// Clean up any symlinks in the executable path.
	exeRealpath, err := filepath.EvalSymlinks(exePath)
	if err != nil {
		return "", fmt.Errorf("getwd: executable real path: %v", err)
	}

	return filepath.Dir(exeRealpath), nil
}
