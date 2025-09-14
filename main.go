package main

import (
	"bufio"
	"bytes"
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
		mode      string                         // Current i3 mode.
		modeWidth int                            // Width of current i3 mode .
		title0    string                         // Current window full title.
		title1    = make([]rune, cnf.MaxWidth)   // Runes of current window title. Helps with counting and less allocation.
		report    = make([]byte, cnf.MaxWidth*5) // Outgoing report (Big enough buffer, even for Chinese characters.)
		reportEnd int                            // Tracks the end of report buffer.
	)
	trimTitle := func(max int) string {
		if len(title0) <= max {
			return replacer(title0)
		}

		s := title1
		i := 0
		for _, r := range title0 {
			if i == max {
				s = s[:max]
				n := lastIndexNonSpace(s) // n will be <= max-1
				if n == max-1 {
					s[n] = '…' // change the last character
				} else { // n < max-1
					s[n+1] = '…'
					s = s[:n+2]
				}
				return replacer(string(s)) // alloc
			}
			s[i] = r
			i++
		}

		return replacer(title0)
	}
	newMode := func() {
		switch mode {
		case "default":
			modeWidth = 0
		default:
			modeWidth = len([]rune(mode)) + cnf.ModeStyleWidth
		}
	}
	newReport := func() {
		c := 0
		if modeWidth > 0 {
			i := cnf.ModeStyleIndex
			c += copy(report[c:], cnf.ModeStyle[:i])
			c += copy(report[c:], mode)
			c += copy(report[c:], cnf.ModeStyle[i+2:]) // 2 == len("%s")
		}
		c += copy(report[c:], trimTitle(cnf.MaxWidth-modeWidth))

		reportEnd = c
	}
	doPrint := func() {
		i := cnf.PHIndex
		if i <= 0 {
			i = bytes.Index(line0, []byte(cnf.PH))
		}
		c := 0
		c += copy(line1[c:], line0[:i])
		c += copy(line1[c:], report[:reportEnd])
		c += copy(line1[c:], line0[i+len(cnf.PH):])
		c += copy(line1[c:], "\n")

		os.Stdout.Write(line1[:c])
	}

	// The first two lines don't contain the placeholder and are printed verbatim.
	for range 2 {
		os.Stdout.Write(append(<-lineCh, byte('\n')))
	}

	ok := true
	for {
		select {
		case line0 = <-lineCh:
			// Just print.
		case title0 = <-titleCh:
			newReport()
		case mode, ok = <-modeCh:
			if !ok {
				modeCh = nil // disable
				continue
			}
			newMode()
			newReport()
		}

		doPrint()
	}
}

func moder(modeCh chan<- string) {
	if cnf.ModeStyleIndex < 0 {
		close(modeCh)
		return
	}
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
