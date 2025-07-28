package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/red1r3ct/knowledge-graph-mcp/internal/connection"
)

// NewListHandler creates a new handler for listing connections with filtering
func NewListHandler(storage connection.Storage) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments, ok := req.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments format")
		}

		listReq := connection.ListConnectionsRequest{
			Limit:    100, // Default limit
			Offset:   0,   // Default offset
			OrderBy:  "id",
			OrderDir: "asc",
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
			listReq.Limit = limit
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
			listReq.Offset = offset
		}

		// Parse optional from_note_id filter
		if fromNoteIDRaw, ok := arguments["from_note_id"]; ok {
			fromNoteID, err := parseInt64(fromNoteIDRaw)
			if err != nil {
				return nil, fmt.Errorf("invalid from_note_id: %w", err)
			}
			listReq.FromNoteID = &fromNoteID
		}

		// Parse optional to_note_id filter
		if toNoteIDRaw, ok := arguments["to_note_id"]; ok {
			toNoteID, err := parseInt64(toNoteIDRaw)
			if err != nil {
				return nil, fmt.Errorf("invalid to_note_id: %w", err)
			}
			listReq.ToNoteID = &toNoteID
		}

		// Parse optional type filter
		if connectionType, ok := arguments["type"].(string); ok && connectionType != "" {
			// Validate connection type
			if !connection.IsValidConnectionType(connectionType) {
				return nil, fmt.Errorf("invalid connection type: %s. Valid types are: %v", connectionType, connection.ValidConnectionTypes())
			}
			listReq.Type = &connectionType
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
			listReq.Strength = &strength
		}

		// Parse optional order_by
		if orderBy, ok := arguments["order_by"].(string); ok && orderBy != "" {
			validOrderBy := []string{"id", "created_at", "updated_at", "strength", "type"}
			isValid := false
			for _, valid := range validOrderBy {
				if orderBy == valid {
					isValid = true
					break
				}
			}
			if !isValid {
				return nil, fmt.Errorf("invalid order_by: %s. Valid values are: %v", orderBy, validOrderBy)
			}
			listReq.OrderBy = orderBy
		}

		// Parse optional order_dir
		if orderDir, ok := arguments["order_dir"].(string); ok && orderDir != "" {
			if orderDir != "asc" && orderDir != "desc" {
				return nil, fmt.Errorf("invalid order_dir: %s. Valid values are: asc, desc", orderDir)
			}
			listReq.OrderDir = orderDir
		}

		response, err := storage.List(ctx, listReq)
		if err != nil {
			return nil, fmt.Errorf("failed to list connections: %w", err)
		}

		result := map[string]interface{}{
			"items": response.Items,
			"total": response.Total,
			"limit": listReq.Limit,
			"offset": listReq.Offset,
		}

		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Found %d connections (showing %d-%d of %d total):\n\n%s", 
						len(response.Items), 
						listReq.Offset+1, 
						listReq.Offset+len(response.Items), 
						response.Total, 
						string(jsonData)),
				},
			},
		}, nil
	}
}