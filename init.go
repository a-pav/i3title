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
// Using os.Args[0] is the surest way to get the actual CWD. i3 seems to run
// everything from /home/$USER.
func getwd() string { return filepath.Dir(os.Args[0]) }

func initConfig(cwd string) {
	readConfigFile(cwd)

	// Default oldnew pairs.
	// https://en.wikipedia.org/wiki/List_of_XML_and_HTML_character_entity_references
	// https://en.wikipedia.org/wiki/Numeric_character_reference
	oldnew := []string{
		"&", "&amp;",
		">", "&gt;",
		"<", "&lt;",
		"\"", "&#34;", // "&#34;" is shorter than "&quot;".
		"\\", "&#92;", // "&#92;" is shorter than "&Backslash;", or anything else.
	}
	// Initialize strings.Replacer.
	if l := len(cnf.OldNew); l == 0 {
		// no op
	} else if l%2 == 1 {
		log.Println(`cnf.OldNew: odd number of arguments, "old_new" list is ignored.`)
	} else {
		oldnew = append(oldnew, cnf.OldNew...)
	}
	cnf.Replacer = strings.NewReplacer(oldnew...)
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
