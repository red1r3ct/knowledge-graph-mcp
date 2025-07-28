package knowledgebase

import (
	"context"
)

//go:generate mockgen -source=storage.go -destination=mock/storage.go -package=mock

// Storage defines the interface for knowledge base storage operations
type Storage interface {
	// Create creates a new knowledge base
	Create(ctx context.Context, req CreateRequest) (*KnowledgeBase, error)
	
	// Get retrieves a knowledge base by ID
	Get(ctx context.Context, id int64) (*KnowledgeBase, error)
	
	// Update updates an existing knowledge base
	Update(ctx context.Context, id int64, req UpdateRequest) (*KnowledgeBase, error)
	
	// Delete deletes a knowledge base by ID
	Delete(ctx context.Context, id int64) error
	
	// List lists knowledge bases with pagination and filtering
	List(ctx context.Context, req ListRequest) (*ListResponse, error)
}