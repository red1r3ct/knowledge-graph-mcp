package connection

import (
	"context"
)

//go:generate mockgen -source=storage.go -destination=mock/storage.go -package=mock

// Storage defines the interface for connection storage operations
type Storage interface {
	// Create creates a new connection
	Create(ctx context.Context, req CreateConnectionRequest) (*Connection, error)
	
	// Get retrieves a connection by ID
	Get(ctx context.Context, id int64) (*Connection, error)
	
	// Update updates an existing connection
	Update(ctx context.Context, id int64, req UpdateConnectionRequest) (*Connection, error)
	
	// Delete deletes a connection by ID
	Delete(ctx context.Context, id int64) error
	
	// List lists connections with pagination and filtering
	List(ctx context.Context, req ListConnectionsRequest) (*ListConnectionsResponse, error)
	
	// GetNoteConnections retrieves all connections for a specific note
	GetNoteConnections(ctx context.Context, req NoteConnectionsRequest) (*NoteConnectionsResponse, error)
	
	// GetConnectionsByType retrieves connections filtered by type
	GetConnectionsByType(ctx context.Context, connectionType string, req ListConnectionsRequest) (*ListConnectionsResponse, error)
	
	// GetBidirectionalConnections retrieves both incoming and outgoing connections for a note
	GetBidirectionalConnections(ctx context.Context, noteID int64) (*NoteConnectionsResponse, error)
	
	// GetConnectionStats retrieves statistics about connections
	GetConnectionStats(ctx context.Context) (*ConnectionStats, error)
	
	// FindConnectionPaths finds paths between two notes (for future graph traversal)
	FindConnectionPaths(ctx context.Context, fromNoteID, toNoteID int64, maxDepth int) ([]ConnectionPath, error)
}