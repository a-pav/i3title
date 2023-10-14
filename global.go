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
	TITLE string
	LINE  []byte

	Config = struct {
		Debug    bool   `json:"debug"`
		BufSize  uint16 `json:"buffer_size"` // max: 65535, 64KB
		TitleMod struct {
			PH         string `json:"placeholder"`
			Format     string `json:"format"`
			MaxLen     int    `json:"max_length"`
			Delay      uint8  `json:"delay"`
			WelcomeMsg string `json:"welcome_msg"` // It is shown until the first window/title event is triggered.
		} `json:"title_module"`
		Filters []string `json:"filters"`

		FiltersCompiled []MatchReplace
	}{}
)
