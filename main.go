package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"

	"go.i3wm.org/i3/v4"
)

type MatchReplace struct {
	Match *regexp.Regexp
	Repl  string
}

var (
	scanner *bufio.Scanner

	// Globally shared variables.
	// For the sake of this program, these should work fine with no mutex mechanism in place.
	TITLE = ""
	LINE  = ""

	config = struct {
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

func readConfig(args0 string) {
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
}

// discardConfig discards parts of the config that are no longer needed.
func discardConfig() {
	config.Filters = nil // discard
	config.TitleModule.WelcomeMsg = ""
}

func initConfig(args0 string) {
	readConfig(args0)

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
}

func initScanner() {
	scanner = bufio.NewScanner(os.Stdin)

	scanner.Scan() // skip 1st line: {"version":1}
	fmt.Fprintf(os.Stdout, "%s\n", scanner.Text())
	scanner.Scan() // skip 2nd line: [
	fmt.Fprintf(os.Stdout, "%s\n", scanner.Text())
	scanner.Scan() // skip 3rd line: the only line without ',' as delimiter
	fmt.Fprintf(os.Stdout, "%s\n", scanner.Text())
}

func init() {
	defer discardConfig()

	args0 := os.Args[0]

	initConfig(args0)
	initScanner()

	TITLE = config.TitleModule.WelcomeMsg
}

func main() {
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		winRecv := i3.Subscribe(i3.WindowEventType)
		for winRecv.Next() {
			ev := winRecv.Event().(*i3.WindowEvent)
			TITLE = trimTitle(ev.Container.WindowProperties.Title)

			// There's no need to signal i3status to refresh. It picks on the stdout by itself.
			writeStdOut()

			// break
		}

		log.Fatal("ending program:", winRecv.Close())
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		for scanner.Scan() {
			LINE = scanner.Text()
			// line = scanner.Text()
			writeStdOut()
		}

		if err := scanner.Err(); err != nil {
			log.Fatal("scanner error:", err)
		}
	}()

	wg.Wait()
}

// trimTitle applies the defined filters, maxlen, etc.
func trimTitle(title string) string {
	for _, filter := range config.FiltersCompiled {
		title = filter.Match.ReplaceAllString(title, filter.Repl)
	}

	// Convert title string to runes, because unicode characters can have length > 1
	// when they're actually one single rune.
	// example:
	//	 str := "·—"
	//	 fmt.Println(len(str)) // prints 5
	//	 fmt.Println(len([]rune(str))) // prints 2
	titleRunes := []rune(title)
	if len(titleRunes) > config.TitleModule.MaxLen {
		title = string(titleRunes[:config.TitleModule.MaxLen]) + "..."
	}

	switch config.Debug {
	case true:
		go log.Printf("%q", title)
	}

	return title
}

// writeStdOut places `TITLE` into the coming stdin/`LINE` then writes it to stdout.
func writeStdOut() {
	sm := []map[string]any{}
	if err := json.Unmarshal([]byte(strings.TrimPrefix(LINE, ",")), &sm); err != nil {
		log.Fatal("failure parsing line:", err)
	}
	sm[config.TitleModule.Index]["full_text"] = fmt.Sprintf(config.TitleModule.Format, TITLE)

	j, err := json.Marshal(sm)
	if err != nil {
		log.Fatal("failure encoding line:", err)
	}
	fmt.Fprintf(os.Stdout, ",%s\n", string(j))
}
