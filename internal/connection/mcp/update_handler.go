package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/red1r3ct/knowledge-graph-mcp/internal/connection"
)

// NewUpdateHandler creates a new handler for updating connections
func NewUpdateHandler(storage connection.Storage) server.ToolHandlerFunc {
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

		updateReq := connection.UpdateConnectionRequest{}

		// Parse optional type
		if connectionType, ok := arguments["type"].(string); ok && connectionType != "" {
			// Validate connection type
			if !connection.IsValidConnectionType(connectionType) {
				return nil, fmt.Errorf("invalid connection type: %s. Valid types are: %v", connectionType, connection.ValidConnectionTypes())
			}
			updateReq.Type = &connectionType
		}

		// Parse optional description
		if desc, ok := arguments["description"].(string); ok {
			updateReq.Description = &desc
		}

		// Parse optional strength
		if strengthRaw, ok := arguments["strength"]; ok {
			strength, err := parseInt(strengthRaw)
			if err != nil {
				return nil, fmt.Errorf("invalid strength: %w", err)
			}
			// Validate strength range
			if strength < 1 || strength > 10 {
				return nil, fmt.Errorf("strength must be between 1 and 10, got: %d", strength)
			}
			updateReq.Strength = &strength
		}

		// Parse optional metadata
		if metadataRaw, ok := arguments["metadata"].(map[string]interface{}); ok {
			updateReq.Metadata = metadataRaw
		}

		conn, err := storage.Update(ctx, id, updateReq)
		if err != nil {
			return nil, fmt.Errorf("failed to update connection: %w", err)
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
					Text: fmt.Sprintf("Successfully updated connection with ID: %d\n\n%s", conn.ID, string(jsonData)),
				},
			},
		}, nil
	}
}