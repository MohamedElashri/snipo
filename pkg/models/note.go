package models

const MaxNoteContentBytes = 1024 * 1024

const (
	KindSnippet = "snippet"
	KindNote    = "note"
)

func IsNoteFormat(language string) bool {
	return language == "markdown" || language == "plaintext"
}

type NoteAppendInput struct {
	RequestID string `json:"request_id,omitempty"`
	Language  string `json:"language,omitempty"`
	FileID    int64  `json:"file_id,omitempty"`
	Content   string `json:"content"`
}

type NoteTarget struct {
	Kind      string `json:"kind"`
	Language  string `json:"language"`
	SnippetID string `json:"snippet_id"`
	Title     string `json:"title"`
	FileID    int64  `json:"file_id,omitempty"`
	Filename  string `json:"filename,omitempty"`
}

type NoteTargets struct {
	Targets []NoteTarget `json:"targets"`
	HasMore bool         `json:"has_more"`
}
