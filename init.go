package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

func init() {
	defer discardConfig()
	// log to stderr since stdout is strictly for valid json/array lines.
	log.SetOutput(os.Stderr)

	initConfig(getwd())

	TITLE = fmt.Sprintf(Config.TitleModule.Format, Config.TitleModule.WelcomeMsg)
}

// getwd returns the application's working directory.
// os.Args[0] is the surest way to get the actual CWD. i3 seems to run everything in /home/$USER.
func getwd() string { return filepath.Dir(os.Args[0]) }

func initConfig(cwd string) {
	readConfigFile(cwd)

	// Compile placeholder text.
	Config.TitlePHRE = regexp.MustCompile(Config.TitleModule.PH)

	filtersCompiled := []MatchReplace{}
	for _, mr := range Config.Filters {
		if l := len(mr); l != 2 {
			continue
		}

		re, err := regexp.Compile(mr[0])
		if err != nil {
			log.Println("error compiling regex:", err)
			continue
		}

		filtersCompiled = append(filtersCompiled, MatchReplace{
			Match: re,
			Repl:  mr[1],
		})
	}

	Config.FiltersCompiled = filtersCompiled
}

func readConfigFile(cwd string) {
	// read config file from '<program-name>.config.json'
	bs, err := os.ReadFile(cwd + "/config.json")
	if err != nil {
		log.Fatal("opening config file: ", err)
	}

	// strip comments
	bs = regexp.MustCompile(`//.*`).ReplaceAll(bs, nil)

	if err := json.Unmarshal(bs, &Config); err != nil {
		log.Fatal("reading config file: ", err)
	}
}

// discardConfig discards parts of the config that are no longer needed.
func discardConfig() {
	Config.Filters = nil // discard uncompiled filters.
	Config.TitleModule.WelcomeMsg = ""
}
