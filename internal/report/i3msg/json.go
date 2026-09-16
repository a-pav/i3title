package i3msg

import (
	"unicode/utf16"
	"unicode/utf8"
)

// nextString extracts a JSON string and unescapes it in-place.
// It assumes b starts immediately AFTER the opening quote (`"`).
// It returns:
//   - s: the cleanly parsed string (reusing the memory of b).
//   - n: total bytes read from b, including the closing quote.
//   - ok: false if the string is malformed or missing a closing quote.
//
// It's correctness matches that of [json.unquoteBytes] from
// stdlib:encoding/json/decode.go. But that function does 2 allocations.
func nextString(b []byte) (s []byte, n int, ok bool) {
	// FAST PATH: Scan for either an escape '\' or the closing '"'
	for i := 0; i < len(b); i++ {
		c := b[i]
		if c == '"' {
			// No escapes found! Just return the slice directly.
			return b[:i], i + 1, true
		}
		if c == '\\' {
			// Slow path: An escape was found. Unescape the rest in-place.
			return unescapeInPlace(b, i)
		}
	}
	return nil, 0, false // Missing closing quote
}

// unescapeInPlace modifies the byte slice starting at `firstSlash`
func unescapeInPlace(b []byte, firstSlash int) ([]byte, int, bool) {
	r := firstSlash
	w := firstSlash

	for r < len(b) {
		c := b[r]

		if c != '\\' {
			if c == '"' {
				// We found the end quote. Return the unescaped slice
				return b[:w], r + 1, true
			}
			// Normal character: copy and advance both pointers
			b[w] = c
			w++
			r++
			continue
		}

		// We hit a backslash '\'. Advance the read pointer.
		r++
		if r >= len(b) {
			return nil, 0, false // Incomplete escape
		}

		// Handle the character after the backslash
		switch b[r] {
		case '"', '\\', '/':
			b[w] = b[r]
			w++
			r++
		case 'n':
			b[w] = '\n'
			w++
			r++
		case 'r':
			b[w] = '\r'
			w++
			r++
		case 't':
			b[w] = '\t'
			w++
			r++
		case 'b':
			b[w] = '\b'
			w++
			r++
		case 'f':
			b[w] = '\f'
			w++
			r++
		case 'u':
			// Handle Unicode escapes (\uXXXX)
			r++
			if r+4 > len(b) {
				return nil, 0, false
			}
			un := decodeHex4(b[r : r+4])
			if un < 0 {
				return nil, 0, false
			}
			r += 4
			// Handle UTF-16 surrogate pairs
			if utf16.IsSurrogate(un) {
				if r+6 <= len(b) && b[r] == '\\' && b[r+1] == 'u' {
					un2 := decodeHex4(b[r+2 : r+6])
					if un2 >= 0 {
						un = utf16.DecodeRune(un, un2)
						r += 6
					}
				}
			}
			// Encode the rune natively back into the byte slice
			w += utf8.EncodeRune(b[w:], un)
		default:
			return nil, 0, false // Invalid escape sequence
		}
	}

	return nil, 0, false // Missing closing quote
}

// decodeHex4 parses exactly 4 hex characters into a rune
func decodeHex4(b []byte) rune {
	var r rune
	for i := 0; i < 4; i++ {
		r <<= 4
		c := b[i]
		switch {
		case '0' <= c && c <= '9':
			r += rune(c - '0')
		case 'a' <= c && c <= 'f':
			r += rune(c - 'a' + 10)
		case 'A' <= c && c <= 'F':
			r += rune(c - 'A' + 10)
		default:
			return -1
		}
	}
	return r
}
