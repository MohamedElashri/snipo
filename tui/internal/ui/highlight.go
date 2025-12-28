package ui

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// HighlightCode applies syntax highlighting to code based on the language
func HighlightCode(code, language string) string {
	// Get the lexer for the specified language
	var lexer chroma.Lexer
	if language != "" {
		lexer = lexers.Get(language)
	}

	// Fallback to analyzing the code if no language specified or lexer not found
	if lexer == nil {
		lexer = lexers.Analyse(code)
	}

	// Final fallback to plain text
	if lexer == nil {
		lexer = lexers.Fallback
	}

	// Coalesce the lexer to ensure it's properly initialized
	lexer = chroma.Coalesce(lexer)

	// Get the style - using monokai which works well in terminals
	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	// Create a terminal formatter with 256 colors
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	// Tokenize the code
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		// If tokenization fails, return the original code
		return code
	}

	// Format the tokens
	var buf strings.Builder
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		// If formatting fails, return the original code
		return code
	}

	return buf.String()
}

// GetLanguageFromFilename extracts language from filename extension
func GetLanguageFromFilename(filename string) string {
	if filename == "" {
		return ""
	}

	// Get lexer by filename
	lexer := lexers.Match(filename)
	if lexer != nil {
		config := lexer.Config()
		if config != nil && len(config.Aliases) > 0 {
			return config.Aliases[0]
		}
	}

	return ""
}

// CreateHighlightedCodeBlock wraps highlighted code in a styled block
func CreateHighlightedCodeBlock(code, language string) string {
	highlighted := HighlightCode(code, language)

	// Apply the code block style
	return codeBlockStyle.Render(highlighted)
}

// CreateHighlightedCodeBlockWithFilename creates a highlighted code block with filename context
func CreateHighlightedCodeBlockWithFilename(code, filename string) string {
	// Try to get language from filename if not explicitly provided
	language := GetLanguageFromFilename(filename)
	return CreateHighlightedCodeBlock(code, language)
}
