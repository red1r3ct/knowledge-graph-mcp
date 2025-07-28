package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/red1r3ct/knowledge-graph-mcp/internal/connection"
)

// NewGetHandler creates a new handler for getting connections by ID
func NewGetHandler(storage connection.Storage) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments, ok := req.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments format")
		}

		// Parse ID
		idRaw, ok := arguments["id"]
		if !ok {
			return nil, fmt.Errorf("id is required")
		}

		id, err := parseInt64(idRaw)
		if err != nil {
			return nil, fmt.Errorf("invalid id: %w", err)
		}

		if id <= 0 {
			return nil, fmt.Errorf("id must be a positive integer")
		}

		conn, err := storage.Get(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to get connection: %w", err)
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
					Text: fmt.Sprintf("Connection found:\n\n%s", string(jsonData)),
				},
			},
		}, nil
	}
}