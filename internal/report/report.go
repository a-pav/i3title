package report

import (
	"bytes"
	"os"
	"time"

	"github.com/a-pav/i3title/internal/bbuf"
	"github.com/a-pav/i3title/internal/config"
	"github.com/a-pav/i3title/internal/report/i3ipc"
	"github.com/a-pav/i3title/internal/report/i3msg"
)

const (
	// Buffer sizes for line and message scanners
	minBufSize = 2 * 1024
	maxBufSize = 3 * 1024
)

var ARS = [1]byte{'\036'} // ASCII Record Separator

func Run(cfg *config.Config) error {
	// We use unbuffered channels to benefit from runtime optimizations and
	// direct stack handoffs. This ultimately gives us a much better performance
	// than having fully async operations.
	var (
		lineCh   = make(chan []byte)
		lineDone = make(chan struct{})

		messageCh   = make(chan []byte)
		messageDone = make(chan struct{})

		titleCh = make(chan []byte)
		modeCh  = make(chan []byte)
		i3Done  = make(chan struct{})

		errCh = make(chan error)
	)
	go reporter(cfg,
		lineCh, titleCh, modeCh, messageCh,
		lineDone, i3Done, messageDone,
	)

	emitLines(lineCh, lineDone, errCh)

	if cfg.Pipe == "" {
		close(messageCh)
	} else {
		emitMessages(cfg.Pipe, messageCh, messageDone, errCh)
	}

	if cfg.ModeFormat == "" { // Disabled by config
		close(modeCh)
		modeCh = nil
	}

	if cfg.I3Msg {
		go i3msg.Subscribe(titleCh, modeCh, i3Done, errCh)
	} else {
		go i3ipc.Subscribe(titleCh, modeCh, i3Done, errCh)
	}

	return <-errCh
}

// reporter is the central event processor that consumes data from all channels
// and handles the unified reporting logic.
func reporter(cfg *config.Config,
	lineCh, titleCh, modeCh, messageCh <-chan []byte,
	lineDone, i3Done, messageDone chan<- struct{},
) {
	// Report should be a big enough buffer, even for all-Unicode characters plus
	// some more bytes to fit the formatings.
	var (
		report = bbuf.New(cfg.MaxWidth * 8) // The report

		title = make([]byte, 0, cfg.MaxWidth*4)            // Current window title.
		mode  = append(make([]byte, 0, 128), "default"...) // Current i3 mode.

		line0 = make([]byte, maxBufSize) // Incoming line from `i3status` stdout.
		line1 = make([]byte, maxBufSize) // Outgoing line which includes the report.

		message []byte             // Piped in message.
		timer   = time.NewTimer(0) // Timer for message.

		offset = bytes.Repeat([]byte{' '}, cfg.MaxWidth) // Offset spaces.
	)

	doPrint := func() {
		c := 0
		c += copy(line1[c:], line0[:2])
		c += cfg.Print(line1[c:], report.Bytes())
		c += copy(line1[c:], line0[2:])
		c += copy(line1[c:], "\n")

		os.Stdout.Write(line1[:c])
	}
	newReport := func() {
		cfg.Report(report, title, mode)
		doPrint()
	}
	newMessage := func(msg []byte) {
		report.Reset()

		const nParts = 9 // number of parts
		var parts [nParts][]byte
		if n := splitN(parts[:], msg, ARS[0]); n == nParts {
			var (
				align       = parts[0]
				minWidth    = parts[1]
				offsetWidth = max(0, min(cfg.MaxWidth, atoi(parts[2])))
				formatLeft  = parts[3]
				formatRight = parts[4]
				formatWidth = atoi(parts[5])
				trimWidth   = atoi(parts[6])
				rawMode     = atoi(parts[7])
				text        = parts[8]
			)
			cfg.SetAlign(align)
			cfg.SetMinWidth(minWidth)

			report.Write(offset[:offsetWidth])
			report.Write(formatLeft)
			if rawMode == 1 {
				report.Write(text)
			} else {
				width := min(cfg.MaxWidth-offsetWidth-formatWidth, trimWidth)
				var lines [2][]byte
				if n := splitN(lines[:], text, ARS[0]); n == 2 {
					report.WriteText(lines[0], width)
					report.WriteString("\u2029")
					report.Write(offset[:offsetWidth])
					report.WriteText(lines[1], width)
				} else {
					report.WriteText(text, width)
				}
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
		c := copy(line1[0:], <-lineCh)
		c += copy(line1[c:], "\n")
		os.Stdout.Write(line1[:c])
		lineDone <- struct{}{}
	}

	timer.Stop()
	for { // main loop
		var ok bool
		select {
		case ln0 := <-lineCh:
			line0 = line0[:copy(line0[:cap(line0)], ln0)]
			doPrint() // just print
			lineDone <- struct{}{}
		case tt0 := <-titleCh:
			title = title[:copy(title[:cap(title)], tt0)]
			if len(message) == 0 {
				newReport()
			}
			i3Done <- struct{}{}
		case md0, ok := <-modeCh:
			if !ok {
				modeCh = nil // disable
				continue
			}
			mode = mode[:copy(mode[:cap(mode)], md0)]
			if len(message) == 0 {
				newReport()
			}
			i3Done <- struct{}{}
		case message, ok = <-messageCh:
			if !ok {
				messageCh = nil // disable
				continue
			}
			var parts [2][]byte
			splitN(parts[:], message, ARS[0])
			var (
				timeout = atoi(parts[0])
				msg     = parts[1]
			)
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
