package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/red1r3ct/knowledge-graph-mcp/internal/note"
)

// NewListHandler creates a new handler for listing notes
func NewListHandler(storage note.Storage) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments, ok := req.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments format")
		}

		listReq := note.ListNotesRequest{}

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

		// Parse type
		if noteType, ok := arguments["type"].(string); ok {
			listReq.Type = noteType
		}

		// Parse order_by
		if orderBy, ok := arguments["order_by"].(string); ok {
			listReq.OrderBy = orderBy
		}

		// Parse order_dir
		if orderDir, ok := arguments["order_dir"].(string); ok {
			listReq.OrderDir = orderDir
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
			return nil, fmt.Errorf("failed to list notes: %w", err)
		}

		if len(response.Items) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: "No notes found",
					},
				},
			}, nil
		}

		var results []map[string]interface{}
		for _, n := range response.Items {
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
			results = append(results, result)
		}

		summary := map[string]interface{}{
			"total": response.Total,
			"count": len(response.Items),
			"items": results,
		}

		jsonData, err := json.MarshalIndent(summary, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal results: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Found %d notes (total: %d):\n\n%s", len(response.Items), response.Total, string(jsonData)),
				},
			},
		}, nil
	}
}