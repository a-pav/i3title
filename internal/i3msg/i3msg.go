package i3msg

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os/exec"
)

var (
	// Magic keys that we index inside the JSON payload received from i3-msg

	changeKey = []byte(`{"change":"`)
	titleKey  = []byte(`"title":"`)
)

func Subscribe(titelCh, modeCh chan<- []byte, get func() (b []byte)) error {
	cmd := exec.Command("i3-msg",
		// The event type `"i3title"` is ignored by i3-msg, but appears in the process name, serving as a visual hint.
		"-t", "subscribe", `[ "i3title", "window", "mode" ]`, "-m",
	)
	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("i3msg: pipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("i3msg: start: %v", err)
	}

	// Prepare the scanner
	scnr := bufio.NewScanner(pipe)
	// minBuf is large enough that it is unlikely that scanner will ever reallocate
	minBuf, maxBuf := 2*1024, 3*1024 // min=2KB, max=3KB
	scnr.Buffer(make([]byte, minBuf), maxBuf)

	go subscribe(scnr, titelCh, modeCh, get)

	return cmd.Wait()
}

func subscribe(scnr *bufio.Scanner, titelCh, modeCh chan<- []byte, get func() (b []byte)) {
	for scnr.Scan() {
		var (
			data   = scnr.Bytes()
			change = changeValue(data)
		)
		if len(change) == 0 {
			continue // invalid
		}

		switch {
		case mode(data):
			modeCh <- append(get(), change...)
		case title(change):
			tv := titleValue(data)
			titelCh <- append(get(), tv...)
		}
	}

	if err := scnr.Err(); err != nil {
		log.Printf("i3msg scanner: %v", err)
	}
}

// mode reports whether data is a mode event payload.
func mode(data []byte) bool {
	// data is expected to not include a new line.
	// mode event payloads end with `true}` or `false}`.
	if data[len(data)-2] == 'e' {
		return true
	}
	return false
}

// title reports whether change indicates a title change.
func title(change []byte) bool {
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
