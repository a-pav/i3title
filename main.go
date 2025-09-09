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
		line0     []byte                         // Incoming line from `i3status` stdout.
		line1     = make([]byte, cnf.BufSize)    // Outgoing line with report in it.
		report    = make([]byte, cnf.MaxWidth*5) // Outgoing report (Big enough buffer, even for Chinese characters.)
		reportEnd int                            // Tracks the end of report buffer.
		title     string                         // Current window title.
		mode      string                         // Current i3 mode.
		modeWidth int                            // Width of current i3 mode .
	)
	trimTitle := func(max int) string {
		// NOTE: `len([]rune(string))` pattern is optimized by compiler.
		if len([]rune(title)) > max {
			s := []rune(title)             // alloc.
			s = s[:max]                    // shrink (no alloc.)
			s = s[:lastIndexNonSpace(s)+1] // drop trailing spaces (no alloc.)
			s[max-1] = '…'                 // append shrinkage indicator (no alloc.)
			return replacer(string(s))     // alloc.
		}
		return replacer(title)
	}
	newMode := func() {
		switch mode {
		case "default":
			modeWidth = 0
			mode = ""
		default:
			modeWidth = len(mode) + cnf.ModeStyleWidth
			mode = fmt.Sprintf(cnf.ModeStyle, mode)
		}
	}
	newReport := func() {
		c := 0
		switch modeWidth {
		case 0:
			c += copy(report[0:], trimTitle(cnf.MaxWidth))
		default:
			c += copy(report[0:], mode)
			c += copy(report[c:], trimTitle(cnf.MaxWidth-modeWidth))
		}
		reportEnd = c
	}
	doPrint := func() {
		i := cnf.PHIndex
		if i <= 0 {
			i = bytes.Index(line0, []byte(cnf.PH))
		}
		c := 0
		c += copy(line1[0:], line0[:i])
		c += copy(line1[c:], report[:reportEnd])
		c += copy(line1[c:], line0[i+len(cnf.PH):])

		fmt.Fprintf(os.Stdout, "%s\n", line1[:c])
	}

	// The first two lines don't contain the placeholder and are printed verbatim.
	for range 2 {
		fmt.Fprintf(os.Stdout, "%s\n", <-lineCh)
	}

	for {
		select {
		case line0 = <-lineCh:
			// Just print.
		case title = <-titleCh:
			newReport()
		case mode = <-modeCh:
			newMode()
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
	// This will be inlined.
	return cnf.Replacer.Replace(title)
}
