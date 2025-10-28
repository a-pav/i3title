package main

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"syscall"

	"go.i3wm.org/i3/v4"
)

func reporter(lineCh, messageCh <-chan []byte, titleCh, modeCh <-chan string) {
	var (
		line0     []byte                         // Incoming line from `i3status` stdout.
		line1     = make([]byte, cnf.BufSize)    // Outgoing line with report in it.
		mode      = "default"                    // Current i3 mode.
		title0    string                         // Current window full title.
		title1    = make([]rune, cnf.MaxWidth)   // Runes of current window title. Helps with rune counting and less allocation.
		report    = make([]byte, cnf.MaxWidth*5) // Outgoing report (Big enough buffer, even for Chinese characters.)
		reportEnd int                            // Tracks the end of report buffer.
		message   []byte                         // Overwrites the report.
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
	newReport := func() {
		if len(message) > 0 {
			return // report doesn't update unless message is cleared.
		}

		c := 0
		m := 0 // mode visible length.
		if mode != "default" {
			m = len([]rune(mode)) + cnf.ModeStyleWidth

			i := cnf.ModeStyleIndex
			c += copy(report[c:], cnf.ModeStyle[:i])
			c += copy(report[c:], mode)
			c += copy(report[c:], cnf.ModeStyle[i+2:]) // 2 == len("%s")
		}
		c += copy(report[c:], trimTitle(cnf.MaxWidth-m))

		reportEnd = c

		doPrint()
	}
	newMessage := func() {
		if len(message) == 0 { // clearing message?
			newReport()
			return
		}
		c := 0
		c += copy(report[c:], message)
		reportEnd = c
		doPrint()
	}

	// The first two lines don't contain the placeholder and are printed verbatim.
	for range 2 {
		line0 = <-lineCh

		c := 0
		c += copy(line1[c:], line0)
		c += copy(line1[c:], "\n")

		os.Stdout.Write(line1[:c])
	}

	for {
		var ok bool
		select {
		case line0 = <-lineCh:
			doPrint() // just print
		case title0 = <-titleCh:
			newReport()
		case mode, ok = <-modeCh:
			if !ok {
				modeCh = nil // disable
				continue
			}
			newReport()
		case message, ok = <-messageCh:
			if !ok {
				messageCh = nil // disable
				continue
			}
			newMessage()
		}
	}
}

func emitModes(modeCh chan<- string) {
	if cnf.ModeStyleIndex < 0 {
		close(modeCh)
		return
	}
	modeER := i3.Subscribe(i3.ModeEventType)

	for modeER.Next() {
		modeCh <- modeER.Event().(*i3.ModeEvent).Change
	}

	log.Printf("WARNING: no more mode event: %v", modeER.Close())
}

func emitTitles(titleCh chan<- string) {
	windowER := i3.Subscribe(i3.WindowEventType)

	for windowER.Next() {
		e := windowER.Event().(*i3.WindowEvent)
		switch e.Change {
		case "title", "focus":
			titleCh <- e.Container.WindowProperties.Title
		}
	}

	log.Printf("WARNING: no more title event: %v", windowER.Close())
}

func emitLines(lineCh chan<- []byte) {
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
	// lineScnr := bufio.NewScanner(stdout)
	// // DEBUG ////////////////////////////////////
	lineScnr := bufio.NewScanner(os.Stdin)
	// Set maximum buffer size.
	buf := make([]byte, cnf.BufSize)
	lineScnr.Buffer(buf, 0)
	if err := lineScnr.Err(); err != nil {
		log.Printf("init line scanner: %v", err)
		return
	}

	// Normal op starts after first 4 lines of output from `i3status`.
	// These look like:
	// 		{"version":1}
	// 		[
	// 		[{"name": ... ]
	// 		,[{"name": ... ]
	for range 4 {
		lineScnr.Scan()
		lineCh <- lineScnr.Bytes()
	}

	go func() {
		for lineScnr.Scan() {
			lineCh <- lineScnr.Bytes()
		}

		if err := lineScnr.Err(); err != nil {
			log.Printf("line scanner: %v", err)
		}
	}()
}

func emitMessages(messageCh chan<- []byte) {
	if len(cnf.Pipe) == 0 {
		close(messageCh)
		return
	}
	// Remove any old pipe.
	os.Remove(cnf.Pipe)
	// Create a new FIFO with 0600 permissions.
	if err := syscall.Mkfifo(cnf.Pipe, 0600); err != nil {
		log.Printf("creating pipe: %v", err)
		return
	}

	fi, err := os.OpenFile(cnf.Pipe, os.O_RDWR, 0600)
	if err != nil {
		log.Printf("opening pipe: %v", err)
		return
	}

	pipeScnr := bufio.NewScanner(fi)
	// Set maximum buffer size.
	buf := make([]byte, cnf.MaxWidth*5)
	pipeScnr.Buffer(buf, 0)
	if err := pipeScnr.Err(); err != nil {
		log.Printf("init pipe scanner: %v", err)
		return
	}

	go func() {
		for pipeScnr.Scan() {
			messageCh <- pipeScnr.Bytes()
		}
		fi.Close()

		if err := pipeScnr.Err(); err != nil {
			log.Println("pipe scanner:", err)
		}
	}()
}

func replacer(title string) string {
	// This will be inlined.
	return cnf.Replacer.Replace(title)
}

// lastIndexNonSpace returns the index of last non-space character in s.
func lastIndexNonSpace(s []rune) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] != ' ' {
			return i
		}
	}
	return -1 // All were space.
}

// Read-only `[]byte(string)` convertions are optimized by compiler:
// https://github.com/golang/go/issues/2205 (commits=c8adb30,925d2fb,d63c88d).
// func string2Bytes(s string) []byte {
// 	if len(s) == 0 {
// 		return nil
// 	}
// 	return unsafe.Slice(unsafe.StringData(s), len(s))
// }

// func bytes2String(b []byte) string {
// 	if len(b) == 0 {
// 		return ""
// 	}
// 	return unsafe.String(unsafe.SliceData(b), len(b))
// }
