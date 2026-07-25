package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Config is the Config struct.
type Config struct {
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
}

func Load() (*Config, error) {
	config := Config{}
	// Read config file from current working directory.
	cwd, err := getwd()
	if err != nil {
		return nil, err
	}
	bs, err := os.ReadFile(cwd + "/config.json")
	if err != nil {
		return nil, fmt.Errorf("opening config: %v", err)
	}
	// Strip comments before decoding.
	bs = regexp.MustCompile(`(?m)^\s*//.*$`).ReplaceAll(bs, nil)

	if err := json.Unmarshal(bs, &config); err != nil {
		return nil, fmt.Errorf("reading config: %v", err)
	}
	defer discard(&config)

	// Strip styling tags, attrs and and any char that doesn't add to the width.
	raw := regexp.MustCompile("</?[^>]+>").ReplaceAllString(config.ModeStyle, "")
	config.ModeStyleWidth = len([]rune(raw)) - len("%s")

	config.ModeStyleIndex = strings.Index(config.ModeStyle, "%s")

	if len(config.OldNew)%2 == 1 {
		log.Println(`loadConfig: "old_new" list is ignored. odd number of arguments.`)
	} else {
		config.Replacer = strings.NewReplacer(config.OldNew...)
	}

	return &config, nil
}

// discard discards parts of the config that are no longer needed.
func discard(c *Config) {
	c.OldNew = nil // release reference
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
