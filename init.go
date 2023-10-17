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

	TITLE = titlef(fmt.Sprintf("Delay: %ds", Config.TiMod.Delay))
}

func initConfig(cwd string) {
	readConfigFile(cwd)

	if len(Config.OldNew)%2 == 1 {
		log.Println("Config.Filters: odd argument count. filter list ignored.")
	} else {
		// Initialize strings.Replacer.
		Config.Replacer = strings.NewReplacer(Config.OldNew...)
		// Determine the biggest escape sequence width.
		w, oldnew := 0, Config.OldNew
		for i := 0; i < len(oldnew); i += 2 {
			if l := len(oldnew[i+1]); l > 0 && oldnew[i+1][0] == '&' {
				w = max(w, l)
			}
		}
		Config.TiMod.MaxEscLen = w
	}
}

func readConfigFile(cwd string) {
	// read config file from current working directory.
	bs, err := os.ReadFile(cwd + "/config.json")
	if err != nil {
		log.Fatal("opening config file: ", err)
	}

	// strip comments.
	bs = regexp.MustCompile(`//.*`).ReplaceAll(bs, nil)

	if err := json.Unmarshal(bs, &Config); err != nil {
		log.Fatal("reading config file: ", err)
	}
}

// discardConfig discards parts of the config that are no longer needed.
func discardConfig() {
	Config.OldNew = nil // release reference
}
