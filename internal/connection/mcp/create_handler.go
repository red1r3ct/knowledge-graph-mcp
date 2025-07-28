package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/red1r3ct/knowledge-graph-mcp/internal/connection"
)

// NewCreateHandler creates a new handler for creating connections
func NewCreateHandler(storage connection.Storage) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments, ok := req.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments format")
		}

		// Parse from_note_id
		fromNoteIDRaw, ok := arguments["from_note_id"]
		if !ok {
			return nil, fmt.Errorf("from_note_id is required")
		}
		fromNoteID, err := parseInt64(fromNoteIDRaw)
		if err != nil {
			return nil, fmt.Errorf("invalid from_note_id: %w", err)
		}

		// Parse to_note_id
		toNoteIDRaw, ok := arguments["to_note_id"]
		if !ok {
			return nil, fmt.Errorf("to_note_id is required")
		}
		toNoteID, err := parseInt64(toNoteIDRaw)
		if err != nil {
			return nil, fmt.Errorf("invalid to_note_id: %w", err)
		}

		// Validate that from_note_id != to_note_id
		if fromNoteID == toNoteID {
			return nil, fmt.Errorf("from_note_id and to_note_id cannot be the same")
		}

		// Parse type
		connectionType, ok := arguments["type"].(string)
		if !ok || connectionType == "" {
			return nil, fmt.Errorf("type is required")
		}

		// Validate connection type
		if !connection.IsValidConnectionType(connectionType) {
			return nil, fmt.Errorf("invalid connection type: %s. Valid types are: %v", connectionType, connection.ValidConnectionTypes())
		}

		// Parse strength (required, default to 5 if not provided)
		strength := 5
		if strengthRaw, ok := arguments["strength"]; ok {
			strengthInt, err := parseInt(strengthRaw)
			if err != nil {
				return nil, fmt.Errorf("invalid strength: %w", err)
			}
			strength = strengthInt
		}

		// Validate strength range
		if strength < 1 || strength > 10 {
			return nil, fmt.Errorf("strength must be between 1 and 10, got: %d", strength)
		}

		// Parse optional description
		var description *string
		if desc, ok := arguments["description"].(string); ok && desc != "" {
			description = &desc
		}

		// Parse optional metadata
		var metadata map[string]interface{}
		if metadataRaw, ok := arguments["metadata"].(map[string]interface{}); ok {
			metadata = metadataRaw
		}

		createReq := connection.CreateConnectionRequest{
			FromNoteID:  fromNoteID,
			ToNoteID:    toNoteID,
			Type:        connectionType,
			Description: description,
			Strength:    strength,
			Metadata:    metadata,
		}

		conn, err := storage.Create(ctx, createReq)
		if err != nil {
			return nil, fmt.Errorf("failed to create connection: %w", err)
		}

		result := map[string]interface{}{
			"id":           conn.ID,
			"from_note_id": conn.FromNoteID,
			"to_note_id":   conn.ToNoteID,
			"type":         conn.Type,
			"description":  conn.Description,
			"strength":     conn.Strength,
			"metadata":     conn.Metadata,
			"created_at":   conn.CreatedAt,
			"updated_at":   conn.UpdatedAt,
		}

		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Successfully created connection with ID: %d\n\n%s", conn.ID, string(jsonData)),
				},
			},
		}, nil
	}
}

// parseInt64 parses various types to int64
func parseInt64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", value)
	}
}

// parseInt parses various types to int
func parseInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		i, err := strconv.Atoi(v)
		return i, err
	default:
		return 0, fmt.Errorf("cannot convert %T to int", value)
	}
}