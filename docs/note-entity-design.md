# Note Entity Design Document

## Overview
This document provides a comprehensive design for the Note entity that will be integrated into the knowledge graph system. The Note entity represents individual pieces of content that can be connected to form a knowledge graph.

## Entity Definition and Fields

### Domain Model: Note
The Note entity represents a single piece of knowledge or content within the system.

```go
type Note struct {
    ID          int64      `json:"id"`
    Title       string     `json:"title"`
    Content     string     `json:"content"`
    Type        string     `json:"type"` // e.g., "text", "markdown", "code", "link"
    Tags        []string   `json:"tags,omitempty"`
    Metadata    JSON       `json:"metadata,omitempty"` // Flexible JSON field for additional properties
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
}
```

### Field Descriptions
- **ID**: Unique identifier for the note (auto-incrementing primary key)
- **Title**: Human-readable title of the note (required, max 255 characters)
- **Content**: The actual content/body of the note (required, text field)
- **Type**: Classification of the note type for filtering and processing
- **Tags**: Array of string tags for categorization and search
- **Metadata**: Flexible JSON field for storing additional properties (e.g., source URL, author, custom fields)
- **CreatedAt**: Timestamp when the note was created
- **UpdatedAt**: Timestamp when the note was last modified

## CRUDL Operations Specification

### Create Operation
Creates a new note in the system.

**Request:**
```go
type CreateNoteRequest struct {
    Title    string   `json:"title"`
    Content  string   `json:"content"`
    Type     string   `json:"type"`
    Tags     []string `json:"tags,omitempty"`
    Metadata JSON     `json:"metadata,omitempty"`
}
```

**Validation Rules:**
- Title: required, 1-255 characters
- Content: required, non-empty
- Type: required, must be one of predefined values: "text", "markdown", "code", "link", "image"
- Tags: optional, each tag max 50 characters
- Metadata: optional, valid JSON object

### Read Operation
Retrieves a single note by ID.

**Request:**
```go
type GetNoteRequest struct {
    ID int64 `json:"id"`
}
```

**Response:**
```go
type NoteResponse struct {
    Note *Note `json:"note"`
}
```

### Update Operation
Updates an existing note.

**Request:**
```go
type UpdateNoteRequest struct {
    Title    *string  `json:"title,omitempty"`
    Content  *string  `json:"content,omitempty"`
    Type     *string  `json:"type,omitempty"`
    Tags     []string `json:"tags,omitempty"`
    Metadata JSON     `json:"metadata,omitempty"`
}
```

**Validation Rules:**
- All fields optional, but at least one must be provided
- Same validation rules as Create for individual fields
- Partial updates supported

### Delete Operation
Deletes a note by ID.

**Request:**
```go
type DeleteNoteRequest struct {
    ID int64 `json:"id"`
}
```

**Behavior:**
- Soft delete (mark as deleted) or hard delete based on configuration
- Cascade delete connections associated with this note

### List Operation
Lists notes with pagination and filtering.

**Request:**
```go
type ListNotesRequest struct {
    Limit    int      `json:"limit,omitempty"`
    Offset   int      `json:"offset,omitempty"`
    Search   string   `json:"search,omitempty"`
    Tags     []string `json:"tags,omitempty"`
    Type     string   `json:"type,omitempty"`
    OrderBy  string   `json:"order_by,omitempty"` // "created_at", "updated_at", "title"
    OrderDir string   `json:"order_dir,omitempty"` // "asc", "desc"
}
```

**Response:**
```go
type ListNotesResponse struct {
    Items []Note `json:"items"`
    Total int64  `json:"total"`
}
```

## Domain Models and DTOs

### Core Domain Models
```go
package note

import "time"

// Note represents the core domain model
type Note struct {
    ID          int64
    Title       string
    Content     string
    Type        string
    Tags        []string
    Metadata    map[string]interface{}
    CreatedAt   time.Time
    UpdatedAt   time.Time
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
```

### Data Transfer Objects (DTOs)
```go
// CreateNoteRequest represents the request to create a note
type CreateNoteRequest struct {
    Title    string
    Content  string
    Type     string
    Tags     []string
    Metadata map[string]interface{}
}

// UpdateNoteRequest represents the request to update a note
type UpdateNoteRequest struct {
    Title    *string
    Content  *string
    Type     *string
    Tags     []string
    Metadata map[string]interface{}
}

// ListNotesRequest represents the request to list notes
type ListNotesRequest struct {
    Limit    int
    Offset   int
    Search   string
    Tags     []string
    Type     string
    OrderBy  string
    OrderDir string
}

// ListNotesResponse represents the response for listing notes
type ListNotesResponse struct {
    Items []Note
    Total int64
}
```

## Storage Interface Design

### Interface Definition
```go
package note

import "context"

// Storage defines the interface for note storage operations
type Storage interface {
    // Create creates a new note
    Create(ctx context.Context, req CreateNoteRequest) (*Note, error)
    
    // Get retrieves a note by ID
    Get(ctx context.Context, id int64) (*Note, error)
    
    // Update updates an existing note
    Update(ctx context.Context, id int64, req UpdateNoteRequest) (*Note, error)
    
    // Delete deletes a note by ID
    Delete(ctx context.Context, id int64) error
    
    // List lists notes with pagination and filtering
    List(ctx context.Context, req ListNotesRequest) (*ListNotesResponse, error)
    
    // GetByTitle retrieves a note by title (for uniqueness checks)
    GetByTitle(ctx context.Context, title string) (*Note, error)
    
    // Search performs full-text search on notes
    Search(ctx context.Context, query string, limit, offset int) (*ListNotesResponse, error)
}
```

### Database Schema
```sql
CREATE TABLE notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('text', 'markdown', 'code', 'link', 'image')),
    tags JSON,
    metadata JSON,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(title)
);

CREATE INDEX idx_notes_type ON notes(type);
CREATE INDEX idx_notes_created_at ON notes(created_at);
CREATE INDEX idx_notes_updated_at ON notes(updated_at);

-- Full-text search index
CREATE VIRTUAL TABLE notes_fts USING fts5(
    title,
    content,
    content='notes',
    content_rowid='id'
);
```

## Test Cases and Acceptance Criteria

### Unit Test Cases

#### TestCreateNote
```go
func TestCreateNote(t *testing.T) {
    tests := []struct {
        name        string
        req         CreateNoteRequest
        want        *Note
        wantErr     bool
        errMessage  string
    }{
        {
            name: "valid note",
            req: CreateNoteRequest{
                Title:   "Test Note",
                Content: "This is a test note",
                Type:    "text",
            },
            want: &Note{
                Title:   "Test Note",
                Content: "This is a test note",
                Type:    "text",
            },
            wantErr: false,
        },
        {
            name: "empty title",
            req: CreateNoteRequest{
                Title:   "",
                Content: "Content",
                Type:    "text",
            },
            wantErr:    true,
            errMessage: "title is required",
        },
        {
            name: "empty content",
            req: CreateNoteRequest{
                Title:   "Title",
                Content: "",
                Type:    "text",
            },
            wantErr:    true,
            errMessage: "content is required",
        },