package config

import "github.com/a-pav/i3title/internal/bbuf"

// Config is the Config struct.
type Config struct {
	Format     string    `json:"format"`      // The general format for i3title module.
	ModeFormat string    `json:"mode_format"` // Format for i3 modes.
	MinWidth   rawString `json:"min_width"`   // integer
	MaxWidth   int       `json:"max_width"`   // Maximum width of printed report in characters.
	Align      string    `json:"align"`       // Text alignment.
	Separator  rawString `json:"separator"`   // boolean
	Pipe       string    `json:"pipe"`        // FIFO named pipe for sending messages to overwrite the report.
	BufSize    uint16    `json:"buffer_size"` // Buffer size of both stdin scanner and stdout printer.

	formatIndex     int `json:"-"` // Index of string `%s` inside [Config.Format].
	modeFormatIndex int `json:"-"` // Index of string `%s` inside [Config.ModeFormat].
	modeFormatWidth int `json:"-"` // Width of characters that will be added to report as the result of wrapping raw i3 mode in [Config.ModeFormat].
}

func (c *Config) WriteReport(report *bbuf.Buffer, title, mode string) {
	mw := 0 // mode visible width.
	if c.ModeFormat != "" && mode != "default" {
		mw = len([]rune(mode)) + c.modeFormatWidth

		report.WriteString(c.ModeFormat[:c.modeFormatIndex])
		report.WriteString(mode)
		report.WriteString(c.ModeFormat[c.modeFormatIndex+2:]) // 2 == len("%s")
	}
	report.WriteTextString(title, c.MaxWidth-mw)
}

func (c *Config) Print(line, fullText []byte) int {
	n := 0
	n += copy(line[n:], `{"name":"i3title","markup":"pango","align":"`)
	n += copy(line[n:], c.Align) // "left"
	n += copy(line[n:], `","separator":`)
	n += copy(line[n:], c.Separator) // false
	n += copy(line[n:], `,"min_width":`)
	n += copy(line[n:], c.MinWidth) // 1234
	n += copy(line[n:], `,"full_text":"`)
	n += copy(line[n:], c.Format[:c.formatIndex]) // "<span>
	n += copy(line[n:], fullText)
	n += copy(line[n:], c.Format[c.formatIndex+2:]) // </span>"
	n += copy(line[n:], `"},`)

	return n
}

type rawString string

func (rs *rawString) UnmarshalJSON(data []byte) error {
	*rs = rawString(data)
	return nil
}

func (rs rawString) MarshalJSON() ([]byte, error) {
	return []byte(rs), nil
}

// newConfig return a usable config.
func newConfig() *Config {
	return &Config{
		BufSize:     4096, // more than it's necessary
		MaxWidth:    60,   // less than it's possible
		Align:       "left",
		Separator:   "false",
		MinWidth:    "400",
		Format:      "%s",
		ModeFormat:  "<i>%s</i> | ",
		Pipe:        "/tmp/i3title.pipe",
		formatIndex: 0,
	}
}
