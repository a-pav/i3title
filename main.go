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
		_LINE     []byte
		_TITLE    string
		_MODE     string
		_MODE_LEN int
		_REPORT   string
	)

	newReport := func() {
		switch _MODE_LEN {
		case 0:
			_REPORT = trimTitle(_TITLE, cnf.MaxLen)
		default:
			_REPORT = _MODE + trimTitle(_TITLE, cnf.MaxLen-_MODE_LEN)
		}
	}

	for {
		select {
		case _LINE = <-lineCh:
			// typical.
		case _TITLE = <-titleCh:
			newReport()
		case _MODE = <-modeCh:
			switch _MODE {
			case "default":
				// _MODE = ""
				_MODE_LEN = 0
			default:
				// == len(<mode-name>) + len(visible_sep_chars)
				_MODE_LEN = len(_MODE) + MODE_SEP_LEN
				_MODE = fmt.Sprintf(
					"<span color='red' font='italic bold'>%s</span>%s",
					_MODE, MODE_SEP,
				)
			}
			newReport()
		}

		// do print
		fmt.Fprintf(os.Stdout, "%s\n",
			// Read-only `[]byte(string)` convertions are optimized by compiler:
			// https://github.com/golang/go/issues/2205 (commits=c8adb30,925d2fb,d63c88d).
			bytes.Replace(_LINE, []byte(cnf.PH), []byte(_REPORT), 1),
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
		s = s[:lastNonSpaceIndex(s)+1] // drop trailing spaces (no alloc.)
		s = append(s, '…')             // append shrinkage indicator (no alloc.)
		title = string(s)              // alloc.
	}

	return replacer(title)
}
