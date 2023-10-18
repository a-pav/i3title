package i3title_test

import (
	"testing"
)

func BenchmarkSliceIndexingI(b *testing.B) {
	const (
		maxlen    = 95
		maxesclen = 5
	)
	title := getBaseString()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {

		if len([]rune(title)) > maxlen {
			rs := []rune(title)                          // alloc.
			cut := lastSafeIndexI(rs, maxlen, maxesclen) // drop half-fromed escape sequence (no alloc.)
			cut = lastNonspaceIndexI(rs, cut) + 1        // drop trailing spaces (no alloc.)
			rs = rs[:cut]                                // shrink (no alloc.)
			rs = append(rs, '…')                         // append shrinkage indicator (no alloc.)
			_ = string(rs)                               // alloc.

			// b.Logf("%q\n", string(rs)) // DEBUG
		}
	}

}

func BenchmarkSliceIndexingII(b *testing.B) {
	const (
		maxlen    = 95
		maxesclen = 5
	)
	title := getBaseString()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if len([]rune(title)) > maxlen {
			rs := []rune(title)                      // alloc.
			rs = rs[:maxlen]                         // shrink (no alloc.)
			rs = rs[:lastSafeIndexII(rs, maxesclen)] // drop half-fromed escape sequence (no alloc.)
			rs = rs[:lastNonspaceIndexII(rs)+1]      // drop trailing spaces (no alloc.)
			rs = append(rs, '…')                     // append shrinkage indicator (no alloc.)
			_ = string(rs)                           // alloc.

			// b.Logf("%q\n", string(rs)) // DEBUG

		}

	}

}

func lastSafeIndexI(rs []rune, cut, width int) int {
	var (
		end   = cut
		start = cut - width

		iSemicolon = -1
		iAmpersand = -1
	)

	for i := end - 1; i >= start; i-- {
		switch {
		case iAmpersand < 0 && rs[i] == '&':
			iAmpersand = i
		case iSemicolon < 0 && rs[i] == ';':
			iSemicolon = i
		}
	}

	switch {
	case iAmpersand < iSemicolon:
		// There's a fully formed escape sequence inside width before end. e.g. `&escape;`
		return end
	case iAmpersand > 0:
		// There's a half-formed escape sequence inside width before end. e.g. `&esca`
		return iAmpersand
	}

	// There was no escape sequence inside the width.
	return end
}

func lastNonspaceIndexI(rs []rune, cut int) int {
	for i := cut - 1; i > 0; i-- {
		if rs[i] != ' ' {
			return i
		}
	}

	// All space!
	return 0
}

func lastSafeIndexII(rs []rune, width int) int {
	var (
		len0  = len(rs)
		start = len0 - width
		end   = len0 - 1

		iSemicolon = -1
		iAmpersand = -1
	)

	for i := end; i >= start; i-- {
		switch {
		case iAmpersand < 0 && rs[i] == '&':
			iAmpersand = i
		case iSemicolon < 0 && rs[i] == ';':
			iSemicolon = i
		}
	}

	switch {
	case iAmpersand < iSemicolon:
		// There's a fully formed escape sequence inside width before end. e.g. `&escape;`
		return len0
	case iAmpersand > 0:
		// There's a half-formed escape sequence inside width before end. e.g. `&esca`
		return iAmpersand
	}

	// There was no escape sequence inside the width.
	return len0
}

func lastNonspaceIndexII(rs []rune) int {
	for i := len(rs) - 1; i > 0; i-- {
		if rs[i] != ' ' {
			return i
		}
	}

	// All space!
	return 0
}
