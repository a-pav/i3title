package report

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"syscall"

	"go.i3wm.org/i3/v4"

	"github.com/a-pav/i3title/internal/bbuf"
	"github.com/a-pav/i3title/internal/config"
)

func Run(cfg *config.Config) error {
	var (
		lineCh    = make(chan []byte)
		messageCh = make(chan []byte)
		titleCh   = make(chan string)
		modeCh    = make(chan string)
	)
	go reporter(cfg, lineCh, messageCh, titleCh, modeCh)

	emitLines(cfg.BufSize, lineCh)

	if cfg.Pipe == "" {
		close(messageCh)
	} else {
		emitMessages(cfg.Pipe, cfg.MaxWidth*5, messageCh)
	}

	go emitTitles(titleCh)

	if cfg.ModeStyleIndex < 0 {
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
	// replacer exists to avoid checking `cfg.Replacer != nil` in the main loop.
	replacer := func() func(b *bbuf.BasicBuffer, s string) (int, error) {
		if cfg.Replacer != nil {
			return func(b *bbuf.BasicBuffer, s string) (int, error) {
				return cfg.Replacer.WriteString(b, s)
			}
		}
		return func(b *bbuf.BasicBuffer, s string) (int, error) {
			return b.WriteString(s)
		}
	}()
	trim := trimmer(cfg.MaxWidth)

	doPrint := func() {
		i := cfg.PHIndex
		if i <= 0 {
			i = bytes.Index(line0, []byte(cfg.PH))
			if cfg.PHIndex == 0 { // omited? then cache it.
				// Why i+1? Because the first line does not have a comma at its
				// beginning, but the rest do.
				cfg.PHIndex = i + 1
			} else {
				// forced to recalculate.
			}
		}
		c := 0
		c += copy(line1[c:], line0[:i])
		c += copy(line1[c:], report.Bytes())
		c += copy(line1[c:], line0[i+len(cfg.PH):])
		c += copy(line1[c:], LF)

		os.Stdout.Write(line1[:c])
	}
	newReport := func() {
		if len(message) > 0 {
			return // report doesn't update unless message is cleared.
		}
		report.Reset()
		mw := 0 // mode visible width.
		if i := cfg.ModeStyleIndex; i >= 0 && mode != "default" {
			mw = len([]rune(mode)) + cfg.ModeStyleWidth

			report.WriteString(cfg.ModeStyle[:i])
			report.WriteString(mode)
			report.WriteString(cfg.ModeStyle[i+2:]) // 2 == len("%s")
		}
		replacer(report, trim(title, cfg.MaxWidth-mw))

		doPrint()
	}
	newMessage := func() {
		if len(message) == 0 { // clearing message?
			newReport()
			return
		}
		report.Reset()

		if parts := bytes.SplitN(message, ARS, 6); len(parts) == 6 {
			var (
				formatLeft  = parts[0]
				formatRight = parts[1]
				formatWidth = atoi(parts[2])
				doTrim      = atoi(parts[3])
				doReplacer  = atoi(parts[4])
				msg         = parts[5]
			)
			report.Write(formatLeft)

			if doTrim == 1 && doReplacer == 1 {
				replacer(report, trim(string(msg), cfg.MaxWidth-formatWidth))
			} else if doReplacer == 1 {
				replacer(report, string(msg))
			} else if doTrim == 1 {
				report.WriteString(trim(string(msg), cfg.MaxWidth-formatWidth))
			} else {
				report.Write(msg) // raw write
			}

			report.Write(formatRight)
		} else {
			report.WriteString("<span font='bold' fgcolor='#ff2b2b'>400 Bad Request</span>")
		}

		doPrint()
	}

	// The first two lines don't contain the placeholder and are printed verbatim.
	for range 2 {
		line0 = <-lineCh

		c := 0
		c += copy(line1[c:], line0)
		c += copy(line1[c:], LF)

		os.Stdout.Write(line1[:c])
	}

	for {
		var ok bool
		select {
		case line0 = <-lineCh:
			doPrint() // just print
		case title = <-titleCh:
			newReport()
		case mode, ok = <-modeCh:
			if !ok {
				modeCh = nil // disable
				continue
			}
			newReport()
		case message, ok = <-messageCh:
			if !ok {
				messageCh = nil // disable
				continue
			}
			newMessage()
		}
	}
}

// emitModes subscribes to i3 mode events and sends the modes to channel.
func emitModes(modeCh chan<- string) {
	modeER := i3.Subscribe(i3.ModeEventType)

	for modeER.Next() {
		modeCh <- modeER.Event().(*i3.ModeEvent).Change
	}

	log.Printf("WARNING: no more mode event: %v", modeER.Close())
}

// emitTitles subscribes to i3 window events and sends the titles to channel.
func emitTitles(titleCh chan<- string) {
	windowER := i3.Subscribe(i3.WindowEventType)

	for windowER.Next() {
		e := windowER.Event().(*i3.WindowEvent)
		switch e.Change {
		case "title", "focus":
			titleCh <- e.Container.WindowProperties.Title
		}
	}

	log.Printf("WARNING: no more title event: %v", windowER.Close())
}

// emitLines scans [os.Stdin], which is presumed to be data coming from i3status,
// and sends the data to channel.
func emitLines(bufferSize uint16, lineCh chan<- []byte) {
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
	}

	go func() {
		for lineScnr.Scan() {
			lineCh <- lineScnr.Bytes()
		}

		if err := lineScnr.Err(); err != nil {
			log.Printf("line scanner: %v", err)
		}
	}()
}

// emitMessages reads from the named pipe at config.Pipe path and sends the data
// to channel.
func emitMessages(pipe string, bufferSize int, messageCh chan<- []byte) {
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

	pipeScnr := bufio.NewScanner(fi)
	// Set maximum buffer size.
	buf := make([]byte, bufferSize)
	pipeScnr.Buffer(buf, 0)
	if err := pipeScnr.Err(); err != nil {
		log.Printf("init pipe scanner: %v", err)
		close(messageCh)
		return
	}

	go func() {
		for pipeScnr.Scan() {
			messageCh <- pipeScnr.Bytes()
		}
		fi.Close()

		if err := pipeScnr.Err(); err != nil {
			log.Println("pipe scanner:", err)
		}
	}()
}

// trimmer returns a function that cuts string str at length max.
func trimmer(maxWidth int) func(str string, max int) string {
	// maxWidth as runes, used in rune counting. Held reference to avoid allocation.
	width := make([]rune, maxWidth)

	return func(str string, mxw int) string {
		mxw = max(mxw, 0) // clamp maximum width at 0
		if mxw == 0 {
			return "…"
		}
		if len(str) <= mxw {
			return str
		}
		// Count the runes.
		s := width
		i := 0
		for _, r := range str {
			if i == mxw {
				s = s[:mxw]
				n := lastIndexNonSpace(s) // n <= max-1
				if n == mxw-1 {
					s[n] = '…'
				} else { // n <= max-2
					s[n+1] = '…'
					s = s[:n+2] // n+2 <= max
				}
				return string(s) // alloc
			}
			s[i] = r
			i++
		}

		return str
	}
}

// lastIndexNonSpace returns the index of last non-space character in s.
func lastIndexNonSpace(s []rune) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] != ' ' {
			return i
		}
	}
	return -1 // All were space.
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
