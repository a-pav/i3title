package main

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"strings"

	"go.i3wm.org/i3/v4"
)

var (
	args0   string // program pathname
	titlefi *os.File
	config  = struct {
		LogTitle  bool `json:"log_title"`
		CharLimit int  `json:"char_limit"`
		Filters   []struct {
			Old string `json:"old"`
			New string `json:"new"`
		} `json:"filters"`
	}{}
)

func init() {
	args0 = os.Args[0]
	// read config file from '<program-name>_config.json'
	bs, err := os.ReadFile(args0 + ".config.json")
	if err != nil {
		log.Fatal("opening config file: ", err)
	}
	if err := json.Unmarshal(bs, &config); err != nil {
		log.Fatal("reading config file: ", err)
	}

	// open file for writing title at '<program-name>.out'
	fi, err := os.Create(args0 + ".out")
	if err != nil {
		log.Fatal("could not open/create window-title file:", err)
	}
	titlefi = fi
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

	for _, repl := range config.Filters {
		title = strings.ReplaceAll(title, repl.Old, repl.New)
	}

	// convert title string to runes, because unicode characters can have length > 1
	// when they're actually one single rune.
	// example:
	//	 str := "·—"
	//	 fmt.Println(len(str)) // prints 5
	//	 fmt.Println(len([]rune(str))) // prints 2
	titleRunes := []rune(title)
	if len(titleRunes) > config.CharLimit {
		title = string(titleRunes[:config.CharLimit]) + "..."
	}

	if err := titlefi.Truncate(0); err != nil {
		log.Println("truncating 'window-title' file:", err)
	}
	if _, err := titlefi.Seek(0, 0); err != nil {
		log.Println("seeking 'window-title' file:", err)

	}
	if _, err := titlefi.Write([]byte(title)); err != nil {
		log.Println("writing 'window-title' file:", err)
	}

	switch config.LogTitle {
	case true:
		go log.Printf("%q", title)
	}
}

func i3StatusRefresh() {
	if err := exec.Command("killall", "-USR1", "i3status").Run(); err != nil {
		log.Println("error refreshing i3status:", err)
	}
}
