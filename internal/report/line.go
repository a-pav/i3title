package report

import (
	"bufio"
	"fmt"
	"os"
)

// emitLines scans [os.Stdin], which is presumed to be data coming from i3status,
// and sends the data to channel.
func emitLines(lineCh chan<- []byte, lineDone <-chan struct{}, errCh chan<- error) {
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
		lineCh <- scnr.Bytes()
		<-lineDone
	}

	go func() {
		for scnr.Scan() {
			lineCh <- scnr.Bytes()
			<-lineDone
		}

		if err := scnr.Err(); err != nil {
			errCh <- fmt.Errorf("line scanner: %v", err)
			return
		}
	}()
}
