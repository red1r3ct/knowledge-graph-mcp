package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/red1r3ct/knowledge-graph-mcp/internal/connection"
)

// NewDeleteHandler creates a new handler for deleting connections
func NewDeleteHandler(storage connection.Storage) server.ToolHandlerFunc {
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

		err = storage.Delete(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to delete connection: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Successfully deleted connection with ID: %d", id),
				},
			},
		}, nil
	}
}