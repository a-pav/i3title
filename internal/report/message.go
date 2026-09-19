package report

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"syscall"
)

// emitMessages reads from the named pipe and sends the data to channel.
func emitMessages(pipe string, messageCh chan<- []byte, messageDone <-chan struct{}, errCh chan<- error) {
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

			messageCh <- message
			<-messageDone

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
