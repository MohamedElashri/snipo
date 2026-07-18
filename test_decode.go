package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/MohamedElashri/snipo/internal/models"
	"github.com/MohamedElashri/snipo/internal/validation"
)

func main() {
	jsonData := []byte(`{"id":1,"app_name":"snipo","custom_css":"","theme":"auto","default_language":"plaintext","s3_enabled":false,"s3_endpoint":"","s3_bucket":"","s3_region":"us-east-1","backup_encryption_enabled":false,"created_at":"2026-07-18T14:12:06Z","updated_at":"2026-07-18T14:12:06Z","archive_enabled":false,"history_enabled":true,"editor_font_size":14,"editor_tab_size":2,"editor_theme":"auto","editor_word_wrap":true,"editor_show_print_margin":false,"editor_show_gutter":true,"editor_show_indent_guides":true,"editor_highlight_active_line":true,"editor_use_soft_tabs":true,"editor_enable_snippets":true,"editor_enable_live_autocompletion":true,"markdown_font_size":14,"disable_login":false,"exclude_first_line_on_copy":false,"trash_enabled":true,"auto_archive_enabled":false,"default_expiration_days":0}`)
	
	var input models.SettingsInput
	decoder := json.NewDecoder(bytes.NewReader(jsonData))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		fmt.Printf("Decode error: %v\n", err)
		return
	}
	
	errs := validation.ValidateSettingsInput(&input)
	if errs.HasErrors() {
		fmt.Printf("Validation errors: %v\n", errs)
	} else {
		fmt.Println("Validation successful")
	}
}
