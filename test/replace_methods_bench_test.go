// Run these benchmarks by `GO111MODULE=off go test -bench=.` inside test directory.

package i3title_test

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

type MatchReplace struct {
	M *regexp.Regexp
	R string
}

var oldnew = []string{
	"&",
	"&amp;",

	">",
	"&gt;",

	"<",
	"&lt;",

	"\"",
	"&#34;", // "&#34;" is shorter than "&quot;".

	"\\\\",
	"&#92;", // "&#92;" is shorter than "&Backslash;", or anything else.

	" — Firefox Developer Edition",
	"",
	" — Mozilla Firefox",
	"",
}

func getBaseString() string {
	return "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx0987654&321"
	// return "~!@#$%^&*()_+-=[]{}|;:'\"<>?,./` — Firefox Developer Edition\\"
	// return "( playground ) bench/re-vs-replacer/re_vs_replacer_test.go"
	// return "\\"
}

func getBaseBytes() []byte {
	return []byte("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx1234567&890")
	// return []byte("~!@#$%^&*()_+-=[]{}|;:'\"<>?,./` — Firefox Developer Edition\\")
	// return []byte("( playground ) bench/re-vs-replacer/re_vs_replacer_test.go")
	// return []byte("\\")
}

func BenchmarkRegexReplaceAllString(b *testing.B) {

	oldnewRECompiled := make([]MatchReplace, 0)
	for i := 0; i < len(oldnew); i += 2 {
		re, err := regexp.Compile(oldnew[i])
		if err != nil {
			println("error compiling regex:", err.Error())
			continue
		}

		oldnewRECompiled = append(oldnewRECompiled, MatchReplace{
			M: re,
			R: oldnew[i+1],
		})
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		title := getBaseString()
		for _, f := range oldnewRECompiled {
			title = f.M.ReplaceAllString(title, f.R)
		}

		// b.Logf("end title: %s", title) // DEBUG
	}
}

func BenchmarkBytesReplace(b *testing.B) {

	oldnewBytes := make([][]byte, len(oldnew))
	for i, v := range oldnew {
		oldnewBytes[i] = []byte(v)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		title := getBaseBytes()
		for i := 0; i < len(oldnewBytes); i += 2 {
			title = bytes.Replace(title, oldnewBytes[i], oldnewBytes[i+1], -1)
		}

		// b.Logf("end title: %s", title) // DEBUG
	}
}

func BenchmarkStringsReplacerReplace(b *testing.B) {

	replacer := strings.NewReplacer(oldnew...)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		title := getBaseString()
		title = replacer.Replace(title)

		// b.Logf("end title: %s", title) // DEBUG
	}
}

func BenchmarkStringsReplace(b *testing.B) {

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		title := getBaseString()
		for i := 0; i < len(oldnew); i += 2 {
			title = strings.Replace(title, oldnew[i], oldnew[i+1], -1)
		}

		// b.Logf("end title: %s", title) // DEBUG
	}
}
