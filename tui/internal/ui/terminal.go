package ui

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// terminalText makes untrusted text visible without executing terminal controls.
// Keep code layout; never apply this to saved files or clipboard contents.
func terminalText(text string) string { return terminalOutput(text, false) }

// terminalOutput preserves only renderer-generated SGR color/style sequences.
// Markdown decoding may introduce controls that were absent from its source.
func terminalOutput(text string, allowStyles bool) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	var out strings.Builder
	for i := 0; i < len(text); {
		if allowStyles && text[i] == '\x1b' && i+1 < len(text) && text[i+1] == '[' {
			end := i + 2
			for end < len(text) && ((text[end] >= '0' && text[end] <= '9') || text[end] == ';' || text[end] == ':') {
				end++
			}
			if end < len(text) && text[end] == 'm' {
				out.WriteString(text[i : end+1])
				i = end + 1
				continue
			}
		}
		r, size := utf8.DecodeRuneInString(text[i:])
		i += size
		if (r < 32 && r != '\n' && r != '\t') || (r >= 0x7f && r <= 0x9f) {
			fmt.Fprintf(&out, "\\x%02x", r)
		} else {
			out.WriteRune(r)
		}
	}
	return out.String()
}
