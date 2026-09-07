package config

import (
	"encoding/json"
	"unicode/utf8"

	"github.com/a-pav/i3title/internal/bbuf"
)

// Config is the Config struct.
type Config struct {
	Format     json.RawMessage `json:"format"`      // The general format for i3title module.
	ModeFormat string          `json:"mode_format"` // Format for i3 modes.
	MinWidth   json.RawMessage `json:"min_width"`   // integer
	MaxWidth   int             `json:"max_width"`   // Maximum width of printed report in characters.
	Align      json.RawMessage `json:"align"`       // Text alignment.
	Separator  json.RawMessage `json:"separator"`   // boolean
	Pipe       string          `json:"pipe"`        // FIFO named pipe for sending messages to overwrite the report.
	BufSize    uint16          `json:"buffer_size"` // Buffer size of both stdin scanner and stdout printer.

	minWidth        []byte `json:"-"` // Clone of [Config.MinWidth] that can be adjusted by incoming message.
	align           []byte `json:"-"` // Clone of [Config.Align] that can be adjusted by incoming message.
	formatIndex     int    `json:"-"` // Index of string `%s` inside [Config.Format].
	modeFormatIndex int    `json:"-"` // Index of string `%s` inside [Config.ModeFormat].
	modeFormatWidth int    `json:"-"` // Width of characters that will be added to report as the result of wrapping raw i3 mode in [Config.ModeFormat].
}

// Report prepares a new report consisting of title and mode.
func (c *Config) Report(report *bbuf.Buffer, title, mode []byte) {
	c.reset()
	report.Reset()

	mw := 0 // mode visible width.
	if string(mode) != "default" && c.ModeFormat != "" {
		mw = utf8.RuneCount(mode) + c.modeFormatWidth

		report.WriteString(c.ModeFormat[:c.modeFormatIndex])
		report.Write(mode)
		report.WriteString(c.ModeFormat[c.modeFormatIndex+2:]) // 2 == len("%s")
	}
	report.WriteText(title, c.MaxWidth-mw)
}

// Print writes the complete i3title's JSON object on the line.
func (c *Config) Print(line, fullText []byte) int {
	n := 0
	n += copy(line[n:], `{"name":"i3title","markup":"pango","align":`)
	n += copy(line[n:], c.align) // "left"
	n += copy(line[n:], `,"separator":`)
	n += copy(line[n:], c.Separator) // false
	n += copy(line[n:], `,"min_width":`)
	n += copy(line[n:], c.minWidth) // 1234
	n += copy(line[n:], `,"full_text":`)
	n += copy(line[n:], c.Format[:c.formatIndex]) // "<span>
	n += copy(line[n:], fullText)
	n += copy(line[n:], c.Format[c.formatIndex+2:]) // </span>"
	n += copy(line[n:], `},`)

	return n
}

func (c *Config) SetMinWidth(m []byte) {
	c.minWidth = append(c.minWidth[:0], m...)
}

func (c *Config) SetAlign(a []byte) {
	switch a[0] {
	case '0':
		c.align = append(c.align[:0], `"left"`...)
	case '1':
		c.align = append(c.align[:0], `"center"`...)
	case '2':
		c.align = append(c.align[:0], `"right"`...)
	}
}

// getAlign is the inverse of [Config.SetAlign]
func (c *Config) getAlign() byte {
	switch string(c.Align) {
	case `"left"`:
		return '0'
	case `"center"`:
		return '1'
	case `"right"`:
		return '2'
	default:
		return '0' // left
	}
}

// reset resets dynamic fields back to what was read from the config file.
func (c *Config) reset() {
	// None of these should allocate if c was created via [newConfig]
	c.minWidth = append(c.minWidth[:0], c.MinWidth...)
	c.align = append(c.align[:0], c.Align...)
}

// newConfig return a usable config.
func newConfig() *Config {
	c := Config{
		BufSize:  4096, // more than it's necessary
		MaxWidth: 60,   // less than it's possible
		Pipe:     "/tmp/i3title.pipe",

		Format:      []byte(`"%s"`),
		ModeFormat:  "<i>%s</i> | ",
		formatIndex: 0,
		// Pre-allocate and initialize with default values
		Align:     append(make([]byte, 0, 8), `"left"`...), // 8 == len(`"center"`)
		Separator: append(make([]byte, 0, 5), `false`...),  // 5 == len(`false`)
		MinWidth:  append(make([]byte, 0, 4), `100`...),    // 4 == len(`1920`)
	}
	c.align = make([]byte, 0, len(c.Align))
	c.minWidth = make([]byte, 0, len(c.MinWidth))
	c.reset()

	return &c
}
