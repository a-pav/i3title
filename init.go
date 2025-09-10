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

	if len(cnf.OldNew)%2 == 1 {
		log.Println("Config.Filters: odd argument count. filter list ignored.")
	} else {
		// Initialize strings.Replacer.
		cnf.Replacer = strings.NewReplacer(cnf.OldNew...)
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

	if err := json.Unmarshal(bs, &cnf); err != nil {
		log.Fatal("reading config file: ", err)
	}
}

// discardConfig discards parts of the config that are no longer needed.
func discardConfig() {
	cnf.OldNew = nil // release reference
}
