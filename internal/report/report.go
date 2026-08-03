package report

import (
	"bytes"
	"os"

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
		report  = bbuf.New(cfg.MaxWidth * 5) // Outgoing report (Big enough buffer, even for all-Chinese characters.)
		mode    = "default"                  // Current i3 mode.
		title   string                       // Current window title.
		message []byte                       // Piped in message.

		line0 []byte                      // Incoming line from `i3status` stdout.
		line1 = make([]byte, cfg.BufSize) // Outgoing line with report in it.

		ARS = []byte{'\036'} // ASCII Record Separator
		LF  = []byte{'\n'}
	)

	doPrint := func() {
		c := 0
		c += copy(line1[c:], line0[:2]) // len(",[") == 2
		c += cfg.Print(line1[c:], report.Bytes())
		c += copy(line1[c:], line0[2:])
		c += copy(line1[c:], LF)

		os.Stdout.Write(line1[:c])
	}
	newReport := func() {
		report.Reset()
		cfg.WriteReport(report, title, mode)

		doPrint()
	}
	newMessage := func() {
		report.Reset()

		const nParts = 5 // number of parts
		if parts := bytes.SplitN(message, ARS, nParts); len(parts) == nParts {
			var (
				formatLeft  = parts[0]
				formatRight = parts[1]
				formatWidth = atoi(parts[2])
				trim        = atoi(parts[3])
				msg         = parts[4]
			)
			report.Write(formatLeft)
			report.WriteText(msg, min(trim, cfg.MaxWidth-formatWidth))
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
		c += copy(line1[c:], LF)

		os.Stdout.Write(line1[:c])

		lineDone <- struct{}{}
	}

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
			if len(message) == 0 { // clearing message?
				newReport()
			} else {
				newMessage()
			}
			messageDone <- struct{}{}
		}
	}
}

func atoi(b []byte) int {
	switch len(b) {
	case 1:
		return int(b[0] - '0')
	case 2:
		return int(b[0]-'0')*10 + int(b[1]-'0')
	case 3:
		return int(b[0]-'0')*100 + int(b[1]-'0')*10 + int(b[2]-'0')
	default:
		return -1
	}
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
