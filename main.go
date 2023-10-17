package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

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
	// Ideally, we want to update statusbar upon each change-of-title event. But,
	// i3 creates too many of those events in less than a second while system is
	// initially starting. To avoid errors, it's better to `discard` a few of those
	// initial events WITHOUT writing to stdout (updating statusbar).
	time.Sleep(time.Duration(Config.TiMod.Delay) * time.Second)

	winRecv := i3.Subscribe(i3.WindowEventType)

	for winRecv.Next() {
		ev := winRecv.Event().(*i3.WindowEvent)
		TITLE = makeTitle(ev.Container.WindowProperties.Title)

		printline()
		// There's no need to signal i3status to refresh. It picks on the stdout by itself.
	}

	log.Fatal("ending program:", winRecv.Close())
}

func readLine() {
	// DEBUG
	// cmd := exec.Command("i3status")
	// stdout, err := cmd.StdoutPipe()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// if err := cmd.Start(); err != nil {
	// 	log.Fatal(err)
	// }
	// defer cmd.Process.Release()
	// scanner := bufio.NewScanner(stdout)

	scanner := bufio.NewScanner(os.Stdin)
	if err := scanner.Err(); err != nil {
		log.Fatal("scanner failed to init: ", err)
	}
	// Set maximum buffer size.
	scanner.Buffer(make([]byte, 0, Config.BufSize), 0)

	for scanner.Scan() {
		LINE = scanner.Bytes()
		printline()
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("scanner error:", err)
	}
}

// makeTitle applies the defined filters, maxlen, format, etc. to title.
func makeTitle(title string) string {
	title = Config.Replacer.Replace(title)
	// Convert title string to runes, because unicode characters can have length > 1
	// when they're actually one single rune.
	// example:
	//	 str := "·—"
	//	 fmt.Println(len(str)) // prints 5
	//	 fmt.Println(len([]rune(str))) // prints 2
	//
	// `len([]rune(s))` pattern is optimized by compiler.
	if len([]rune(title)) > Config.TiMod.MaxLen {
		rs := []rune(title)                                 // alloc.
		rs = rs[:Config.TiMod.MaxLen]                       // shrink (no alloc.)
		rs = rs[:lastSafeIndex(rs, Config.TiMod.MaxEscLen)] // drop possible half-fromed escape sequence (no alloc.)
		rs = rs[:lastNonspaceIndex(rs)+1]                   // drop possible trailing spaces (no alloc.)
		rs = append(rs, '…')                                // append shrinkage indicator (no alloc.)
		title = string(rs)                                  // alloc.
	}

	return titlef(title)
}

// printline inserts `TITLE` into `LINE` (the coming stdin) then prints the result to stdout.
func printline() {
	fmt.Fprintf(os.Stdout, "%s\n",
		bytes.Replace(LINE, []byte(Config.TiMod.PH), []byte(TITLE), 1),
	)
}
