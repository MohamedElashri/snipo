package ui

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/glamour"
	glamourStyles "github.com/charmbracelet/glamour/styles"
)

func init() {
	// Register a custom style that clones monokai but removes the background
	// from ALL tokens to ensure no dark boxes appear.
	baseStyle := styles.Get("monokai")
	if baseStyle != nil {
		builder := baseStyle.Builder()
		
		// 1. Unset global background
		bgEntry := builder.Get(chroma.Background)
		builder.Add(chroma.Background, bgEntry.Colour.String())

		// 2. Iterate over known token types to ensure consistent transparency
		tokens := []chroma.TokenType{

			chroma.Background,
			chroma.Text,
			chroma.Whitespace,
			chroma.Comment,
			chroma.Keyword,
			chroma.Name,
			chroma.Literal,
			chroma.Operator,
			chroma.Punctuation,
		}
		
		for _, t := range tokens {
			if entry := builder.Get(t); entry.Background.IsSet() {
				// Re-add with only foreground (removing background)
				builder.Add(t, entry.Colour.String())
			}
		}

		newStyle, err := builder.Build()
		if err == nil {
			newStyle.Name = "snipo-dark"
			styles.Register(newStyle)
		}
	}
}

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

	// Get the style - using snipo-dark if available, otherwise monokai
	style := styles.Get("snipo-dark")
	if style == nil {
		style = styles.Get("monokai")
	}
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

// IsMarkdown checks if the language or filename indicates markdown content
func IsMarkdown(language, filename string) bool {
	if language != "" {
		langLower := strings.ToLower(language)
		if langLower == "markdown" || langLower == "md" {
			return true
		}
	}

	if filename != "" {
		filenameLower := strings.ToLower(filename)
		if strings.HasSuffix(filenameLower, ".md") || strings.HasSuffix(filenameLower, ".markdown") {
			return true
		}
	}

	return false
}

// RenderMarkdown renders markdown content with proper formatting
func RenderMarkdown(content string, width int) string {
	// Create a custom style based on DarkStyleConfig but without backgrounds (nil)
	style := glamourStyles.DarkStyleConfig
	style.Code.StylePrimitive.BackgroundColor = nil
	style.CodeBlock.StylePrimitive.BackgroundColor = nil
	
	// Use our custom chroma theme that has no background
	style.CodeBlock.Theme = "snipo-dark"

	// Create a glamour renderer with the custom style
	r, err := glamour.NewTermRenderer(
		glamour.WithStyles(style),
		glamour.WithWordWrap(width),
	)

	if err != nil {
		// If renderer creation fails, return original content
		return content
	}

	// Render the markdown
	rendered, err := r.Render(content)
	if err != nil {
		// If rendering fails, return original content
		return content
	}

	return strings.TrimSpace(rendered)
}

// RenderContent renders content based on type (markdown or code with syntax highlighting)
func RenderContent(content, language, filename string, width int) string {
	// Check if this is markdown
	if IsMarkdown(language, filename) {
		return RenderMarkdown(content, width)
	}

	// Otherwise, apply syntax highlighting
	highlightLanguage := language
	if highlightLanguage == "" && filename != "" {
		highlightLanguage = GetLanguageFromFilename(filename)
	}

	return HighlightCode(content, highlightLanguage)
}


