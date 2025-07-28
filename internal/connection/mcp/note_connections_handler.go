package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/red1r3ct/knowledge-graph-mcp/internal/connection"
)

// NewNoteConnectionsHandler creates a new handler for getting all connections of a note
func NewNoteConnectionsHandler(storage connection.Storage) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments, ok := req.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments format")
		}

		// Parse note_id
		noteIDRaw, ok := arguments["note_id"]
		if !ok {
			return nil, fmt.Errorf("note_id is required")
		}

		noteID, err := parseInt64(noteIDRaw)
		if err != nil {
			return nil, fmt.Errorf("invalid note_id: %w", err)
		}

		if noteID <= 0 {
			return nil, fmt.Errorf("note_id must be a positive integer")
		}

		noteConnReq := connection.NoteConnectionsRequest{
			NoteID: noteID,
			Limit:  100, // Default limit
			Offset: 0,   // Default offset
		}

		// Parse optional type filter
		if connectionType, ok := arguments["type"].(string); ok && connectionType != "" {
			// Validate connection type
			if !connection.IsValidConnectionType(connectionType) {
				return nil, fmt.Errorf("invalid connection type: %s. Valid types are: %v", connectionType, connection.ValidConnectionTypes())
			}
			noteConnReq.Type = &connectionType
		}

		// Parse optional strength filter
		if strengthRaw, ok := arguments["strength"]; ok {
			strength, err := parseInt(strengthRaw)
			if err != nil {
				return nil, fmt.Errorf("invalid strength: %w", err)
			}
			if strength < 1 || strength > 10 {
				return nil, fmt.Errorf("strength must be between 1 and 10, got: %d", strength)
			}
			noteConnReq.Strength = &strength
		}

		// Parse optional limit
		if limitRaw, ok := arguments["limit"]; ok {
			limit, err := parseInt(limitRaw)
			if err != nil {
				return nil, fmt.Errorf("invalid limit: %w", err)
			}
			if limit < 1 || limit > 1000 {
				return nil, fmt.Errorf("limit must be between 1 and 1000, got: %d", limit)
			}
			noteConnReq.Limit = limit
		}

		// Parse optional offset
		if offsetRaw, ok := arguments["offset"]; ok {
			offset, err := parseInt(offsetRaw)
			if err != nil {
				return nil, fmt.Errorf("invalid offset: %w", err)
			}
			if offset < 0 {
				return nil, fmt.Errorf("offset must be non-negative, got: %d", offset)
			}
			noteConnReq.Offset = offset
		}

		response, err := storage.GetNoteConnections(ctx, noteConnReq)
		if err != nil {
			return nil, fmt.Errorf("failed to get note connections: %w", err)
		}

		result := map[string]interface{}{
			"note_id":      response.NoteID,
			"outgoing":     response.Outgoing,
			"incoming":     response.Incoming,
			"total_count":  response.TotalCount,
			"types_count":  response.TypesCount,
		}

		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}

		outgoingCount := len(response.Outgoing)
		incomingCount := len(response.Incoming)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Found connections for note %d:\n- %d outgoing connections\n- %d incoming connections\n- %d total connections\n\n%s", 
						noteID, 
						outgoingCount, 
						incomingCount, 
						response.TotalCount, 
						string(jsonData)),
				},
			},
		}, nil
	}
}