package main

import (
"fmt"
"github.com/charmbracelet/bubbles/textinput"
)

func main() {
	m := textinput.New()
	fmt.Printf("AcceptSuggestion: %v\n", m.KeyMap.AcceptSuggestion.Keys())
	fmt.Printf("NextSuggestion: %v\n", m.KeyMap.NextSuggestion.Keys())
	fmt.Printf("PrevSuggestion: %v\n", m.KeyMap.PrevSuggestion.Keys())
}
