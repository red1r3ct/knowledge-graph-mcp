package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/red1r3ct/knowledge-graph-mcp/internal/note"
)

// NewGetHandler creates a new handler for getting a note by ID
func NewGetHandler(storage note.Storage) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments, ok := req.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments format")
		}

		idStr, ok := arguments["id"].(string)
		if !ok || idStr == "" {
			return nil, fmt.Errorf("id is required")
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid id format: %w", err)
		}

		n, err := storage.Get(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to get note: %w", err)
		}

		if n == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("Note with ID %d not found", id),
					},
				},
			}, nil
		}

		result := map[string]interface{}{
			"id":         n.ID,
			"title":      n.Title,
			"content":    n.Content,
			"type":       n.Type,
			"tags":       n.Tags,
			"metadata":   n.Metadata,
			"created_at": n.CreatedAt,
			"updated_at": n.UpdatedAt,
		}

		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: string(jsonData),
				},
			},
		}, nil
	}
}