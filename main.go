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
	if cnf.StartDelay > 0 {
		// Ideally, we want to update statusbar upon each change-of-title event.
		// But i3 creates too many of those events in less than a second while
		// system and/or i3 itself is initially starting. To avoid errors, it's
		// best to ignore first few initial events and not write to stdout (i.e. update statusbar).
		TITLE = fmt.Sprintf("i3title start delay: %ds", cnf.StartDelay)
		time.Sleep(time.Duration(cnf.StartDelay) * time.Second)
		// Sudden empty title shuold indicate that normal operation has started.
		TITLE = ""
	}

	winRecv := i3.Subscribe(i3.WindowEventType)

	for winRecv.Next() {
		ev := winRecv.Event().(*i3.WindowEvent)
		TITLE = makeTitle(ev.Container.WindowProperties.Title)

		printline()
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
	scanner.Buffer(make([]byte, 0, cnf.BufSize), 0)

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
	title = cnf.Replacer.Replace(title)
	// Note: `len([]rune(string))` pattern is optimized by compiler.
	if len([]rune(title)) > cnf.MaxLen {
		// This may look cumbersome, but it's clear and easy to maintain.
		// And as shown by the benchmarks, slicing a slice multiple times rather
		// than once, does not affect performance in any meaningful way.
		s := []rune(title)                           // alloc.
		s = s[:cnf.MaxLen]                           // shrink (no alloc.)
		s = s[:lastNonEscapeIndex(s, cnf.MaxEscLen)] // drop trailing half-fromed escape sequence (no alloc.)
		s = s[:lastNonSpaceIndex(s)+1]               // drop trailing spaces (no alloc.)
		s = append(s, '…')                           // append shrinkage indicator (no alloc.)
		title = string(s)                            // alloc.
	}

	return title
}

// printline inserts `TITLE` into `LINE` (the coming stdin) then prints the result to stdout.
func printline() {
	fmt.Fprintf(os.Stdout, "%s\n",
		// Read-only `[]byte(string)` convertions are optimized by compiler:
		// https://github.com/golang/go/issues/2205 (commits=c8adb30,925d2fb,d63c88d).
		bytes.Replace(LINE, []byte(cnf.PH), []byte(TITLE), 1),
	)
}
