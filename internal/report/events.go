package report

import (
	"bufio"
	"log"
	"os"
	"syscall"

	"github.com/a-pav/i3title/internal/bbuf"
	"go.i3wm.org/i3/v4"
)

// emitLines scans [os.Stdin], which is presumed to be data coming from i3status,
// and sends the data to channel.
func emitLines(bufferSize uint16, lineCh chan<- []byte, lineDone <-chan struct{}) {
	lineScnr := bufio.NewScanner(os.Stdin)
	// Set maximum buffer size.
	buf := make([]byte, bufferSize)
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
		<-lineDone
	}

	go func() {
		for lineScnr.Scan() {
			lineCh <- lineScnr.Bytes()
			<-lineDone
		}

		if err := lineScnr.Err(); err != nil {
			log.Printf("line scanner: %v", err)
		}
	}()
}

// emitModes subscribes to i3 mode events and sends the modes to channel.
func emitModes(modeCh chan<- []byte) {
	modeER := i3.Subscribe(i3.ModeEventType)

	for modeER.Next() {
		modeCh <- bbuf.AsBytes(modeER.Event().(*i3.ModeEvent).Change)
	}

	log.Printf("WARNING: no more mode event: %v", modeER.Close())
}

// emitTitles subscribes to i3 window events and sends the titles to channel.
func emitTitles(titleCh chan<- []byte) {
	windowER := i3.Subscribe(i3.WindowEventType)

	for windowER.Next() {
		e := windowER.Event().(*i3.WindowEvent)
		switch e.Change {
		case "title", "focus":
			titleCh <- bbuf.AsBytes(e.Container.WindowProperties.Title)
		}
	}

	log.Printf("WARNING: no more title event: %v", windowER.Close())
}

// emitMessages reads from the named pipe at config.Pipe path and sends the data
// to channel.
func emitMessages(pipe string, bufferSize int, messageCh chan<- []byte, messageDone <-chan struct{}) {
	// Remove any old pipe.
	os.Remove(pipe)
	// Create a new FIFO with 0600 permissions.
	if err := syscall.Mkfifo(pipe, 0600); err != nil {
		log.Printf("creating pipe: %v", err)
		close(messageCh)
		return
	}

	fi, err := os.OpenFile(pipe, os.O_RDWR, 0600)
	if err != nil {
		log.Printf("opening pipe: %v", err)
		close(messageCh)
		return
	}
	// NO defer fi.Close() here - it would close before goroutine finishes.

	// Set maximum buffer size.
	pipeRd := bufio.NewReaderSize(fi, bufferSize)

	go func() {
		for {
			message, isPrefix, err := pipeRd.ReadLine()
			if err != nil {
				log.Println("pipe reader:", err)
				break
			}

			messageCh <- message
			<-messageDone

			// Discard the remainder of an overlong line.
			for isPrefix {
				_, isPrefix, err = pipeRd.ReadLine()
				if err != nil {
					log.Println("pipe reader: discard:", err)
					break
				}
			}
		}
		fi.Close()
	}()
}
