// Library for inserting ANSI escapes into prompt strings.
package sbpgo

import (
	"bytes"
	"fmt"
	"strings"
	"unicode/utf8"
)

// A string of text, with some formatting markers.
type StyledString []StyledRune

type StyledRune struct {
	Style Style
	Text  rune
}

// TODO: support more colors and modifiers with fish
// http://go/wiki/ANSI_escape_code#24-bit
// http://go/wiki/ANSI_escape_code#SGR_parameters

// 24-bit color.
type Color struct {
	R byte
	G byte
	B byte
}

func Rgb(r, g, b byte) *Color {
	return &Color{r, g, b}
}

func (self Color) Join(sep string) string {
	return fmt.Sprintf("%d%s%d%s%d", self.R, sep, self.G, sep, self.B)
}

// Some standard colors.
// TODO: audit use of these. Adopt a cooler scheme (blue/black/white/yellow?)
var Black = Rgb(0, 0, 0)
var Red = Rgb(255, 0, 0)
var Green = Rgb(0, 255, 0)
var Blue = Rgb(0, 0, 255)
var Cyan = Rgb(0, 255, 255)
var Magenta = Rgb(255, 0, 255)
var Yellow = Rgb(255, 255, 0)
var White = Rgb(255, 255, 255)

type Style struct {
	// Nil means use the default.
	Foreground *Color
	Background *Color
	Bold       bool
}

// Constructs a StyledString containing the given 'text' with the given
// 'color' and style 'modifier'.
func Stylize(text string, foreground *Color, background *Color, bold bool) StyledString {
	var runes = utf8.RuneCountInString(text)
	var result StyledString = make([]StyledRune, runes)
	var i int = 0
	for _, r := range text {
		result[i] = StyledRune{Style{foreground, background, bold}, r}
		i++
	}
	return result
}

// Constructs a StyledString containing the given 'text' and a "don't care"
// style. Good for use with whitespace.
func Unstyled(text string) StyledString {
	return Stylize(text, nil, nil, false)
}

// Formats a Style as an ANSI escape sequence and returns the escape sequence.
// See https://en.wikipedia.org/wiki/ANSI_escape_code#Escape_sequences.
func (self Style) toAnsi() string {
	// Start by clearing any pre-existing style.
	var commands = []string{"0"}

	if self.Foreground != nil {
		commands = append(commands, "38;2;"+self.Foreground.Join(";"))
	}
	if self.Background != nil {
		commands = append(commands, "48;2;"+self.Background.Join(";"))
	}
	if self.Bold {
		commands = append(commands, "1")
	}

	return "\033[" + strings.Join(commands, ";") + "m"
}

// Serializes this StyledString to a string with embedded ANSI escape
// sequences.
func (self StyledString) AnsiString() string {
	var buffer bytes.Buffer
	var first = true
	var lastStyle Style

	for _, r := range self {
		if first || lastStyle != r.Style {
			// The style is changing, so insert a new style escape.
			buffer.WriteString(r.Style.toAnsi())
			lastStyle = r.Style
		}

		buffer.WriteRune(r.Text)
    first = false
	}

	// Clear style before ending.
	if !first {
		buffer.WriteString("\033[0m")
	}
	return buffer.String()
}

// Returns just the text fro this StyledString, without any formatting.
func (self StyledString) PlainString() string {
	var buffer = bytes.NewBuffer(make([]byte, 0, len(self)))
	for _, r := range self {
		buffer.WriteRune(r.Text)
	}
	return buffer.String()
}
