// Package bbuf implements a basic buffer.
package bbuf

import "io"

// basicBuffer is a fixed-size buffer of bytes that does not grow after initialization.
// Implements [io.Writer] and [io.StringWriter]
type basicBuffer struct {
	buf []byte
	end int
}

var (
	_ io.Writer       = (*basicBuffer)(nil)
	_ io.StringWriter = (*basicBuffer)(nil)
)

// New returns a basic buffer with a fixed size of n.
func New(n int) *basicBuffer {
	return &basicBuffer{
		buf: make([]byte, n),
	}
}

func (b *basicBuffer) Write(p []byte) (int, error) {
	n := copy(b.buf[b.end:], p)
	b.end += n
	return n, nil
}

func (b *basicBuffer) WriteString(s string) (int, error) {
	n := copy(b.buf[b.end:], s)
	b.end += n
	return n, nil
}

func (b *basicBuffer) Bytes() []byte { return b.buf[:b.end] }

func (b *basicBuffer) Reset() { b.end = 0 }
