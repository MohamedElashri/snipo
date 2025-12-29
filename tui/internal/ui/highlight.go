package ui

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func init() {
	// Register custom styles that clone existing ones but remove the background
	// from ALL tokens to ensure no dark boxes appear.
	registerTransparentStyle("monokai", "snipo-dark")
	registerTransparentStyle("friendly", "snipo-light")
}

func registerTransparentStyle(baseName, newName string) {
	baseStyle := styles.Get(baseName)
	if baseStyle != nil {
		builder := baseStyle.Builder()
		
		// 1. Unset global background
		bgEntry := builder.Get(chroma.Background)
		builder.Add(chroma.Background, bgEntry.Colour.String())

		// 2. Override specific tokens to ensure visibility (No Grey)
		// Set Comments to Green (#00aa00) which maps to ANSI 2
		builder.Add(chroma.Comment, "#00aa00")
		builder.Add(chroma.CommentPreproc, "#00aa00")
		builder.Add(chroma.CommentSingle, "#00aa00")
		builder.Add(chroma.CommentSpecial, "#00aa00")
		builder.Add(chroma.CommentMultiline, "#00aa00")

		// 3. Iterate over other known token types to ensure transparency
		tokens := []chroma.TokenType{
			chroma.Background,
			chroma.Text,
			chroma.Whitespace,
			// Comments handled above
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
			newStyle.Name = newName
			styles.Register(newStyle)
		}
	}
}

// IsDarkMode detects if the terminal has a dark background
func IsDarkMode() bool {
	return lipgloss.HasDarkBackground()
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

	// Determine style based on background
	styleName := "snipo-dark"
	fallbackName := "monokai"
	if !IsDarkMode() {
		styleName = "snipo-light"
		fallbackName = "friendly"
	}

	// Get the style
	style := styles.Get(styleName)
	if style == nil {
		style = styles.Get(fallbackName)
	}
	if style == nil {
		style = styles.Fallback
	}

	// Create a terminal formatter (ANSI 16 colors) to respect user terminal theme
	formatter := formatters.Get("terminal") 
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

// getUniversalANSIStyle returns a glamour style config strictly using ANSI 0-15 colors
func getUniversalANSIStyle() ansi.StyleConfig {
	s := ansi.StyleConfig{}
	
	// Headers - Magenta (ANSI 5)
	headerColor := pointer("5")
	s.H1.Color = headerColor
	s.H1.Bold = pointer(true)
	s.H2.Color = headerColor
	s.H2.Bold = pointer(true)
	s.H3.Color = headerColor
	s.H3.Bold = pointer(true)
	
	// Links - Blue (ANSI 4)
	s.Link.Color = pointer("4")
	s.LinkText.Color = pointer("4")
	
	// Code - Cyan (ANSI 6) for inline, no background
	s.Code.Color = pointer("6")
	s.Code.BackgroundColor = nil
	s.Code.BlockPrefix = ""
	s.Code.BlockSuffix = ""
	
	// Code Block - Transparent, syntax highlighted
	s.CodeBlock.BackgroundColor = nil
	s.CodeBlock.Margin = pointer(uint(0))
	
	// Text - Default (nil) means strictly terminal foreground
	// Emphasis/Strong
	s.Strong.Bold = pointer(true)
	s.Emph.Italic = pointer(true)
	s.Emph.Color = pointer("3") // Yellow for emphasis instead of grey
	
	// Lists
	s.Item.Color = pointer("5") // Magenta bullet points
	s.Enumeration.Color = pointer("5")

	// BlockQuote - Blue (ANSI 4) instead of grey
	s.BlockQuote.Color = pointer("4")
	s.BlockQuote.Indent = pointer(uint(1))
	
	// Horizontal Rule - Magenta (ANSI 5)
	s.HorizontalRule.Color = pointer("5")
	
	// Table - Magenta (ANSI 5)
	s.Table.Color = pointer("5")
	
	return s
}

func pointer[T any](v T) *T {
	return &v
}

// RenderMarkdown renders markdown content with proper formatting
func RenderMarkdown(content string, width int) string {
	// Use universal ANSI style
	styleConfig := getUniversalANSIStyle()
	
	// Maintain dynamic chroma theme selection for best contrast within code blocks
	// even though the container is transparent.
	themeName := "snipo-dark"
	if !IsDarkMode() {
		themeName = "snipo-light"
	}
	styleConfig.CodeBlock.Theme = themeName

	// Enforce ANSI color profile
	r, err := glamour.NewTermRenderer(
		glamour.WithStyles(styleConfig),
		glamour.WithWordWrap(width),
		glamour.WithColorProfile(termenv.ANSI),
	)

	if err != nil {
		return content
	}

	rendered, err := r.Render(content)
	if err != nil {
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
