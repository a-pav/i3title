package main

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"regexp"

	"go.i3wm.org/i3/v4"
)

type MatchReplace struct {
	Match *regexp.Regexp
	Repl  string
}

var (
	titlefi *os.File
	config  = struct {
		Debug   bool       `json:"debug"`
		MaxLen  int        `json:"max_length"`
		Filters [][]string `json:"filters"`

		FiltersCompiled []MatchReplace
	}{}
)

func initReadConfig(args0 string) {
	// read config file from '<program-name>.config.json'
	bs, err := os.ReadFile(args0 + ".config.json")
	if err != nil {
		log.Fatal("opening config file: ", err)
	}

	// strip comments
	bs = regexp.MustCompile(`//.*`).ReplaceAll(bs, nil)

	if err := json.Unmarshal(bs, &config); err != nil {
		log.Fatal("reading config file: ", err)
	}

	filtersCompiled := []MatchReplace{}
	for _, mr := range config.Filters {
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

	config.FiltersCompiled = filtersCompiled
	config.Filters = nil // discard
}

func initOpenTitleFile(args0 string) {
	// open file for writing title at '<program-name>.out'
	fi, err := os.Create(args0 + ".out")
	if err != nil {
		log.Fatal("could not open/create window-title file:", err)
	}
	titlefi = fi
}

func init() {
	args0 := os.Args[0]

	initReadConfig(args0)
	initOpenTitleFile(args0)
}

func main() {
	winRecv := i3.Subscribe(i3.WindowEventType)
	for winRecv.Next() {
		ev := winRecv.Event().(*i3.WindowEvent)
		writeTitle(ev.Container.WindowProperties.Title)
		i3StatusRefresh()

		// break
	}

	log.Fatal("ending program:", winRecv.Close())
}

func writeTitle(title string) {
	for _, filter := range config.FiltersCompiled {
		title = filter.Match.ReplaceAllString(title, filter.Repl)
	}

	// convert title string to runes, because unicode characters can have length > 1
	// when they're actually one single rune.
	// example:
	//	 str := "·—"
	//	 fmt.Println(len(str)) // prints 5
	//	 fmt.Println(len([]rune(str))) // prints 2
	titleRunes := []rune(title)
	if len(titleRunes) > config.MaxLen {
		title = string(titleRunes[:config.MaxLen]) + "..."
	}

	if err := titlefi.Truncate(0); config.Debug && err != nil {
		log.Println("truncating 'window-title' file:", err)
	}
	if _, err := titlefi.Seek(0, 0); config.Debug && err != nil {
		log.Println("seeking 'window-title' file:", err)

	}
	if _, err := titlefi.Write([]byte(title)); config.Debug && err != nil {
		log.Println("writing 'window-title' file:", err)
	}

	switch config.Debug {
	case true:
		go log.Printf("%q", title)
	}
}

func i3StatusRefresh() {
	if err := exec.Command("killall", "-USR1", "i3status").Run(); config.Debug && err != nil {
		log.Println("error refreshing i3status:", err)
	}
}
