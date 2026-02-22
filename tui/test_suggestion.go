package main

import (
"fmt"
"github.com/charmbracelet/bubbles/textinput"
)

func main() {
	m := textinput.New()
	m.ShowSuggestions = true
	m.SetSuggestions([]string{"go", "python"})
	m.SetValue("g")
	fmt.Printf("Suggestion: %v\n", m.CurrentSuggestion())
}
