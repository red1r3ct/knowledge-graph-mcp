package note

import (
	"context"
)

//go:generate mockgen -source=storage.go -destination=mock/storage.go -package=mock

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
}