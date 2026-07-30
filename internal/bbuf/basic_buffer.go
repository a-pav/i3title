// Package bbuf implements a basic buffer.
package bbuf

import "io"

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

func (b *BasicBuffer) Bytes() []byte { return b.buf[:b.end] }

func (b *BasicBuffer) Reset() { b.end = 0 }
