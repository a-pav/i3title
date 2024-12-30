package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"

	"go.i3wm.org/i3/v4"
)

func main() {
	var (
		lineCh  = make(chan []byte)
		titleCh = make(chan string)
		modeCh  = make(chan string)
	)
	go reporter(lineCh, titleCh, modeCh)

	liner(lineCh)
	go titler(titleCh)
	go moder(modeCh)

	select {} // Block forever.
}

func reporter(lineCh <-chan []byte, titleCh, modeCh <-chan string) {
	var (
		LINE     []byte // LINE comes from `i3status` stdout.
		TITLE    string // TITLE is current window title.
		MODE     string // MODE is current i3 mode.
		MODE_LEN int    // MODE_LEN is visible length of current i3 mode.
		REPORT   string // REPORT is what goes into LINE before printing.
	)
	newReport := func() {
		switch MODE_LEN {
		case 0:
			REPORT = trimTitle(TITLE, cnf.MaxLen)
		default:
			REPORT = MODE + trimTitle(TITLE, cnf.MaxLen-MODE_LEN)
		}
	}

	for {
		select {
		case LINE = <-lineCh:
			// Just print.
		case TITLE = <-titleCh:
			newReport()
		case MODE = <-modeCh:
			switch MODE {
			case "default":
				MODE_LEN = 0
			default:
				MODE_LEN = len(MODE) + cnf.ModeStyleLen
				MODE = fmt.Sprintf(cnf.ModeStyle, MODE)
			}
			newReport()
		}
		// Do print.
		fmt.Fprintf(os.Stdout, "%s\n",
			// Read-only `[]byte(string)` convertions are optimized by compiler:
			// https://github.com/golang/go/issues/2205 (commits=c8adb30,925d2fb,d63c88d).
			bytes.Replace(LINE, []byte(cnf.PH), []byte(REPORT), 1),
		)
	}
}

func moder(modeCh chan<- string) {
	modeER := i3.Subscribe(i3.ModeEventType)

	for modeER.Next() {
		modeCh <- modeER.Event().(*i3.ModeEvent).Change
	}

	log.Fatal("ending program:", modeER.Close())
}

func titler(titleCh chan<- string) {
	windowER := i3.Subscribe(i3.WindowEventType)

	for windowER.Next() {
		e := windowER.Event().(*i3.WindowEvent)
		switch e.Change {
		case "title", "focus":
			titleCh <- e.Container.WindowProperties.Title
		}
	}

	log.Fatal("ending program:", windowER.Close())
}

func liner(lineCh chan<- []byte) {
	// // DEBUG ////////////////////////////////////
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
	// // DEBUG ////////////////////////////////////
	scanner := bufio.NewScanner(os.Stdin)
	if err := scanner.Err(); err != nil {
		log.Fatal("scanner failed to init: ", err)
	}
	// Set maximum buffer size.
	scanner.Buffer(make([]byte, 0, cnf.BufSize), 0)

	// Normal op starts after first 4 lines of output from `i3status`.
	// These look like:
	// 		{"version":1}
	// 		[
	// 		[{"name": ... ]
	// 		,[{"name": ... ]
	for range 4 {
		scanner.Scan()
		lineCh <- scanner.Bytes()
	}

	go func() {
		for scanner.Scan() {
			lineCh <- scanner.Bytes()
		}

		if err := scanner.Err(); err != nil {
			log.Fatal("scanner error:", err)
		}
	}()
}

func replacer(title string) string {
	return cnf.Replacer.Replace(title)
}

// trimTitle cuts title at maxlen, ensuring that it doesn't end with white space,
// and runs the replacer on it.
func trimTitle(title string, maxlen int) string {
	// Note: `len([]rune(string))` pattern is optimized by compiler.
	if len([]rune(title)) > maxlen {
		// This may look cumbersome, but it's clear and easy to maintain.
		// And as shown by the benchmarks, slicing a slice multiple times rather
		// than once, does not affect performance in any meaningful way.
		s := []rune(title)             // alloc.
		s = s[:maxlen]                 // shrink (no alloc.)
		s = s[:lastNonspaceIndex(s)+1] // drop trailing spaces (no alloc.)
		s = append(s, '…')             // append shrinkage indicator (no alloc.)
		title = string(s)              // alloc.
	}

	return replacer(title)
}
