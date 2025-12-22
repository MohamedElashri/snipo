-- Add editor settings to settings table
-- Version: 4

-- Add editor-related settings columns
ALTER TABLE settings ADD COLUMN editor_font_size INTEGER DEFAULT 14;
ALTER TABLE settings ADD COLUMN editor_tab_size INTEGER DEFAULT 2;
ALTER TABLE settings ADD COLUMN editor_theme TEXT DEFAULT 'auto';
ALTER TABLE settings ADD COLUMN editor_word_wrap INTEGER DEFAULT 1;
ALTER TABLE settings ADD COLUMN editor_show_print_margin INTEGER DEFAULT 0;
ALTER TABLE settings ADD COLUMN editor_show_gutter INTEGER DEFAULT 1;
ALTER TABLE settings ADD COLUMN editor_show_indent_guides INTEGER DEFAULT 1;
ALTER TABLE settings ADD COLUMN editor_highlight_active_line INTEGER DEFAULT 1;
ALTER TABLE settings ADD COLUMN editor_use_soft_tabs INTEGER DEFAULT 1;
ALTER TABLE settings ADD COLUMN editor_enable_snippets INTEGER DEFAULT 1;
ALTER TABLE settings ADD COLUMN editor_enable_live_autocompletion INTEGER DEFAULT 1;
