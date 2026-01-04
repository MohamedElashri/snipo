-- Add exclude_first_line_on_copy setting
-- Migration: 006

ALTER TABLE settings ADD COLUMN exclude_first_line_on_copy BOOLEAN DEFAULT FALSE;
