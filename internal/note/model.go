package note

import (
	"time"
)

// Note represents the domain model for a note entity
type Note struct {
	ID        int64                  `json:"id"`
	Title     string                 `json:"title"`
	Content   string                 `json:"content"`
	Type      string                 `json:"type"`
	Tags      []string               `json:"tags,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// NoteType represents the type classification of a note
type NoteType string

const (
	NoteTypeText     NoteType = "text"
	NoteTypeMarkdown NoteType = "markdown"
	NoteTypeCode     NoteType = "code"
	NoteTypeLink     NoteType = "link"
	NoteTypeImage    NoteType = "image"
)

// CreateNoteRequest represents the DTO for creating a note
type CreateNoteRequest struct {
	Title    string                 `json:"title"`
	Content  string                 `json:"content"`
	Type     string                 `json:"type"`
	Tags     []string               `json:"tags,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateNoteRequest represents the DTO for updating a note
type UpdateNoteRequest struct {
	Title    *string                `json:"title,omitempty"`
	Content  *string                `json:"content,omitempty"`
	Type     *string                `json:"type,omitempty"`
	Tags     []string               `json:"tags,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ListNotesRequest represents the DTO for listing notes
type ListNotesRequest struct {
	Limit    int      `json:"limit,omitempty"`
	Offset   int      `json:"offset,omitempty"`
	Search   string   `json:"search,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Type     string   `json:"type,omitempty"`
	OrderBy  string   `json:"order_by,omitempty"`
	OrderDir string   `json:"order_dir,omitempty"`
}

// ListNotesResponse represents the DTO for listing response
type ListNotesResponse struct {
	Items []Note `json:"items"`
	Total int64  `json:"total"`
}