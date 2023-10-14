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
	log.SetPrefix("i3title: ")
	log.SetFlags(log.Lmsgprefix)

	initConfig(getwd())

	TITLE = fmt.Sprintf(Config.TitleMod.Format, Config.TitleMod.WelcomeMsg)
}

// getwd returns the application's working directory.
// os.Args[0] is the surest way to get the actual CWD. i3 seems to run everything in /home/$USER.
func getwd() string { return filepath.Dir(os.Args[0]) }

func initConfig(cwd string) {
	readConfigFile(cwd)

	if len(Config.Filters)%2 == 1 {
		log.Println("Config.Filters: odd argument count. filter list ignored.")
	} else {
		filtersCompiled := []MatchReplace{}
		for i := 0; i < len(Config.Filters); i += 2 {
			re, err := regexp.Compile(Config.Filters[i])
			if err != nil {
				log.Println("error compiling regex:", err)
				continue
			}

			filtersCompiled = append(filtersCompiled, MatchReplace{
				Match: re,
				Repl:  Config.Filters[i+1],
			})
		}

		Config.FiltersCompiled = filtersCompiled
	}
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
	Config.TitleMod.WelcomeMsg = ""
}
