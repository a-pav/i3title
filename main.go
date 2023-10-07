package main

import (
	"bufio"
	"bytes"
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
		readLine()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		readTitle()
	}()

	wg.Wait()
}

func readTitle() {
	winRecv := i3.Subscribe(i3.WindowEventType)

	for winRecv.Next() {
		ev := winRecv.Event().(*i3.WindowEvent)
		TITLE = trimTitle(ev.Container.WindowProperties.Title)

		printline()
		// There's no need to signal i3status to refresh. It picks on the stdout by itself.

		// break
	}

	log.Fatal("ending program:", winRecv.Close())
}

func readLine() {
	scanner := bufio.NewScanner(os.Stdin)
	if err := scanner.Err(); err != nil {
		log.Fatal("scanner failed to init: ", err)
	}

	// Redirect stdin to stdout until a valid i3bar array line is reached.
	// i3status' first few lines are NOT a valid array line. They usually look
	// like a `{"version":1}`, a `[` and possible errors it ran into during startup.
	// This loop is to skip them all.
	for scanner.Scan() {
		LINE = scanner.Bytes()
		if bytes.HasPrefix(LINE, []byte(",[{\"")) { // this is our cue that i3status has started printing valid array lines.
			printline()
			break // break to get rid of this check.
		}

		fmt.Fprintf(os.Stdout, "%s\n", LINE)
	}

	for scanner.Scan() {
		LINE = scanner.Bytes()
		printline()
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("scanner error:", err)
	}
}

// trimTitle applies the defined filters, maxlen, format, etc. to title.
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
	if titleRunes := []rune(title); len(titleRunes) > Config.TitleModule.MaxLen {
		title = strings.TrimSpace(string(titleRunes[:Config.TitleModule.MaxLen])) + "…"
	}

	switch Config.Debug {
	case true:
		go log.Printf("%q", title)
	}

	return fmt.Sprintf(Config.TitleModule.Format, title)
}

// printline inserts `TITLE` into `LINE` (the coming stdin) then prints the result to stdout.
func printline() {
	sm := []map[string]any{}
	// bytes.TrimPrefix(LINE, []byte(","))
	if err := json.Unmarshal(LINE[1:], &sm); err != nil {
		log.Fatal("failure parsing line:", err)
	}

	sm[Config.TitleModule.Index]["full_text"] = TITLE // insert

	j, err := json.Marshal(sm)
	if err != nil {
		log.Fatal("failure encoding line:", err)
	}

	// if _, err := fmt.Fprintf(os.Stdout, ",%s\n", j); err != nil {
	// 	log.Fatal("failure writing stdout:", err)
	// }

	os.Stdout.Write(append([]byte(","), append(j, byte('\n'))...))
}
