package ui

import (
	"github.com/MohamedElashri/snipo/tui/internal/api"
	"github.com/MohamedElashri/snipo/tui/internal/config"
	"strings"
	"testing"
)

func TestUntrustedContentCannotEmitTerminalCommands(t *testing.T) {
	// Harmless markers: capture output bytes without sending commands to a terminal.
	payload := "before\x1b]2;SNIPO-TITLE-MARKER\a\x1b]52;c;bWFya2Vy\a\x1b[2Jafter"
	cases := map[string]string{
		"plain code":       HighlightCode(payload, "plaintext"),
		"highlighted code": HighlightCode(payload, "python"),
		"markdown":         RenderMarkdown(payload, 80),
	}
	snippet := api.Snippet{Title: payload, Description: payload, Content: payload, Language: "plaintext", Tags: []api.Tag{{Name: payload}}, Files: []api.File{{Filename: payload + ".txt", Content: payload}}}
	m := Model{mode: ViewList, width: 100, height: 40, config: &config.Config{}, snippets: []api.Snippet{snippet}, detailSnippet: &snippet}
	cases["list"] = m.View()
	m.mode = ViewDetail
	cases["detail"] = m.View()
	for name, rendered := range cases {
		for _, command := range []string{"\x1b]2;", "\x1b]52;", "\x1b[2J"} {
			if strings.Contains(rendered, command) {
				t.Errorf("%s emitted untrusted terminal command %q", name, command)
			}
		}
	}
}

func TestTerminalDisplayPreservesCodeDataAndLayout(t *testing.T) {
	payload := "    before\t\n\x1b]52;c;bWFya2Vy\a\u009b2Jafter\n"
	snippet := api.Snippet{Title: "Example", Content: payload, Language: "plaintext"}
	m := Model{mode: ViewDetail, width: 100, height: 40, config: &config.Config{}, detailSnippet: &snippet}
	displayed := m.View()
	if m.detailSnippet.Content != payload {
		t.Fatal("rendering modified original snippet data")
	}
	if strings.Contains(displayed, "\x1b]52;") || strings.ContainsRune(displayed, '\u009b') {
		t.Fatal("unsafe terminal controls in output")
	}
	if plain := terminalText("    日本語\t\r\nnext\n"); plain != "    日本語\t\nnext\n" {
		t.Errorf("layout/text changed: %q", plain)
	}
	// Preserve syntax highlighting while neutralizing controls introduced by parsing.
	styled := "\x1b[31mred\x1b[0m"
	if got := terminalOutput(styled, true); got != styled {
		t.Errorf("renderer styles changed: %q", got)
	}
	for _, markdown := range []string{"&#27;]52;c;bWFya2Vy&#7;", "[example](https://example.test/&#27;]2;MARKER&#7;)"} {
		rendered := RenderMarkdown(markdown, 80)
		if strings.Contains(rendered, "\x1b]52;") || strings.Contains(rendered, "\x1b]2;") {
			t.Fatal("markdown decoding introduced terminal commands")
		}
	}
}
