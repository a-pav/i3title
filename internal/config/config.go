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
	PH        string   `json:"placeholder"`       // Placerhoder that is defined in i3status config file. (default: I3TITLE)
	PHIndex   int      `json:"placeholder_index"` // Index of placeholder in i3status output. To force recalculation on each print, explicitly set to -1.
	MaxWidth  int      `json:"max_width"`         // Maximum width of printed report in characters.
	ModeStyle string   `json:"mode_style"`        // Pango styling to be used for i3 modes.
	Pipe      string   `json:"pipe"`              // FIFO named pipe for sending messages to overwrite the report.
	OldNew    []string `json:"old_new"`           // List of old-new string pairs that will be used for the replacer.

	ModeStyleWidth int               `json:"-"` // Width of characters that will be added to report as the result of wrapping raw i3 mode in ModeStyle.
	ModeStyleIndex int               `json:"-"` // Index of string `%s` inside ModeStyle.
	Replacer       *strings.Replacer `json:"-"`
}

func Load() (*Config, error) {
	path, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf("config: %s", err)
	}
	bs, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("opening config: %v", err)
	}
	// Strip comments before decoding.
	bs = regexp.MustCompile(`(?m)^\s*//.*$`).ReplaceAll(bs, nil)

	config := Config{}
	if err := json.Unmarshal(bs, &config); err != nil {
		return nil, fmt.Errorf("reading config: %v", err)
	}
	defer discard(&config)

	if config.PH == "" {
		config.PH = "I3TITLE" // default placeholder.
	}

	// Strip styling tags, attrs and and any char that doesn't add to the width.
	raw := regexp.MustCompile("</?[^>]+>").ReplaceAllString(config.ModeStyle, "")
	config.ModeStyleWidth = len([]rune(raw)) - len("%s")

	config.ModeStyleIndex = strings.Index(config.ModeStyle, "%s")

	if len(config.OldNew)%2 == 1 {
		log.Println(`loadConfig: "old_new" list is ignored. odd number of arguments.`)
	} else {
		config.Replacer = strings.NewReplacer(config.OldNew...)
	}

	log.Printf("config: loaded from: %s", path)
	dumpConfig(&config)

	return &config, nil
}

// dumpConfig dumps the config into the user config directory if the file doesn't
// exist already.
func dumpConfig(cfg *Config) {
	ucd, err := os.UserConfigDir()
	if err != nil {
		log.Printf("config: dump: os.UserConfigDir(): %s", err)
		return
	}
	dir := filepath.Join(ucd, "i3title")
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("config: dump: os.MkdirAll(): %s", err)
		return
	}
	path := filepath.Join(dir, "config.json")
	if err := fileExists(path); err == nil {
		return
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("config: dump: %s", err)
		return
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "\t")
	enc.SetEscapeHTML(false)

	if err := enc.Encode(cfg); err != nil {
		log.Printf("config: dump: failed to write %q: %v", path, err)
	}
}

// discard discards parts of the config that are no longer needed.
func discard(c *Config) {
	c.OldNew = nil // release reference
}

func getConfigPath() (string, error) {
	path, provided, err := configPathFromArgs()
	if provided {
		if err != nil {
			return "", err
		}
		return path, fileExists(path)
	}

	if path := os.Getenv("I3TITLE_CONFIG"); path != "" {
		return path, fileExists(path)
	}

	if ucd, err := os.UserConfigDir(); err == nil {
		path := filepath.Join(ucd, "i3title/config.json")
		if err := fileExists(path); err == nil {
			return path, nil
		}
	}

	// Read config file from current working directory.
	if cwd, err := getwd(); err == nil {
		path := filepath.Join(cwd, "config.json")
		if err := fileExists(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("failed to get config path.")
}

func configPathFromArgs() (path string, provided bool, err error) {
	args := os.Args
	for i, arg := range args {
		switch arg {
		case "-c", "-config", "--config":
			if i+1 < len(args) {
				return args[i+1], true, nil
			}
			return "", true, fmt.Errorf("%q flag is present but no path is provided", arg)
		}
		if strings.HasPrefix(arg, "-config=") {
			return strings.TrimPrefix(arg, "-config="), true, nil
		}
		if strings.HasPrefix(arg, "--config=") {
			return strings.TrimPrefix(arg, "--config="), true, nil
		}
	}
	return "", false, nil
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

// fileExists returns nil if the file path exists and is a regular file. Otherwise
// it returns a non-nil error explaining why.
func fileExists(path string) error {
	if path == "" {
		return fmt.Errorf("file path is empty")
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if !info.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file (mode: %v)", path, info.Mode())
	}

	return nil // File exists and is regular
}
