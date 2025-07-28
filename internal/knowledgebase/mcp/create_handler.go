package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/red1r3ct/knowledge-graph-mcp/internal/knowledgebase"
)

// NewCreateHandler creates a new handler for creating knowledge base entries
func NewCreateHandler(storage knowledgebase.Storage) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments, ok := req.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments format")
		}

		name, ok := arguments["name"].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("name is required")
		}

		description := ""
		if desc, ok := arguments["description"].(string); ok {
			description = desc
		}

		var tags []string
		if tagsRaw, ok := arguments["tags"].([]interface{}); ok {
			for _, tag := range tagsRaw {
				if tagStr, ok := tag.(string); ok {
					tags = append(tags, tagStr)
				}
			}
		}

		createReq := knowledgebase.CreateRequest{
			Name:        name,
			Description: &description,
			Tags:        tags,
		}

		kb, err := storage.Create(ctx, createReq)
		if err != nil {
			return nil, fmt.Errorf("failed to create knowledge base: %w", err)
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
					Text: fmt.Sprintf("Successfully created knowledge base entry with ID: %d\n\n%s", kb.ID, string(jsonData)),
				},
			},
		}, nil
	}
}
