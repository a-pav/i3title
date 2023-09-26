package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"regexp"
)

func init() {
	defer discardConfig()
	// log to stderr since stdout is strictly for valid json/array lines.
	log.SetOutput(os.Stderr)

	initScanner()
	initConfig(os.Args[0])

	TITLE = Config.TitleModule.WelcomeMsg
}

func initScanner() {
	Scanner = bufio.NewScanner(os.Stdin)

	if err := Scanner.Err(); err != nil {
		log.Fatal("scanner failed to init: ", err)
	}
}

func initConfig(args0 string) {
	readConfigFile(args0)

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

func readConfigFile(args0 string) {
	// read config file from '<program-name>.config.json'
	bs, err := os.ReadFile(args0 + ".config.json")
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
