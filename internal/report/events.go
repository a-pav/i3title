package report

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/a-pav/i3title/internal/bbuf"
	"go.i3wm.org/i3/v4"
)

// emitLines scans [os.Stdin], which is presumed to be data coming from i3status,
// and sends the data to channel.
func emitLines(get func() (b []byte), lineCh chan<- []byte, errCh chan<- error) {
	// Initialize the scanner
	scnr := bufio.NewScanner(os.Stdin)
	scnr.Buffer(make([]byte, minBufSize), maxBufSize)
	if err := scnr.Err(); err != nil {
		errCh <- fmt.Errorf("init line scanner: %v", err)
		return
	}

	// Normal op starts after first 4 lines of output from `i3status`.
	// These look like:
	// 		{"version":1}
	// 		[
	// 		[{"name": ... ]
	// 		,[{"name": ... ]
	for range 4 {
		scnr.Scan()
		buf := get()
		lineCh <- append(buf, scnr.Bytes()...)
	}

	go func() {
		for scnr.Scan() {
			buf := get()
			lineCh <- append(buf, scnr.Bytes()...)
		}

		if err := scnr.Err(); err != nil {
			errCh <- fmt.Errorf("line scanner: %v", err)
			return
		}
	}()
}

// subscribe to title and mode events via i3 Go package.
func subscribe(titleCh, modeCh chan<- []byte, errCh chan<- error) {
	var recvr *i3.EventReceiver
	if modeCh == nil {
		recvr = i3.Subscribe(i3.WindowEventType) // Only window events
	} else {
		recvr = i3.Subscribe(i3.WindowEventType, i3.ModeEventType)
	}

	for recvr.Next() {
		switch ev := recvr.Event().(type) {
		case *i3.WindowEvent:
			switch ev.Change {
			case "title", "focus":
				titleCh <- bbuf.AsBytes(ev.Container.WindowProperties.Title)
			}
		case *i3.ModeEvent:
			modeCh <- bbuf.AsBytes(ev.Change)
		}
	}

	errCh <- fmt.Errorf("no more i3 events: %v", recvr.Close())
}

// emitMessages reads from the named pipe at config.Pipe path and sends the data
// to channel.
func emitMessages(pipe string, get func() (b []byte), messageCh chan<- []byte, errCh chan<- error) {
	// Remove any old pipe.
	os.Remove(pipe)
	// Create a new FIFO with 0600 permissions.
	if err := syscall.Mkfifo(pipe, 0600); err != nil {
		close(messageCh)
		errCh <- fmt.Errorf("creating message pipe: %v", err)
		return
	}

	fi, err := os.OpenFile(pipe, os.O_RDWR, 0600)
	if err != nil {
		close(messageCh)
		errCh <- fmt.Errorf("opening message pipe: %v", err)
		return
	}
	// NO defer fi.Close() here - it would close before goroutine finishes.

	// Set maximum buffer size.
	pipeRd := bufio.NewReaderSize(fi, minBufSize)

	go func() {
		for {
			message, isPrefix, err := pipeRd.ReadLine()
			if err != nil {
				log.Println("message pipe read:", err)
				break
			}

			buf := get()
			messageCh <- append(buf, message...)

			// Discard the remainder of an overlong line.
			for isPrefix {
				_, isPrefix, err = pipeRd.ReadLine()
				if err != nil {
					log.Println("message pipe discard:", err)
					break
				}
			}
		}
		errCh <- fmt.Errorf("message pipe closed: %v", fi.Close())
	}()
}
