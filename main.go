package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"time"

	"go.i3wm.org/i3/v4"
)

func main() {
	go readLine()
	go readTitle()
	go readMode()

	select {} // Block forever.
}

func readMode() {
	modeER := i3.Subscribe(i3.ModeEventType)

	for modeER.Next() {
		e := modeER.Event().(*i3.ModeEvent)
		switch e.Change {
		case "default":
			MODE = ""
			MODE_LEN = 0
		default:
			switch e.PangoMarkup {
			case false:
				MODE = fmt.Sprintf(
					"<span color='red' font='italic bold'>%s</span>%s",
					e.Change, MODE_SEP,
				)
			default:
				MODE = fmt.Sprintf("%s%s", e.Change, MODE_SEP)
			}
			// == len(<mode-name>) + len(visible_sep_chars)
			MODE_LEN = len(e.Change) + MODE_SEP_LEN
		}
		buildReport()
		printline()
	}

	log.Fatal("ending program:", modeER.Close())
}

func readTitle() {
	// TODO: Remove: After commit 57ea2c088 there might be no need to delay.
	if cnf.StartDelay > 0 {
		// i3 creates too many change-of-title events in a row while system and/or
		// i3 itself is initially starting. To avoid errors, it's best not to
		// subscribe to the events too early.
		REPORT = fmt.Sprintf("<i>i3title start delay: %ds</i>", cnf.StartDelay)
		time.Sleep(time.Duration(cnf.StartDelay) * time.Second)
		// Sudden empty title shuold indicate that normal operation has started.
		REPORT = ""
	}

	windowER := i3.Subscribe(i3.WindowEventType)

	for windowER.Next() {
		e := windowER.Event().(*i3.WindowEvent)
		switch e.Change {
		case "title", "focus":
		default:
			continue
		}

		if t := e.Container.WindowProperties.Title; t != "" && TITLE != t {
			TITLE = t
			buildReport()
			printline()
		}
	}

	log.Fatal("ending program:", windowER.Close())
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
		s = s[:lastNonSpaceIndex(s)+1] // drop trailing spaces (no alloc.)
		s = append(s, '…')             // append shrinkage indicator (no alloc.)
		title = string(s)              // alloc.
	}

	return replacer(title)
}

func buildReport() {
	switch MODE {
	case "":
		REPORT = trimTitle(TITLE, cnf.MaxLen)
	default:
		REPORT = MODE + trimTitle(TITLE, cnf.MaxLen-MODE_LEN)
	}
}

// printline inserts `REPORT` into `LINE` (incoming stdin) then prints it to stdout.
func printline() {
	fmt.Fprintf(os.Stdout, "%s\n",
		// Read-only `[]byte(string)` convertions are optimized by compiler:
		// https://github.com/golang/go/issues/2205 (commits=c8adb30,925d2fb,d63c88d).
		bytes.Replace(LINE, []byte(cnf.PH), []byte(REPORT), 1),
	)
}
