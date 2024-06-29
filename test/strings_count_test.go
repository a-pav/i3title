// Run these benchmarks by `GO111MODULE=off go test -bench=.` inside test directory.

package i3title_test

import (
	"bytes"
	"strings"
	"testing"
)

var s = "some XXX string XXX with XXX repeating substring XXX"

func BenchmarkStringsCount(b *testing.B) {
	// s := "some XXX string XXX with XXX repeating substring XXX"

	b.ResetTimer()
	b.ReportAllocs()
	// n := 0
	for i := 0; i < b.N; i++ {
		for i := 0; i < len(oldnew); i += 2 {
			n := strings.Count(s, "XXX")
			b.Logf("n: %v", n) // DEBUG
			_ = n
		}
	}
}

func BenchmarkBytesCount(b *testing.B) {
	// s := "some XXX string XXX with XXX repeating substring XXX"

	b.ResetTimer()
	b.ReportAllocs()
	// n := 0
	for i := 0; i < b.N; i++ {
		for i := 0; i < len(oldnew); i += 2 {
			n := bytes.Count([]byte(s), []byte("XXX"))
			b.Logf("n: %v", n) // DEBUG
			_ = n
		}
	}
}
