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

// NewUpdateHandler creates a new handler for updating knowledge base entries
func NewUpdateHandler(storage knowledgebase.Storage) server.ToolHandlerFunc {
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

		updateReq := knowledgebase.UpdateRequest{}

		if name, ok := arguments["name"].(string); ok && name != "" {
			updateReq.Name = &name
		}

		if desc, ok := arguments["description"].(string); ok {
			updateReq.Description = &desc
		}

		if tagsRaw, ok := arguments["tags"].([]interface{}); ok {
			var tags []string
			for _, tag := range tagsRaw {
				if tagStr, ok := tag.(string); ok {
					tags = append(tags, tagStr)
				}
			}
			updateReq.Tags = tags
		}

		kb, err := storage.Update(ctx, id, updateReq)
		if err != nil {
			return nil, fmt.Errorf("failed to update knowledge base: %w", err)
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
					Text: fmt.Sprintf("Successfully updated knowledge base entry with ID: %d\n\n%s", kb.ID, string(jsonData)),
				},
			},
		}, nil
	}
}
