package repository

import "errors"

// Common repository errors
var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrFolderCycle   = errors.New("cannot move folder: would create circular reference")
)
