# Connection Entity Design Document

## Overview
This document provides a comprehensive design for the Connection entity that establishes relationships between Notes in the knowledge graph system. Connections represent semantic links between notes, enabling the creation of a connected knowledge graph.

## Entity Definition and Relationship to Notes

### Domain Model: Connection
The Connection entity represents a directed relationship between two notes with semantic meaning.

```go
type Connection struct {
    ID          int64      `json:"id"`
    FromNoteID  int64      `json:"from_note_id"`
    ToNoteID    int64      `json:"to_note_id"`
    Type        string     `json:"type"` // e.g., "relates_to", "references", "contradicts", "supports"
    Description *string    `json:"description,omitempty"`
    Strength    int        `json:"strength"` // 1-10 scale for relationship strength
    Metadata    JSON       `json:"metadata,omitempty"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
}
```

### Connection Types
```go
type ConnectionType string

const (
    ConnectionTypeRelatesTo    ConnectionType = "relates_to"
    ConnectionTypeReferences   ConnectionType = "references"
    ConnectionTypeSupports     ConnectionType = "supports"
    ConnectionTypeContradicts  ConnectionType = "contradicts"
    ConnectionTypeInfluences   ConnectionType = "influences"
    ConnectionTypeDependsOn    ConnectionType = "depends_on"
    ConnectionTypeSimilarTo    ConnectionType = "similar_to"
    ConnectionTypePartOf       ConnectionType = "part_of"
    ConnectionTypeCites        ConnectionType = "cites"
    ConnectionTypeFollows      ConnectionType = "follows"
    ConnectionTypePrecedes     ConnectionType = "precedes"
)
```

### Relationship to Notes
- **FromNoteID**: The source note in the relationship (outgoing connection)
- **ToNoteID**: The target note in the relationship (incoming connection)
- **Directed Graph**: Connections are directional (A→B is different from B→A)
- **Multiple Connections**: Multiple connections can exist between the same pair of notes with different types
- **Self-Connections**: Notes can have connections to themselves (useful for self-references)

## CRUDL Operations Specification

### Create Connection
Creates a new connection between two notes.

**Request:**
```go
type CreateConnectionRequest struct {
    FromNoteID  int64   `json:"from_note_id"`
    ToNoteID    int64   `json:"to_note_id"`
    Type        string  `json:"type"`
    Description *string `json:"description,omitempty"`
    Strength    int     `json:"strength"` // 1-10
    Metadata    JSON    `json:"metadata,omitempty"`
}
```

**Validation Rules:**
- FromNoteID: required, must exist in notes table
- ToNoteID: required, must exist in notes table
- Type: required, must be one of predefined connection types
- Strength: optional, must be between 1 and 10 (default: 5)
- Description: optional, max 500 characters
- Prevent duplicate connections: (from_note_id, to_note_id, type) must be unique

### Read Connection
Retrieves a single connection by ID.

**Request:**
```go
type GetConnectionRequest struct {
    ID int64 `json:"id"`
}
```

### Update Connection
Updates an existing connection's properties.

**Request:**
```go
type UpdateConnectionRequest struct {
    Type        *string `json:"type,omitempty"`
    Description *string `json:"description,omitempty"`
    Strength    *int    `json:"strength,omitempty"`
    Metadata    JSON    `json:"metadata,omitempty"`
}
```

### Delete Connection
Deletes a connection by ID.

**Request:**
```go
type DeleteConnectionRequest struct {
    ID int64 `json:"id"`
}
```

### List Connections
Lists connections with various filtering options.

**Request:**
```go
type ListConnectionsRequest struct {
    Limit      int      `json:"limit,omitempty"`
    Offset     int      `json:"offset,omitempty"`
    FromNoteID *int64   `json:"from_note_id,omitempty"`
    ToNoteID   *int64   `json:"to_note_id,omitempty"`
    Type       *string  `json:"type,omitempty"`
    Strength   *int     `json:"strength,omitempty"`
    OrderBy    string   `json:"order_by,omitempty"`
    OrderDir   string   `json:"order_dir,omitempty"`
}
```

## Domain Models and DTOs

### Core Domain Models
```go
package connection

import "time"

// Connection represents the core domain model
type Connection struct {
    ID          int64
    FromNoteID  int64
    ToNoteID    int64
    Type        string
    Description *string
    Strength    int
    Metadata    map[string]interface{}
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// ConnectionType represents the type of relationship
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
```

### Data Transfer Objects
```go
// CreateConnectionRequest represents the request to create a connection
type CreateConnectionRequest struct {
    FromNoteID  int64
    ToNoteID    int64
    Type        string
    Description *string
    Strength    int
    Metadata    map[string]interface{}
}

// UpdateConnectionRequest represents the request to update a connection
type UpdateConnectionRequest struct {
    Type        *string
    Description *string
    Strength    *int
    Metadata    map[string]interface{}
}

// ListConnectionsRequest represents the request to list connections
type ListConnectionsRequest struct {
    Limit      int
    Offset     int
    FromNoteID *int64
    ToNoteID   *int64
    Type       *string
    Strength   *int
    OrderBy    string
    OrderDir   string
}

// ListConnectionsResponse represents the response for listing connections
type ListConnectionsResponse struct {
    Items []Connection
    Total int64
}

// NoteConnections represents all connections for a specific note
type NoteConnections struct {
    NoteID       int64
    Outgoing     []Connection // Connections FROM this note
    Incoming     []Connection // Connections TO this note
    TotalCount   int64
    TypesCount   map[string]int64 // Count by connection type
}

// ConnectionPath represents a path between two notes
type ConnectionPath struct {
    FromNoteID int64
    ToNoteID   int64
    Path       []Connection // Ordered list of connections forming the path
    Length     int          // Number of connections in the path
    Strength   int          // Combined strength of the path
}
```

## Storage Interface Design

### Interface Definition
```go
package connection

import "context"

// Storage defines the interface for