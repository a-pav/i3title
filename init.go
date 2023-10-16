package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

func init() {
	defer discardConfig()
	log.SetPrefix("i3title: ")
	log.SetFlags(log.Lmsgprefix)

	initConfig(getwd())

	TITLE = titlef(fmt.Sprintf("%s | Delay: %d", Config.TitleMod.WelcomeMsg, Config.TitleMod.Delay))
}

func initConfig(cwd string) {
	readConfigFile(cwd)

	if len(Config.OldNew)%2 == 1 {
		log.Println("Config.Filters: odd argument count. filter list ignored.")
	} else {
		Config.Replacer = strings.NewReplacer(Config.OldNew...)
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
	Config.OldNew = nil // release reference
	Config.TitleMod.WelcomeMsg = ""
}
