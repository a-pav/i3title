package i3title_test

import (
	"testing"
)

// Slicing one time.
func BenchmarkSliceIndexingI(b *testing.B) {
	bs := makeSlice()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		x := find(bs, 'A', 3)
		x = find(bs, 'D', 5)
		x = find(bs, 'Z', 6)
		x = find(bs, 'M', 12)
		x = find(bs, 'D', 14)

		_ = string(bs[:x])
		// b.Logf("%q\n", string(bs[:x])) // DEBUG
	}
}

// Slicing multiple times.
func BenchmarkSliceIndexingII(b *testing.B) {
	bs := makeSlice()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bsii := bs[:find(bs, 'A', 3)]
		bsii = bsii[:find(bs, 'D', 5)]
		bsii = bsii[:find(bs, 'Z', 6)]
		bsii = bsii[:find(bs, 'M', 12)]
		bsii = bsii[:find(bs, 'D', 14)]

		_ = string(bsii)
		// b.Logf("%q\n", string(bsii)) // DEBUG
	}
}

func makeSlice() []byte {
	n := 20
	b := []byte{'A', 'B', 'C', 'D', 'X', 'Z', 'Y', 'M', 'N', 'O'}
	bs := make([]byte, 0, n*len(b))
	for i := 0; i < n; i++ {
		bs = append(bs, b...)
	}

	return bs
}

func find(in []byte, f byte, count int) int {
	for i := len(in) - 1; i > 0; i-- {
		if in[i] == f {
			count--
			if count == 0 {
				return i
			}
		}
	}

	return 0
}
