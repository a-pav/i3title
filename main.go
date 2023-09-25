package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"go.i3wm.org/i3/v4"
)

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

		for Scanner.Scan() {
			LINE = Scanner.Text()
			// line = scanner.Text()
			writeStdOut()
		}

		if err := Scanner.Err(); err != nil {
			log.Fatal("scanner error:", err)
		}
	}()

	wg.Wait()
}

// trimTitle applies the defined filters, maxlen, etc.
func trimTitle(title string) string {
	for _, filter := range Config.FiltersCompiled {
		title = filter.Match.ReplaceAllString(title, filter.Repl)
	}

	// Convert title string to runes, because unicode characters can have length > 1
	// when they're actually one single rune.
	// example:
	//	 str := "·—"
	//	 fmt.Println(len(str)) // prints 5
	//	 fmt.Println(len([]rune(str))) // prints 2
	titleRunes := []rune(title)
	if len(titleRunes) > Config.TitleModule.MaxLen {
		title = string(titleRunes[:Config.TitleModule.MaxLen]) + "…"
	}

	switch Config.Debug {
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

	sm[Config.TitleModule.Index]["full_text"] = fmt.Sprintf(Config.TitleModule.Format, TITLE)

	j, err := json.Marshal(sm)
	if err != nil {
		log.Fatal("failure encoding line:", err)
	}
	fmt.Fprintf(os.Stdout, ",%s\n", string(j))
}
