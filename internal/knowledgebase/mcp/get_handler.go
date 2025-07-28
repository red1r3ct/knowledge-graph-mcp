package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/red1r3ct/knowledge-graph-mcp/internal/knowledgebase"
)

// NewGetHandler creates a new handler for getting a knowledge base entry by ID
func NewGetHandler(storage knowledgebase.Storage) server.ToolHandlerFunc {
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

		kb, err := storage.Get(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to get knowledge base: %w", err)
		}

		if kb == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("Knowledge base entry with ID %d not found", id),
					},
				},
			}, nil
		}

		result := map[string]interface{}{
			"id":          kb.ID,
			"name":        kb.Name,
			"description": kb.Description,
			"tags":        kb.Tags,
			"created_at":  kb.CreatedAt,
			"updated_at":  kb.UpdatedAt,
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
