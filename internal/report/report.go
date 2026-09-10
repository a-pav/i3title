package report

import (
	"bytes"
	"os"
	"time"

	"github.com/a-pav/i3title/internal/bbuf"
	"github.com/a-pav/i3title/internal/config"
	"github.com/a-pav/i3title/internal/i3msg"
)

const (
	// Buffer sizes for line and message scanners
	minBufSize = 2 * 1024
	maxBufSize = 3 * 1024
)

var ARS = [1]byte{'\036'} // ASCII Record Separator

func Run(cfg *config.Config) error {
	var (
		lineCh    = make(chan []byte, 1)
		messageCh = make(chan []byte, 1)

		titleCh = make(chan []byte, 1)
		modeCh  = make(chan []byte, 1)
		putI3   func(b []byte)
		getI3   func() (b []byte)

		errCh = make(chan error, 1)
	)
	if cfg.I3Msg {
		getI3, putI3 = createPool(1, cfg.MaxWidth*5)
	} else {
		putI3 = func(b []byte) {} // noop
	}
	// Initialize the memory pool for line and message scanners
	get, put := createPool(2, minBufSize)

	go reporter(cfg,
		lineCh, messageCh, put,
		titleCh, modeCh, putI3,
	)

	emitLines(get, lineCh, errCh)

	if cfg.Pipe == "" {
		close(messageCh)
	} else {
		emitMessages(cfg.Pipe, get, messageCh, errCh)
	}

	if cfg.ModeFormat == "" { // Disabled by config
		close(modeCh)
		modeCh = nil
	}

	if cfg.I3Msg {
		go i3msg.Subscribe(getI3, titleCh, modeCh, errCh)
	} else {
		go subscribe(titleCh, modeCh, errCh)
	}

	return <-errCh
}

// reporter is the central event processor that consumes data from all channels
// and handles the unified reporting logic.
func reporter(cfg *config.Config,
	lineCh, messageCh <-chan []byte, put func(b []byte),
	titleCh, modeCh <-chan []byte, putI3 func(b []byte),
) {
	// This should be a big enough buffer, even for all-Unicode characters plus
	// some more bytes to fit formatings.
	reportSize := cfg.MaxWidth * 5
	var (
		report = bbuf.New(reportSize)                      // Outgoing report
		title  = make([]byte, 0, reportSize)               // Current window title.
		mode   = append(make([]byte, 0, 50), "default"...) // Current i3 mode.

		lineIn = make([]byte, 0, minBufSize) // Incoming line from `i3status` stdout.
		lineOt = make([]byte, 0, minBufSize) // Outgoing line which includes report.

		message = make([]byte, 0, minBufSize) // Piped in message.
		timer   = time.NewTimer(0)            // Timer for message.

		offset = bytes.Repeat([]byte{' '}, cfg.MaxWidth) // Offset spaces.
	)

	doPrint := func() {
		c := 0
		c += copy(lineOt[c:], lineIn[:2]) // 2 == len(",[")
		c += cfg.Print(lineOt[c:], report.Bytes())
		c += copy(lineOt[c:], lineIn[2:])
		c += copy(lineOt[c:], "\n")

		os.Stdout.Write(lineOt[:c])
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
		l := <-lineCh
		lineIn = append(lineIn[:0], l...)
		put(l)

		lineOt = lineOt[:cap(lineOt)] // TODO: drop `copy` and adopt `append` semantics, everywhere
		c := 0
		c += copy(lineOt[c:], lineIn)
		c += copy(lineOt[c:], "\n")

		os.Stdout.Write(lineOt[:c])
	}

	timer.Stop()
	for { // main loop
		select {
		case l := <-lineCh:
			lineIn = append(lineIn[:0], l...)
			put(l)
			doPrint() // just print
		case t := <-titleCh:
			title = append(title[:0], t...)
			putI3(t)
			if len(message) == 0 {
				newReport()
			}
		case m, ok := <-modeCh:
			mode = append(mode[:0], m...)
			putI3(m)
			if !ok {
				modeCh = nil // disable
				continue
			}
			if len(message) == 0 {
				newReport()
			}
		case m, ok := <-messageCh:
			message = append(message[:0], m...)
			put(m)
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
		case <-timer.C:
			message = message[:0] // Erase message
			newReport()
		}
	}
}

// createPool creates a pool of n zero-length []byte buffers, each with the given
// capacity size.
//
// It returns a pair of functions for acquiring and releasing buffers.
func createPool(n, size int) (get func() []byte, put func([]byte)) {
	pool := make(chan []byte, n)
	for range cap(pool) {
		pool <- make([]byte, 0, size)
	}
	get = func() []byte { return <-pool }
	put = func(b []byte) { pool <- b[:0] }

	return
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
