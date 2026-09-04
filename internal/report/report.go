package report

import (
	"bytes"
	"os"
	"time"

	"github.com/a-pav/i3title/internal/bbuf"
	"github.com/a-pav/i3title/internal/config"
)

func Run(cfg *config.Config) error {
	var (
		lineCh      = make(chan []byte)
		lineDone    = make(chan struct{})
		messageCh   = make(chan []byte)
		messageDone = make(chan struct{})
		titleCh     = make(chan string)
		modeCh      = make(chan string)
	)
	go reporter(cfg,
		lineCh, messageCh,
		lineDone, messageDone,
		titleCh, modeCh,
	)

	emitLines(cfg.BufSize, lineCh, lineDone)

	if cfg.Pipe == "" {
		close(messageCh)
	} else {
		emitMessages(cfg.Pipe, cfg.MaxWidth*5, messageCh, messageDone)
	}

	go emitTitles(titleCh)

	if cfg.ModeFormat == "" {
		close(modeCh)
	} else {
		go emitModes(modeCh)
	}

	return nil
}

// reporter is the central event processor that consumes data from all channels
// and handles the unified reporting logic.
func reporter(cfg *config.Config,
	lineCh, messageCh <-chan []byte,
	lineDone, messageDone chan<- struct{},
	titleCh, modeCh <-chan string,
) {
	var (
		report  = bbuf.New(cfg.MaxWidth * 5) // Outgoing report (Big enough buffer, even for all-Unicode characters.)
		mode    = "default"                  // Current i3 mode.
		title   string                       // Current window title.
		message []byte                       // Piped in message.
		timer   = time.NewTimer(0)           // Timer for message.

		line0 []byte                      // Incoming line from `i3status` stdout.
		line1 = make([]byte, cfg.BufSize) // Outgoing line with report in it.

		offset = bytes.Repeat([]byte{' '}, cfg.MaxWidth) // Offset spaces.
		ARS    = []byte{'\036'}                          // ASCII Record Separator
	)

	doPrint := func() {
		c := 0
		c += copy(line1[c:], line0[:2]) // 2 == len(",[")
		c += cfg.Print(line1[c:], report.Bytes())
		c += copy(line1[c:], line0[2:])
		c += copy(line1[c:], "\n")

		os.Stdout.Write(line1[:c])
	}
	newReport := func() {
		report.Reset()
		cfg.WriteReport(report, title, mode)

		doPrint()
	}
	newMessage := func(msg []byte) {
		report.Reset()

		const nParts = 7 // number of parts
		var parts [nParts][]byte
		if n := splitN(parts[:], msg, ARS[0]); n == nParts {
			var (
				offsetWidth = max(0, min(cfg.MaxWidth, atoi(parts[0])))
				formatLeft  = parts[1]
				formatRight = parts[2]
				formatWidth = atoi(parts[3])
				trimWidth   = atoi(parts[4])
				rawMode     = atoi(parts[5])
				text        = parts[6]
			)
			report.Write(offset[:offsetWidth])
			report.Write(formatLeft)
			if rawMode == 1 {
				report.Write(text)
			} else {
				formatWidth += offsetWidth
				report.WriteText(text, min(cfg.MaxWidth-formatWidth, trimWidth))
			}
			report.Write(formatRight)
		} else {
			report.WriteString("<span font='bold' fgcolor='#ff2b2b'>400 Bad Request</span>")
		}

		doPrint()
	}

	// The first two lines are i3bar protocol handshake and the third line is the
	// odd one without a comma `,` at its front, so these are printed verbatim.
	for range 3 {
		line0 = <-lineCh

		c := 0
		c += copy(line1[c:], line0)
		c += copy(line1[c:], "\n")

		os.Stdout.Write(line1[:c])

		lineDone <- struct{}{}
	}

	timer.Stop()
	for { // main loop
		var ok bool
		select {
		case line0 = <-lineCh:
			doPrint() // just print
			lineDone <- struct{}{}
		case title = <-titleCh:
			if len(message) == 0 {
				newReport()
			}
		case mode, ok = <-modeCh:
			if !ok {
				modeCh = nil // disable
				continue
			}
			if len(message) == 0 {
				newReport()
			}
		case message, ok = <-messageCh:
			if !ok {
				messageCh = nil // disable
				continue
			}
			var parts [2][]byte
			splitN(parts[:], message, ARS[0])
			timeout, msg := atoi(parts[0]), parts[1]
			switch {
			case timeout > 0: // Common path
				newMessage(msg)
				timer.Reset(time.Duration(timeout) * time.Second)
			case timeout == -1:
				newMessage(msg)
				timer.Stop()
			case timeout == 0:
				message = message[:0] // Erase message
				timer.Stop()
				newReport()
			}
			messageDone <- struct{}{}
		case <-timer.C:
			message = message[:0] // Erase message
			newReport()
		}
	}
}

// splitN is an allocation-free version of [bytes.SplitN] that writes into parts.
// N is implied by the len(parts). The last part will be the unsplit remainder.
func splitN(parts [][]byte, s []byte, sep byte) int {
	if len(parts) == 0 {
		return 0
	}

	p := 0
	for p < len(parts)-1 {
		i := bytes.IndexByte(s, sep)
		if i < 0 {
			break // No more separators, dump the remainder
		}

		parts[p] = s[:i:i]
		s = s[i+1:]
		p++
	}
	// Last part holds the remainder of 's'
	parts[p] = s

	return p + 1
}

func atoi(b []byte) int {
	interror := -128

	if len(b) == 0 {
		return interror
	}

	neg := b[0] == '-'
	if neg {
		b = b[1:]
		if len(b) == 0 {
			return interror // Handle a bare "-"
		}
	}

	var n int
	switch len(b) {
	case 1:
		n = int(b[0] - '0')
	case 2:
		n = int(b[0]-'0')*10 + int(b[1]-'0')
	case 3:
		n = int(b[0]-'0')*100 + int(b[1]-'0')*10 + int(b[2]-'0')
	default: // Fallback for > 3 digits
		for _, ch := range b {
			n = n*10 + int(ch-'0')
		}
	}

	if neg {
		return -n
	}
	return n
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
