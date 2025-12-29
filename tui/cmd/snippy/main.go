package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/MohamedElashri/snipo/tui/internal/app"
	"github.com/MohamedElashri/snipo/tui/internal/config"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "config" {
		if err := runConfigWizard(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runConfigWizard() error {
	reader := bufio.NewReader(os.Stdin)
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	fmt.Println("Snippy Configuration Wizard")
	fmt.Println("---------------------------")

	fmt.Printf("Server URL [%s]: ", cfg.ServerURL)
	url, _ := reader.ReadString('\n')
	url = strings.TrimSpace(url)
	if url != "" {
		cfg.ServerURL = url
	}

	keyHint := ""
	if cfg.APIKey != "" {
		keyHint = "(set)"
	}
	fmt.Printf("API Key %s: ", keyHint) 
	key, _ := reader.ReadString('\n')
	key = strings.TrimSpace(key)
	if key != "" {
		cfg.APIKey = key
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("Configuration saved successfully!")
	return nil
}
