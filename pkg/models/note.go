package models

const MaxNoteContentBytes = 1024 * 1024

type NoteAppendInput struct {
	FileID  int64  `json:"file_id,omitempty"`
	Content string `json:"content"`
}

type NoteTarget struct {
	SnippetID string `json:"snippet_id"`
	Title     string `json:"title"`
	FileID    int64  `json:"file_id,omitempty"`
	Filename  string `json:"filename,omitempty"`
}

type NoteTargets struct {
	Targets []NoteTarget `json:"targets"`
	HasMore bool         `json:"has_more"`
}
