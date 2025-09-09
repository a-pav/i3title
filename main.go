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
		line0    []byte                      // Incoming line from `i3status` stdout.
		line1    = make([]byte, cnf.BufSize) // Outgoing line with report in it.
		TITLE    string                      // TITLE is current window title.
		MODE     string                      // MODE is current i3 mode.
		MODE_LEN int                         // MODE_LEN is visible length of current i3 mode.
		REPORT   string                      // REPORT is what goes into LINE before printing.
	)
	newReport := func() {
		switch MODE_LEN {
		case 0:
			REPORT = trimTitle(TITLE, cnf.MaxWidth)
		default:
			REPORT = MODE + trimTitle(TITLE, cnf.MaxWidth-MODE_LEN)
		}
	}
	doPrint := func() {
		i := cnf.PHIndex
		if i <= 0 {
			i = bytes.Index(line0, []byte(cnf.PH))
		}
		copy(line1[0:], line0[:i])
		copy(line1[i:], REPORT)
		copy(line1[i+len(REPORT):], line0[i+len(cnf.PH):])

		cut := len(line0) + len(REPORT) - len(cnf.PH)
		out := line1[:cut]

		fmt.Fprintf(os.Stdout, "%s\n", out)
	}

	// The first two lines don't contain the placeholder and are printed verbatim.
	for range 2 {
		fmt.Fprintf(os.Stdout, "%s\n", <-lineCh)
	}

	for {
		select {
		case line0 = <-lineCh:
			// Just print.
		case TITLE = <-titleCh:
			newReport()
		case MODE = <-modeCh:
			switch MODE {
			case "default":
				MODE_LEN = 0
				MODE = ""
			default:
				MODE_LEN = len(MODE) + cnf.ModeStyleWidth
				MODE = fmt.Sprintf(cnf.ModeStyle, MODE)
			}
			newReport()
		}

		doPrint()
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
