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
	titlefi *os.File
	config  = struct {
		CharLimit int `json:"char_limit"`
		EscapeMap []struct {
			Old string `json:"old"`
			New string `json:"new"`
		} `json:"escape_map"`
	}{}
)

func init() {
	// read config file from '<program-name>_config.json'
	bs, err := os.ReadFile(os.Args[0] + ".config.json")
	if err != nil {
		log.Fatal("opening config file: ", err)
	}
	if err := json.Unmarshal(bs, &config); err != nil {
		log.Fatal("reading config file: ", err)
	}

	// open file for writing title at '<program-name>.out'
	fi, err := os.Create(os.Args[0] + ".out")
	if err != nil {
		log.Fatal("could not open/create window-title file:", err)
	}
	titlefi = fi

}

func main() {
	winRecv := i3.Subscribe(i3.WindowEventType)
	for winRecv.Next() {
		ev := winRecv.Event().(*i3.WindowEvent)
		writeWindowTitle(ev.Container.WindowProperties.Title)

		// refresh i3status
		if err := exec.Command("killall", "-USR1", "i3status").Run(); err != nil {
			log.Println("error refreshing i3status:", err)
		}

		// break
	}

	log.Fatal("ending program:", winRecv.Close())
}

func writeWindowTitle(title string) {
	for _, repl := range config.EscapeMap {
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
}
