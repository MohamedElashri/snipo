package ui

import (
	"strings"
)

// visualWidth calculates the visual width of a string (accounting for ANSI codes)
func visualWidth(s string) int {
	// Simple ANSI stripping - count actual visible characters
	inEscape := false
	width := 0

	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEscape = false
			}
			continue
		}
		width++
	}

	return width
}

// wrapLine wraps a single line at the specified width, preserving ANSI codes
func wrapLine(line string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{line}
	}

	// If line fits, return as-is
	if visualWidth(line) <= maxWidth {
		return []string{line}
	}

	// For lines with ANSI codes, we need to be careful
	// Simple approach: break at maxWidth visible characters
	var result []string
	var currentLine strings.Builder
	var currentWidth int
	inEscape := false
	var escapeSeq strings.Builder

	for _, r := range line {
		if r == '\x1b' {
			inEscape = true
			escapeSeq.Reset()
			escapeSeq.WriteRune(r)
			continue
		}

		if inEscape {
			escapeSeq.WriteRune(r)
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEscape = false
				currentLine.WriteString(escapeSeq.String())
			}
			continue
		}

		// Regular character
		if currentWidth >= maxWidth {
			// Start new line, preserve formatting by adding escape sequences
			result = append(result, currentLine.String())
			currentLine.Reset()
			currentWidth = 0
		}

		currentLine.WriteRune(r)
		currentWidth++
	}

	if currentLine.Len() > 0 {
		result = append(result, currentLine.String())
	}

	if len(result) == 0 {
		return []string{line}
	}

	return result
}

// wrapContent wraps all lines in content at the specified width
func wrapContent(content string, maxWidth int) string {
	if maxWidth <= 0 {
		return content
	}

	lines := strings.Split(content, "\n")
	var wrappedLines []string

	for _, line := range lines {
		wrapped := wrapLine(line, maxWidth)
		wrappedLines = append(wrappedLines, wrapped...)
	}

	return strings.Join(wrappedLines, "\n")
}

// padLinesToWidth pads all lines to the specified width for consistent rendering
func padLinesToWidth(lines []string, width int) []string {
	paddedLines := make([]string, len(lines))

	for i, line := range lines {
		lineWidth := visualWidth(line)
		if lineWidth < width {
			// Pad with spaces
			padding := strings.Repeat(" ", width-lineWidth)
			paddedLines[i] = line + padding
		} else {
			paddedLines[i] = line
		}
	}

	return paddedLines
}

// calculateMaxLineWidth calculates the maximum visual width of all lines
func calculateMaxLineWidth(lines []string) int {
	maxWidth := 0
	for _, line := range lines {
		width := visualWidth(line)
		if width > maxWidth {
			maxWidth = width
		}
	}
	return maxWidth
}
