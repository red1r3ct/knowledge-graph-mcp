package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/red1r3ct/knowledge-graph-mcp/internal/knowledgebase"
)

// NewListHandler creates a new handler for listing knowledge base entries
func NewListHandler(storage knowledgebase.Storage) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments, ok := req.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments format")
		}

		listReq := knowledgebase.ListRequest{}

		// Parse limit
		if limit, ok := arguments["limit"].(float64); ok {
			listReq.Limit = int(limit)
		} else {
			listReq.Limit = 100 // default
		}

		// Parse offset
		if offset, ok := arguments["offset"].(float64); ok {
			listReq.Offset = int(offset)
		} else {
			listReq.Offset = 0 // default
		}

		// Parse search
		if search, ok := arguments["search"].(string); ok {
			listReq.Search = search
		}

		// Parse tags
		if tagsRaw, ok := arguments["tags"].([]interface{}); ok {
			var tags []string
			for _, tag := range tagsRaw {
				if tagStr, ok := tag.(string); ok {
					tags = append(tags, tagStr)
				}
			}
			listReq.Tags = tags
		}

		response, err := storage.List(ctx, listReq)
		if err != nil {
			return nil, fmt.Errorf("failed to list knowledge bases: %w", err)
		}

		if len(response.Items) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: "No knowledge base entries found",
					},
				},
			}, nil
		}

		var results []map[string]interface{}
		for _, kb := range response.Items {
			result := map[string]interface{}{
				"id":          kb.ID,
				"name":        kb.Name,
				"description": kb.Description,
				"tags":        kb.Tags,
				"created_at":  kb.CreatedAt,
				"updated_at":  kb.UpdatedAt,
			}
			results = append(results, result)
		}

		jsonData, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal results: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Found %d knowledge base entries:\n\n%s", len(response.Items), string(jsonData)),
				},
			},
		}, nil
	}
}
