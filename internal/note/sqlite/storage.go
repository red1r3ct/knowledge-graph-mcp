package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"

	"github.com/red1r3ct/knowledge-graph-mcp/internal/note"
)

// Storage implements the note.Storage interface using SQLite
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

// Create creates a new note
func (s *Storage) Create(ctx context.Context, req note.CreateNoteRequest) (*note.Note, error) {
	var tagsJSON string
	var metadataJSON string

	if req.Tags != nil {
		tagsBytes, err := json.Marshal(req.Tags)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tags: %w", err)
		}
		tagsJSON = string(tagsBytes)
	} else {
		tagsJSON = "[]"
	}

	if req.Metadata != nil {
		metadataBytes, err := json.Marshal(req.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataJSON = string(metadataBytes)
	} else {
		metadataJSON = "null"
	}

	query := `
		INSERT INTO notes (title, content, type, tags, metadata)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := s.db.ExecContext(ctx, query, req.Title, req.Content, req.Type, tagsJSON, metadataJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return s.Get(ctx, id)
}

// Get retrieves a note by ID
func (s *Storage) Get(ctx context.Context, id int64) (*note.Note, error) {
	query := `
		SELECT id, title, content, type, tags, metadata, created_at, updated_at
		FROM notes
		WHERE id = ?
	`

	var n note.Note
	var tagsJSON string
	var metadataJSON string

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&n.ID,
		&n.Title,
		&n.Content,
		&n.Type,
		&tagsJSON,
		&metadataJSON,
		&n.CreatedAt,
		&n.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("note not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	if err := json.Unmarshal([]byte(tagsJSON), &n.Tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
	}

	if metadataJSON != "null" {
		if err := json.Unmarshal([]byte(metadataJSON), &n.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &n, nil
}

// Update updates an existing note
func (s *Storage) Update(ctx context.Context, id int64, req note.UpdateNoteRequest) (*note.Note, error) {
	// Build dynamic update query
	var setClauses []string
	var args []interface{}

	if req.Title != nil {
		setClauses = append(setClauses, "title = ?")
		args = append(args, *req.Title)
	}

	if req.Content != nil {
		setClauses = append(setClauses, "content = ?")
		args = append(args, *req.Content)
	}

	if req.Type != nil {
		setClauses = append(setClauses, "type = ?")
		args = append(args, *req.Type)
	}

	if req.Tags != nil {
		tagsJSON, err := json.Marshal(req.Tags)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tags: %w", err)
		}
		setClauses = append(setClauses, "tags = ?")
		args = append(args, string(tagsJSON))
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
		UPDATE notes
		SET %s, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, strings.Join(setClauses, ", "))

	args = append(args, id)

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update note: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("note not found: %d", id)
	}

	return s.Get(ctx, id)
}

// Delete deletes a note by ID
func (s *Storage) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM notes WHERE id = ?"

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("note not found: %d", id)
	}

	return nil
}

// List lists notes with pagination and filtering
func (s *Storage) List(ctx context.Context, req note.ListNotesRequest) (*note.ListNotesResponse, error) {
	// Build query
	var whereClauses []string
	var args []interface{}

	if req.Search != "" {
		// Use FTS for full-text search
		whereClauses = append(whereClauses, "notes.id IN (SELECT rowid FROM notes_fts WHERE notes_fts MATCH ?)")
		args = append(args, req.Search)
	}

	if len(req.Tags) > 0 {
		for _, tag := range req.Tags {
			whereClauses = append(whereClauses, "tags LIKE ?")
			args = append(args, "%\""+tag+"\"%")
		}
	}

	if req.Type != "" {
		whereClauses = append(whereClauses, "type = ?")
		args = append(args, req.Type)
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM notes " + whereClause
	var total int64
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count notes: %w", err)
	}

	// Build order clause
	orderBy := "created_at"
	orderDir := "DESC"
	if req.OrderBy != "" {
		orderBy = req.OrderBy
	}
	if req.OrderDir != "" {
		orderDir = strings.ToUpper(req.OrderDir)
	}

	// Get items
	query := fmt.Sprintf(`
		SELECT id, title, content, type, tags, metadata, created_at, updated_at
		FROM notes
		%s
		ORDER BY %s %s
		LIMIT ? OFFSET ?
	`, whereClause, orderBy, orderDir)

	args = append(args, req.Limit, req.Offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query notes: %w", err)
	}
	defer rows.Close()

	var items []note.Note
	for rows.Next() {
		var n note.Note
		var tagsJSON string
		var metadataJSON string

		if err := rows.Scan(
			&n.ID,
			&n.Title,
			&n.Content,
			&n.Type,
			&tagsJSON,
			&metadataJSON,
			&n.CreatedAt,
			&n.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}

		if err := json.Unmarshal([]byte(tagsJSON), &n.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}

		if metadataJSON != "null" {
			if err := json.Unmarshal([]byte(metadataJSON), &n.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		items = append(items, n)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return &note.ListNotesResponse{
		Items: items,
		Total: total,
	}, nil
}