-- Add disable_login column to settings table
-- This allows users to disable the login page while keeping password protection
-- for administrative operations like API token creation

ALTER TABLE settings ADD COLUMN disable_login INTEGER DEFAULT 0 NOT NULL;
