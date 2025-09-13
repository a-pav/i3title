package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func init() {
	defer discardConfig()
	log.SetPrefix("i3title: ")
	log.SetFlags(log.Lmsgprefix)

	initConfig(getwd())
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

func initConfig(cwd string) {
	readConfigFile(cwd)

	// Strip styling tags, attrs and and any char that doesn't add to the width.
	raw := regexp.MustCompile("</?[^>]+>").ReplaceAllString(cnf.ModeStyle, "")
	cnf.ModeStyleWidth = len(raw) - len("%s")

	cnf.ModeStyleIndex = strings.Index(cnf.ModeStyle, "%s")

	if len(cnf.OldNew)%2 == 1 {
		log.Println(`cnf.OldNew: odd number of arguments, "old_new" list is ignored.`)
	} else {
		// Initialize strings.Replacer.
		cnf.Replacer = strings.NewReplacer(cnf.OldNew...)
	}
}

func readConfigFile(cwd string) {
	// Read config file from current working directory.
	bs, err := os.ReadFile(cwd + "/config.json")
	if err != nil {
		log.Fatal("opening config file: ", err)
	}

	// Strip comments.
	bs = regexp.MustCompile(`(?m)^\s*//.*$`).ReplaceAll(bs, nil)

	if err := json.Unmarshal(bs, &cnf); err != nil {
		log.Fatal("reading config file: ", err)
	}
}

// discardConfig discards parts of the config that are no longer needed.
func discardConfig() {
	cnf.OldNew = nil // release reference
}
