// Package bbuf implements a basic buffer.
package bbuf

import (
	"io"
	"unicode"
	"unicode/utf8"
)

const ellipsis = "…"

// BasicBuffer is a fixed-size buffer of bytes that does not grow after initialization.
// Implements [io.Writer] and [io.StringWriter]
type BasicBuffer struct {
	buf []byte
	end int
}

var (
	_ io.Writer       = (*BasicBuffer)(nil)
	_ io.StringWriter = (*BasicBuffer)(nil)
)

// New returns a basic buffer with a fixed size of n.
func New(n int) *BasicBuffer {
	return &BasicBuffer{
		buf: make([]byte, n),
	}
}

func (b *BasicBuffer) Write(p []byte) (int, error) {
	n := copy(b.buf[b.end:], p)
	b.end += n
	return n, nil
}

func (b *BasicBuffer) WriteString(s string) (int, error) {
	n := copy(b.buf[b.end:], s)
	b.end += n
	return n, nil
}

// Available returns how many bytes are unused in the buffer.
func (b *BasicBuffer) Available() int { return len(b.buf) - b.end }

func (b *BasicBuffer) Bytes() []byte { return b.buf[:b.end] }

func (b *BasicBuffer) Reset() { b.end = 0 }

func (b *BasicBuffer) WriteText(p []byte, width int) (int, error) {
	width = max(width, 0) // clamp width at 0
	if width == 0 {
		return b.markTruncated()
	}

	i := 0
	for n := 0; i < len(p); {
		r, size := utf8.DecodeRune(p[i:])

		if repl, ok := safeRune(r); ok {
			if n == width-1 && len(p[i+size:]) > 0 {
				b.markTruncated()
				break
			}
			n++

			if repl != "" {
				// Write replacement string
				if b.Available() < len(repl) {
					break
				}
				b.WriteString(repl)
			} else {
				// Write original bytes
				if b.Available() < size {
					break
				}
				b.Write(p[i : i+size])
			}
		}

		i += size
	}

	return i, nil
}

func (b *BasicBuffer) WriteTextString(s string, width int) (int, error) {
	width = max(width, 0) // clamp width at 0
	if width == 0 {
		return b.markTruncated()
	}

	i := 0
	for n := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])

		if repl, ok := safeRune(r); ok {
			if n == width-1 && len(s[i+size:]) > 0 {
				b.markTruncated()
				break
			}
			n++

			if repl != "" {
				// Write replacement string
				if b.Available() < len(repl) {
					break
				}
				b.WriteString(repl)
			} else {
				// Write original string bytes
				if b.Available() < size {
					break
				}
				b.WriteString(s[i : i+size])
			}
		}

		i += size
	}

	return i, nil
}

func (b *BasicBuffer) markTruncated() (int, error) {
	if b.Available() >= len(ellipsis) {
		return b.WriteString(ellipsis)
	}
	return 0, nil
}

// safeRune evaluates a rune and returns a replacement string and a boolean.
//   - If ok is false, the rune is stripped.
//   - If ok is true and repl is "", the original rune is kept.
//   - If ok is true and repl is not "", the replacement string is used.
//
// To test, craft titles using i3toast or zenity:
//   - i3toast -m "$(printf "'"'==\t=~\!@#$%%^&*()_+-=[]{}|\;:"<>?,./\`')"
//   - zenity --info --title="$(printf "'"'==\t=~\!@#$%%^&*()_+-=[]{}|\;:"<>?,./\`')"
func safeRune(r rune) (repl string, ok bool) {
	switch r {
	case '\t':
		return " ", true
	case '&':
		return "&amp;", true
	case '>':
		return "&gt;", true
	case '<':
		return "&lt;", true
	case '"':
		return "&#34;", true
	case '\\':
		return "&#92;", true
	case
		'\n',
		'\r',
		'\u2029', // PARAGRAPH SEPARATOR (PS)
		'\u2028', // LINE SEPARATOR (LS)
		utf8.RuneError:

		return "", false // Strip
	}

	switch {
	case r >= ' ' && r <= '~':
		// Fast Path: Printable ASCII range (' ' to '~', 0x20 - 0x7E)
		return "", true
	case r <= 0x7F:
		// Fast Path: ASCII Control characters (0x00 - 0x1F and 0x7F 'DEL')
		return "", false // Strip
	case !unicode.IsControl(r):
		// Slow Path: Non-ASCII multi-byte UTF-8 (Chinese, Emojis, accented letters, etc.)
		return "", true
	default:
		return "", false // Strip
	}
}
