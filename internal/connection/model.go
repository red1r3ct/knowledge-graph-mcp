package connection

import (
	"time"
)

// Connection represents the domain model for a connection entity
type Connection struct {
	ID          int64                  `json:"id"`
	FromNoteID  int64                  `json:"from_note_id"`
	ToNoteID    int64                  `json:"to_note_id"`
	Type        string                 `json:"type"`
	Description *string                `json:"description,omitempty"`
	Strength    int                    `json:"strength"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ConnectionType represents the type of relationship between notes
type ConnectionType string

const (
	ConnectionTypeRelatesTo   ConnectionType = "relates_to"
	ConnectionTypeReferences  ConnectionType = "references"
	ConnectionTypeSupports    ConnectionType = "supports"
	ConnectionTypeContradicts ConnectionType = "contradicts"
	ConnectionTypeInfluences  ConnectionType = "influences"
	ConnectionTypeDependsOn   ConnectionType = "depends_on"
	ConnectionTypeSimilarTo   ConnectionType = "similar_to"
	ConnectionTypePartOf      ConnectionType = "part_of"
	ConnectionTypeCites       ConnectionType = "cites"
	ConnectionTypeFollows     ConnectionType = "follows"
	ConnectionTypePrecedes    ConnectionType = "precedes"
)

// ValidConnectionTypes returns a slice of all valid connection types
func ValidConnectionTypes() []string {
	return []string{
		string(ConnectionTypeRelatesTo),
		string(ConnectionTypeReferences),
		string(ConnectionTypeSupports),
		string(ConnectionTypeContradicts),
		string(ConnectionTypeInfluences),
		string(ConnectionTypeDependsOn),
		string(ConnectionTypeSimilarTo),
		string(ConnectionTypePartOf),
		string(ConnectionTypeCites),
		string(ConnectionTypeFollows),
		string(ConnectionTypePrecedes),
	}
}

// IsValidConnectionType checks if the given type is a valid connection type
func IsValidConnectionType(connectionType string) bool {
	for _, validType := range ValidConnectionTypes() {
		if validType == connectionType {
			return true
		}
	}
	return false
}

// CreateConnectionRequest represents the DTO for creating a connection
type CreateConnectionRequest struct {
	FromNoteID  int64                  `json:"from_note_id"`
	ToNoteID    int64                  `json:"to_note_id"`
	Type        string                 `json:"type"`
	Description *string                `json:"description,omitempty"`
	Strength    int                    `json:"strength"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateConnectionRequest represents the DTO for updating a connection
type UpdateConnectionRequest struct {
	Type        *string                `json:"type,omitempty"`
	Description *string                `json:"description,omitempty"`
	Strength    *int                   `json:"strength,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ListConnectionsRequest represents the DTO for listing connections
type ListConnectionsRequest struct {
	Limit      int     `json:"limit,omitempty"`
	Offset     int     `json:"offset,omitempty"`
	FromNoteID *int64  `json:"from_note_id,omitempty"`
	ToNoteID   *int64  `json:"to_note_id,omitempty"`
	Type       *string `json:"type,omitempty"`
	Strength   *int    `json:"strength,omitempty"`
	OrderBy    string  `json:"order_by,omitempty"`
	OrderDir   string  `json:"order_dir,omitempty"`
}

// ListConnectionsResponse represents the DTO for listing response
type ListConnectionsResponse struct {
	Items []Connection `json:"items"`
	Total int64        `json:"total"`
}

// NoteConnectionsRequest represents the DTO for getting all connections for a specific note
type NoteConnectionsRequest struct {
	NoteID   int64   `json:"note_id"`
	Type     *string `json:"type,omitempty"`
	Strength *int    `json:"strength,omitempty"`
	Limit    int     `json:"limit,omitempty"`
	Offset   int     `json:"offset,omitempty"`
}

// NoteConnectionsResponse represents all connections for a specific note
type NoteConnectionsResponse struct {
	NoteID     int64              `json:"note_id"`
	Outgoing   []Connection       `json:"outgoing"`   // Connections FROM this note
	Incoming   []Connection       `json:"incoming"`   // Connections TO this note
	TotalCount int64              `json:"total_count"`
	TypesCount map[string]int64   `json:"types_count"` // Count by connection type
}

// ConnectionPath represents a path between two notes through connections
type ConnectionPath struct {
	FromNoteID int64        `json:"from_note_id"`
	ToNoteID   int64        `json:"to_note_id"`
	Path       []Connection `json:"path"`     // Ordered list of connections forming the path
	Length     int          `json:"length"`   // Number of connections in the path
	Strength   int          `json:"strength"` // Combined strength of the path
}

// ConnectionStats represents statistics about connections
type ConnectionStats struct {
	TotalConnections     int64            `json:"total_connections"`
	ConnectionsByType    map[string]int64 `json:"connections_by_type"`
	ConnectionsByStrength map[int]int64   `json:"connections_by_strength"`
	MostConnectedNotes   []NoteConnection `json:"most_connected_notes"`
}

// NoteConnection represents a note with its connection count
type NoteConnection struct {
	NoteID          int64 `json:"note_id"`
	IncomingCount   int64 `json:"incoming_count"`
	OutgoingCount   int64 `json:"outgoing_count"`
	TotalCount      int64 `json:"total_count"`
}