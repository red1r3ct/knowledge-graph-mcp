package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"

	"github.com/red1r3ct/knowledge-graph-mcp/internal/knowledgebase"
)

// Storage implements the knowledgebase.Storage interface using SQLite
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

// Create creates a new knowledge base
func (s *Storage) Create(ctx context.Context, req knowledgebase.CreateRequest) (*knowledgebase.KnowledgeBase, error) {
	tagsJSON, err := json.Marshal(req.Tags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tags: %w", err)
	}

	query := `
		INSERT INTO knowledge_base (name, description, tags)
		VALUES (?, ?, ?)
	`

	result, err := s.db.ExecContext(ctx, query, req.Name, req.Description, string(tagsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create knowledge base: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return s.Get(ctx, id)
}

// Get retrieves a knowledge base by ID
func (s *Storage) Get(ctx context.Context, id int64) (*knowledgebase.KnowledgeBase, error) {
	query := `
		SELECT id, name, description, tags, created_at, updated_at
		FROM knowledge_base
		WHERE id = ?
	`

	var kb knowledgebase.KnowledgeBase
	var description sql.NullString
	var tagsJSON string

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&kb.ID,
		&kb.Name,
		&description,
		&tagsJSON,
		&kb.CreatedAt,
		&kb.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("knowledge base not found: %d", id)
		}
		return nil, fmt.Errorf("failed to get knowledge base: %w", err)
	}

	if description.Valid {
		kb.Description = &description.String
	}

	if err := json.Unmarshal([]byte(tagsJSON), &kb.Tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
	}

	return &kb, nil
}

// Update updates an existing knowledge base
func (s *Storage) Update(ctx context.Context, id int64, req knowledgebase.UpdateRequest) (*knowledgebase.KnowledgeBase, error) {
	// Build dynamic update query
	var setClauses []string
	var args []interface{}

	if req.Name != nil {
		setClauses = append(setClauses, "name = ?")
		args = append(args, *req.Name)
	}

	if req.Description != nil {
		setClauses = append(setClauses, "description = ?")
		args = append(args, *req.Description)
	}

	if req.Tags != nil {
		tagsJSON, err := json.Marshal(req.Tags)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tags: %w", err)
		}
		setClauses = append(setClauses, "tags = ?")
		args = append(args, string(tagsJSON))
	}

	if len(setClauses) == 0 {
		return s.Get(ctx, id)
	}

	query := fmt.Sprintf(`
		UPDATE knowledge_base
		SET %s, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, strings.Join(setClauses, ", "))

	args = append(args, id)

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update knowledge base: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("knowledge base not found: %d", id)
	}

	return s.Get(ctx, id)
}

// Delete deletes a knowledge base by ID
func (s *Storage) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM knowledge_base WHERE id = ?"

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete knowledge base: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("knowledge base not found: %d", id)
	}

	return nil
}

// List lists knowledge bases with pagination and filtering
func (s *Storage) List(ctx context.Context, req knowledgebase.ListRequest) (*knowledgebase.ListResponse, error) {
	// Build query
	var whereClauses []string
	var args []interface{}

	if req.Search != "" {
		whereClauses = append(whereClauses, "(name LIKE ? OR description LIKE ?)")
		searchPattern := "%" + req.Search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	if len(req.Tags) > 0 {
		for _, tag := range req.Tags {
			whereClauses = append(whereClauses, "tags LIKE ?")
			args = append(args, "%\""+tag+"\"%")
		}
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM knowledge_base " + whereClause
	var total int64
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count knowledge bases: %w", err)
	}

	// Get items
	query := fmt.Sprintf(`
		SELECT id, name, description, tags, created_at, updated_at
		FROM knowledge_base
		%s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	args = append(args, req.Limit, req.Offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query knowledge bases: %w", err)
	}
	defer rows.Close()

	var items []knowledgebase.KnowledgeBase
	for rows.Next() {
		var kb knowledgebase.KnowledgeBase
		var description sql.NullString
		var tagsJSON string

		if err := rows.Scan(
			&kb.ID,
			&kb.Name,
			&description,
			&tagsJSON,
			&kb.CreatedAt,
			&kb.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan knowledge base: %w", err)
		}

		if description.Valid {
			kb.Description = &description.String
		}

		if err := json.Unmarshal([]byte(tagsJSON), &kb.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}

		items = append(items, kb)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return &knowledgebase.ListResponse{
		Items: items,
		Total: total,
	}, nil
}
