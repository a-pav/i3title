package main

import (
	"regexp"
)

type MatchReplace struct {
	Match *regexp.Regexp
	Repl  string
}

var (
	// Globally shared variables.
	// For the sake of this program, these should work fine with no mutex mechanism in place.
	TITLE = ""
	LINE  = ""

	Config = struct {
		Debug       bool `json:"debug"`
		TitleModule struct {
			Index      int    `json:"index"`
			Format     string `json:"format"`
			MaxLen     int    `json:"max_length"`
			WelcomeMsg string `json:"welcome_msg"` // It is shown until the first window/title event is triggered.
		} `json:"title_module"`

		Filters [][]string `json:"filters"`

		FiltersCompiled []MatchReplace
	}{}
)
