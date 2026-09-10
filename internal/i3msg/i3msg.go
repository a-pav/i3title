package i3msg

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
)

const (
	// Buffer sizes for `i3-msg` scanner
	minBufSize = 2 * 1024 // large enough that it's unlikely for scanner to reallocate
	maxBufSize = 3 * 1024
)

var (
	// Magic keys that we index inside the JSON payload received from i3-msg

	changeKey = []byte(`{"change":"`)
	titleKey  = []byte(`"title":"`)
)

func Subscribe(get func() (b []byte), titelCh, modeCh chan<- []byte, errCh chan<- error) {
	cmd := exec.Command("i3-msg",
		// The event type `"i3title"` is ignored by i3-msg, but appears in the process name, serving as a visual hint.
		"-t", "subscribe", `[ "i3title", "window", "mode" ]`, "-m",
	)
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		errCh <- fmt.Errorf("i3msg: pipe: %v", err)
		return
	}
	if err := cmd.Start(); err != nil {
		errCh <- fmt.Errorf("i3msg: start: %v", err)
		return
	}
	// Initialize the scanner
	scnr := bufio.NewScanner(pipe)
	scnr.Buffer(make([]byte, minBufSize), maxBufSize)

	go subscribe(scnr, get, titelCh, modeCh, errCh)

	errCh <- fmt.Errorf("i3-msg: %v", cmd.Wait())
}

func subscribe(scnr *bufio.Scanner, get func() (b []byte), titelCh, modeCh chan<- []byte, errCh chan<- error) {
	for scnr.Scan() {
		var (
			data   = scnr.Bytes()
			change = changeValue(data)
		)
		if len(change) == 0 {
			continue // invalid
		}

		switch {
		case modeEvent(data):
			mode := change
			buf := get()
			modeCh <- append(buf, mode...)
		case titleEvent(change):
			title := titleValue(data)
			buf := get()
			titelCh <- append(buf, title...)
		}
	}

	if err := scnr.Err(); err != nil {
		errCh <- fmt.Errorf("i3msg scanner: %v", err)
	}
}

// modeEvent reports whether data is a mode event payload.
func modeEvent(data []byte) bool {
	// data is expected to not include a new line.
	// mode event payloads end with `true}` or `false}`.
	if data[len(data)-2] == 'e' {
		return true
	}
	return false
}

// titleEvent reports whether change indicates a title event.
func titleEvent(change []byte) bool {
	switch string(change) {
	case "title", "focus":
		return true
	default:
		return false
	}
}

// changeValue extracts the value of "change" key inside the JSON payload.
//
// When the payload corresponds to a mode event, the value contains the mode name.
//
// When the payload corresponds to a window event, the value indicates the type.
func changeValue(b []byte) []byte {
	i := bytes.Index(b, changeKey)
	if i < 0 {
		return nil
	}
	i += len(changeKey)

	change, _, ok := nextString(b[i:])
	if ok {
		return change
	}
	return nil
}

// titleValue returns the window title
func titleValue(b []byte) []byte {
	// [changeValue] has done all the necessary bound checkings already
	i := bytes.Index(b, titleKey) + len(titleKey)
	title, _, ok := nextString(b[i:])
	if ok {
		return title
	}
	return nil
}
