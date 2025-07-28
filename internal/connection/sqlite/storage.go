package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"

	"github.com/red1r3ct/knowledge-graph-mcp/internal/connection"
)

// Storage implements the connection.Storage interface using SQLite
type Storage struct {
	db *sql.DB
}

// NewStorage creates a new SQLite storage instance
func NewStorage(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Storage{db: db}, nil
}

// Close closes the database connection
func (s *Storage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Create creates a new connection
func (s *Storage) Create(ctx context.Context, req connection.CreateConnectionRequest) (*connection.Connection, error) {
	// Validate connection type
	if !connection.IsValidConnectionType(req.Type) {
		return nil, fmt.Errorf("invalid connection type: %s", req.Type)
	}

	// Validate strength
	if req.Strength < 1 || req.Strength > 10 {
		return nil, fmt.Errorf("strength must be between 1 and 10, got: %d", req.Strength)
	}

	// Validate description length
	if req.Description != nil && len(*req.Description) > 500 {
		return nil, fmt.Errorf("description must be 500 characters or less")
	}

	// Serialize metadata
	var metadataJSON string
	if req.Metadata != nil {
		metadataBytes, err := json.Marshal(req.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataJSON = string(metadataBytes)
	} else {
		metadataJSON = "{}"
	}

	query := `
		INSERT INTO connections (from_note_id, to_note_id, type, description, strength, metadata)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.ExecContext(ctx, query, req.FromNoteID, req.ToNoteID, req.Type, req.Description, req.Strength, metadataJSON)
	if err != nil {
		// Check for foreign key constraint violations
		if strings.Contains(err.Error(), "FOREIGN KEY constraint failed") {
			return nil, fmt.Errorf("invalid note ID: one or both notes do not exist")
		}
		// Check for unique constraint violations
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, fmt.Errorf("connection already exists between these notes with this type")
		}
		// Check for self-connection prevention
		if strings.Contains(err.Error(), "Self-connections are not allowed") {
			return nil, fmt.Errorf("self-connections are not allowed")
		}
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return s.Get(ctx, id)
}

// Get retrieves a connection by ID
func (s *Storage) Get(ctx context.Context, id int64) (*connection.Connection, error) {
	query := `
		SELECT id, from_note_id, to_note_id, type, description, strength, metadata, created_at, updated_at
		FROM connections
		WHERE id = ?
	`

	var conn connection.Connection
	var description sql.NullString
	var metadataJSON string

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&conn.ID,
		&conn.FromNoteID,
		&conn.ToNoteID,
		&conn.Type,
		&description,
		&conn.Strength,
		&metadataJSON,
		&conn.CreatedAt,
		&conn.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("connection not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	if description.Valid {
		conn.Description = &description.String
	}

	if err := json.Unmarshal([]byte(metadataJSON), &conn.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &conn, nil
}

// Update updates an existing connection
func (s *Storage) Update(ctx context.Context, id int64, req connection.UpdateConnectionRequest) (*connection.Connection, error) {
	// Build dynamic update query
	var setClauses []string
	var args []interface{}

	if req.Type != nil {
		if !connection.IsValidConnectionType(*req.Type) {
			return nil, fmt.Errorf("invalid connection type: %s", *req.Type)
		}
		setClauses = append(setClauses, "type = ?")
		args = append(args, *req.Type)
	}

	if req.Description != nil {
		if len(*req.Description) > 500 {
			return nil, fmt.Errorf("description must be 500 characters or less")
		}
		setClauses = append(setClauses, "description = ?")
		args = append(args, *req.Description)
	}

	if req.Strength != nil {
		if *req.Strength < 1 || *req.Strength > 10 {
			return nil, fmt.Errorf("strength must be between 1 and 10, got: %d", *req.Strength)
		}
		setClauses = append(setClauses, "strength = ?")
		args = append(args, *req.Strength)
	}

	if req.Metadata != nil {
		metadataJSON, err := json.Marshal(req.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		setClauses = append(setClauses, "metadata = ?")
		args = append(args, string(metadataJSON))
	}

	if len(setClauses) == 0 {
		return s.Get(ctx, id)
	}

	query := fmt.Sprintf(`
		UPDATE connections
		SET %s, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, strings.Join(setClauses, ", "))

	args = append(args, id)

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		// Check for unique constraint violations
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, fmt.Errorf("connection already exists between these notes with this type")
		}
		return nil, fmt.Errorf("failed to update connection: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("connection not found: %d", id)
	}

	return s.Get(ctx, id)
}

// Delete deletes a connection by ID
func (s *Storage) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM connections WHERE id = ?"

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete connection: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("connection not found: %d", id)
	}

	return nil
}

// List lists connections with pagination and filtering
func (s *Storage) List(ctx context.Context, req connection.ListConnectionsRequest) (*connection.ListConnectionsResponse, error) {
	// Build query
	var whereClauses []string
	var args []interface{}

	if req.FromNoteID != nil {
		whereClauses = append(whereClauses, "from_note_id = ?")
		args = append(args, *req.FromNoteID)
	}

	if req.ToNoteID != nil {
		whereClauses = append(whereClauses, "to_note_id = ?")
		args = append(args, *req.ToNoteID)
	}

	if req.Type != nil {
		whereClauses = append(whereClauses, "type = ?")
		args = append(args, *req.Type)
	}

	if req.Strength != nil {
		whereClauses = append(whereClauses, "strength = ?")
		args = append(args, *req.Strength)
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM connections " + whereClause
	var total int64
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count connections: %w", err)
	}

	// Build order clause
	orderClause := "ORDER BY created_at DESC"
	if req.OrderBy != "" {
		direction := "ASC"
		if req.OrderDir == "desc" {
			direction = "DESC"
		}
		orderClause = fmt.Sprintf("ORDER BY %s %s", req.OrderBy, direction)
	}

	// Get items
	query := fmt.Sprintf(`
		SELECT id, from_note_id, to_note_id, type, description, strength, metadata, created_at, updated_at
		FROM connections
		%s
		%s
		LIMIT ? OFFSET ?
	`, whereClause, orderClause)

	args = append(args, req.Limit, req.Offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query connections: %w", err)
	}
	defer rows.Close()

	var items []connection.Connection
	for rows.Next() {
		var conn connection.Connection
		var description sql.NullString
		var metadataJSON string

		if err := rows.Scan(
			&conn.ID,
			&conn.FromNoteID,
			&conn.ToNoteID,
			&conn.Type,
			&description,
			&conn.Strength,
			&metadataJSON,
			&conn.CreatedAt,
			&conn.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan connection: %w", err)
		}

		if description.Valid {
			conn.Description = &description.String
		}

		if err := json.Unmarshal([]byte(metadataJSON), &conn.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		items = append(items, conn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return &connection.ListConnectionsResponse{
		Items: items,
		Total: total,
	}, nil
}

// GetNoteConnections retrieves all connections for a specific note
func (s *Storage) GetNoteConnections(ctx context.Context, req connection.NoteConnectionsRequest) (*connection.NoteConnectionsResponse, error) {
	var whereClauses []string

	// Base conditions for outgoing and incoming connections
	outgoingWhere := "from_note_id = ?"
	incomingWhere := "to_note_id = ?"
	outgoingArgs := []interface{}{req.NoteID}
	incomingArgs := []interface{}{req.NoteID}

	// Add optional filters
	if req.Type != nil {
		whereClauses = append(whereClauses, "type = ?")
		outgoingArgs = append(outgoingArgs, *req.Type)
		incomingArgs = append(incomingArgs, *req.Type)
	}

	if req.Strength != nil {
		whereClauses = append(whereClauses, "strength = ?")
		outgoingArgs = append(outgoingArgs, *req.Strength)
		incomingArgs = append(incomingArgs, *req.Strength)
	}

	if len(whereClauses) > 0 {
		additionalWhere := strings.Join(whereClauses, " AND ")
		outgoingWhere += " AND " + additionalWhere
		incomingWhere += " AND " + additionalWhere
	}

	// Get outgoing connections
	outgoingQuery := fmt.Sprintf(`
		SELECT id, from_note_id, to_note_id, type, description, strength, metadata, created_at, updated_at
		FROM connections
		WHERE %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, outgoingWhere)

	outgoingArgs = append(outgoingArgs, req.Limit, req.Offset)
	outgoing, err := s.queryConnections(ctx, outgoingQuery, outgoingArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get outgoing connections: %w", err)
	}

	// Get incoming connections
	incomingQuery := fmt.Sprintf(`
		SELECT id, from_note_id, to_note_id, type, description, strength, metadata, created_at, updated_at
		FROM connections
		WHERE %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, incomingWhere)

	incomingArgs = append(incomingArgs, req.Limit, req.Offset)
	incoming, err := s.queryConnections(ctx, incomingQuery, incomingArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to get incoming connections: %w", err)
	}

	// Get total count and type statistics
	totalCount := int64(len(outgoing) + len(incoming))
	typesCount := make(map[string]int64)

	for _, conn := range outgoing {
		typesCount[conn.Type]++
	}
	for _, conn := range incoming {
		typesCount[conn.Type]++
	}

	return &connection.NoteConnectionsResponse{
		NoteID:     req.NoteID,
		Outgoing:   outgoing,
		Incoming:   incoming,
		TotalCount: totalCount,
		TypesCount: typesCount,
	}, nil
}

// GetConnectionsByType retrieves connections filtered by type
func (s *Storage) GetConnectionsByType(ctx context.Context, connectionType string, req connection.ListConnectionsRequest) (*connection.ListConnectionsResponse, error) {
	// Add type filter to the request
	req.Type = &connectionType
	return s.List(ctx, req)
}

// GetBidirectionalConnections retrieves both incoming and outgoing connections for a note
func (s *Storage) GetBidirectionalConnections(ctx context.Context, noteID int64) (*connection.NoteConnectionsResponse, error) {
	req := connection.NoteConnectionsRequest{
		NoteID: noteID,
		Limit:  1000, // Large limit to get all connections
		Offset: 0,
	}
	return s.GetNoteConnections(ctx, req)
}

// GetConnectionStats retrieves statistics about connections
func (s *Storage) GetConnectionStats(ctx context.Context) (*connection.ConnectionStats, error) {
	// Get total connections
	var totalConnections int64
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM connections").Scan(&totalConnections)
	if err != nil {
		return nil, fmt.Errorf("failed to get total connections: %w", err)
	}

	// Get connections by type
	connectionsByType := make(map[string]int64)
	typeRows, err := s.db.QueryContext(ctx, "SELECT type, COUNT(*) FROM connections GROUP BY type")
	if err != nil {
		return nil, fmt.Errorf("failed to get connections by type: %w", err)
	}
	defer typeRows.Close()

	for typeRows.Next() {
		var connType string
		var count int64
		if err := typeRows.Scan(&connType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan type count: %w", err)
		}
		connectionsByType[connType] = count
	}

	// Get connections by strength
	connectionsByStrength := make(map[int]int64)
	strengthRows, err := s.db.QueryContext(ctx, "SELECT strength, COUNT(*) FROM connections GROUP BY strength")
	if err != nil {
		return nil, fmt.Errorf("failed to get connections by strength: %w", err)
	}
	defer strengthRows.Close()

	for strengthRows.Next() {
		var strength int
		var count int64
		if err := strengthRows.Scan(&strength, &count); err != nil {
			return nil, fmt.Errorf("failed to scan strength count: %w", err)
		}
		connectionsByStrength[strength] = count
	}

	// Get most connected notes
	mostConnectedNotes := []connection.NoteConnection{}
	noteRows, err := s.db.QueryContext(ctx, `
		SELECT 
			note_id,
			SUM(incoming_count) as incoming_count,
			SUM(outgoing_count) as outgoing_count,
			SUM(incoming_count + outgoing_count) as total_count
		FROM (
			SELECT from_note_id as note_id, COUNT(*) as outgoing_count, 0 as incoming_count
			FROM connections 
			GROUP BY from_note_id
			UNION ALL
			SELECT to_note_id as note_id, 0 as outgoing_count, COUNT(*) as incoming_count
			FROM connections 
			GROUP BY to_note_id
		) 
		GROUP BY note_id 
		ORDER BY total_count DESC 
		LIMIT 10
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get most connected notes: %w", err)
	}
	defer noteRows.Close()

	for noteRows.Next() {
		var noteConn connection.NoteConnection
		if err := noteRows.Scan(&noteConn.NoteID, &noteConn.IncomingCount, &noteConn.OutgoingCount, &noteConn.TotalCount); err != nil {
			return nil, fmt.Errorf("failed to scan note connection: %w", err)
		}
		mostConnectedNotes = append(mostConnectedNotes, noteConn)
	}

	return &connection.ConnectionStats{
		TotalConnections:      totalConnections,
		ConnectionsByType:     connectionsByType,
		ConnectionsByStrength: connectionsByStrength,
		MostConnectedNotes:    mostConnectedNotes,
	}, nil
}

// FindConnectionPaths finds paths between two notes (basic implementation)
func (s *Storage) FindConnectionPaths(ctx context.Context, fromNoteID, toNoteID int64, maxDepth int) ([]connection.ConnectionPath, error) {
	// This is a basic implementation that finds direct connections
	// A more sophisticated implementation would use graph traversal algorithms
	
	query := `
		SELECT id, from_note_id, to_note_id, type, description, strength, metadata, created_at, updated_at
		FROM connections
		WHERE from_note_id = ? AND to_note_id = ?
	`

	connections, err := s.queryConnections(ctx, query, fromNoteID, toNoteID)
	if err != nil {
		return nil, fmt.Errorf("failed to find direct connections: %w", err)
	}

	var paths []connection.ConnectionPath
	for _, conn := range connections {
		path := connection.ConnectionPath{
			FromNoteID: fromNoteID,
			ToNoteID:   toNoteID,
			Path:       []connection.Connection{conn},
			Length:     1,
			Strength:   conn.Strength,
		}
		paths = append(paths, path)
	}

	return paths, nil
}

// queryConnections is a helper method to query connections and scan results
func (s *Storage) queryConnections(ctx context.Context, query string, args ...interface{}) ([]connection.Connection, error) {
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []connection.Connection
	for rows.Next() {
		var conn connection.Connection
		var description sql.NullString
		var metadataJSON string

		if err := rows.Scan(
			&conn.ID,
			&conn.FromNoteID,
			&conn.ToNoteID,
			&conn.Type,
			&description,
			&conn.Strength,
			&metadataJSON,
			&conn.CreatedAt,
			&conn.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if description.Valid {
			conn.Description = &description.String
		}

		if err := json.Unmarshal([]byte(metadataJSON), &conn.Metadata); err != nil {
			return nil, err
		}

		connections = append(connections, conn)
	}

	return connections, rows.Err()
}