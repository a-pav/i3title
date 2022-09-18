package main

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"sync"

	"go.i3wm.org/i3/v4"
)

var (
	titlefi         *os.File
	i3statusRefresh *exec.Cmd
	config          = struct {
		CharLimit int               `json:"char_limit"`
		EscapeMap map[string]string `json:"escape_map"`
	}{}
)

func init() {
	// read config file from '<program-name>_config.json'
	bs, err := os.ReadFile(os.Args[0] + "_config.json")
	if err != nil {
		log.Fatal("opening config file: ", err)
	}
	if err := json.Unmarshal(bs, &config); err != nil {
		log.Fatal("reading config file: ", err)
	}

	// create a command for refreshing i3status
	i3statusRefresh = exec.Command("killall", "-USR1", "i3status")

	// open file for writing title at '<program-name>.out'
	fi, err := os.Create(os.Args[0] + ".out")
	if err != nil {
		log.Fatal("could not open/create window-title file:", err)
	}
	titlefi = fi

}

func main() {
	winRecv := i3.Subscribe(i3.WindowEventType)
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		for winRecv.Next() {
			ev := winRecv.Event().(*i3.WindowEvent)
			writeWindowTitle(ev.Container.WindowProperties.Title)
			i3statusRefresh.Run()
			// break
		}

		wg.Done()
	}()

	wg.Wait()

	log.Fatal("ending program:", winRecv.Close())
}

func writeWindowTitle(title string) {
	if len(title) > config.CharLimit {
		title = title[:config.CharLimit] + "..."
	}

	var titleEsc string
	for _, r := range title {
		if repl, ok := config.EscapeMap[string(r)]; ok {
			titleEsc += repl
		} else {
			titleEsc += string(r)
		}
	}

	if err := titlefi.Truncate(0); err != nil {
		log.Println("truncating 'window-title' file:", err)
	}
	if _, err := titlefi.Seek(0, 0); err != nil {
		log.Println("seeking 'window-title' file:", err)

	}
	if _, err := titlefi.Write([]byte(titleEsc)); err != nil {
		log.Println("writing 'window-title' file:", err)
	}
}
