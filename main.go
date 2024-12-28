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
	lineCh := make(chan []byte)
	go liner(lineCh)

	titleCh := make(chan string)
	go titler(titleCh)

	modeCh := make(chan string)
	go moder(modeCh)

	// select {} // Block forever.
	reporter(lineCh, titleCh, modeCh)
}

func reporter(lineCh chan []byte, titleCh, modeCh chan string) {
	var (
		LINE     []byte // LINE comes from `i3status` stdout.
		TITLE    string // TITLE is current windows title.
		MODE     string // MODE is the current i3 mode.
		MODE_LEN int    // MODE_LEN is the visible length of current i3 mode.
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
			// typical.
		case TITLE = <-titleCh:
			newReport()
		case MODE = <-modeCh:
			switch MODE {
			case "default":
				MODE_LEN = 0
			default:
				MODE_LEN = len(MODE) + MODE_SEP_LEN
				MODE = fmt.Sprintf(
					"<span color='red' font='italic bold'>%s</span>%s",
					MODE, MODE_SEP,
				)
			}
			newReport()
		}

		// do print
		fmt.Fprintf(os.Stdout, "%s\n",
			// Read-only `[]byte(string)` convertions are optimized by compiler:
			// https://github.com/golang/go/issues/2205 (commits=c8adb30,925d2fb,d63c88d).
			bytes.Replace(LINE, []byte(cnf.PH), []byte(REPORT), 1),
		)

	}
}

func moder(modeCh chan string) {
	modeER := i3.Subscribe(i3.ModeEventType)

	for modeER.Next() {
		modeCh <- modeER.Event().(*i3.ModeEvent).Change
	}

	log.Fatal("ending program:", modeER.Close())
}

func titler(titleCh chan string) {
	// TODO: Remove: After commit 57ea2c088 there might be no need to delay.
	// if cnf.StartDelay > 0 {
	// 	// i3 creates too many change-of-title events in a row while system and/or
	// 	// i3 itself is initially starting. To avoid errors, it's best not to
	// 	// subscribe to the events too early.
	// 	REPORT = fmt.Sprintf("<i>i3title start delay: %ds</i>", cnf.StartDelay)
	// 	time.Sleep(time.Duration(cnf.StartDelay) * time.Second)
	// 	// Sudden empty title shuold indicate that normal operation has started.
	// 	REPORT = ""
	// }

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

func liner(lineCh chan []byte) {
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

	for scanner.Scan() {
		lineCh <- scanner.Bytes()
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("scanner error:", err)
	}
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
