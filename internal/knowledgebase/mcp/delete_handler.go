package mcp

import (
	"context"
	"fmt"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/red1r3ct/knowledge-graph-mcp/internal/knowledgebase"
)

// NewDeleteHandler creates a new handler for deleting knowledge base entries
func NewDeleteHandler(storage knowledgebase.Storage) server.ToolHandlerFunc {
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

		err = storage.Delete(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to delete knowledge base: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Successfully deleted knowledge base entry with ID: %d", id),
				},
			},
		}, nil
	}
}
