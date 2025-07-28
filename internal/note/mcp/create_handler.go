package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/red1r3ct/knowledge-graph-mcp/internal/note"
)

// NewCreateHandler creates a new handler for creating notes
func NewCreateHandler(storage note.Storage) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments, ok := req.Params.Arguments.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid arguments format")
		}

		title, ok := arguments["title"].(string)
		if !ok || title == "" {
			return nil, fmt.Errorf("title is required")
		}

		content, ok := arguments["content"].(string)
		if !ok || content == "" {
			return nil, fmt.Errorf("content is required")
		}

		noteType := "text"
		if typ, ok := arguments["type"].(string); ok && typ != "" {
			noteType = typ
		}

		var tags []string
		if tagsRaw, ok := arguments["tags"].([]interface{}); ok {
			for _, tag := range tagsRaw {
				if tagStr, ok := tag.(string); ok {
					tags = append(tags, tagStr)
				}
			}
		}

		var metadata map[string]interface{}
		if metadataRaw, ok := arguments["metadata"].(map[string]interface{}); ok {
			metadata = metadataRaw
		}

		createReq := note.CreateNoteRequest{
			Title:    title,
			Content:  content,
			Type:     noteType,
			Tags:     tags,
			Metadata: metadata,
		}

		n, err := storage.Create(ctx, createReq)
		if err != nil {
			return nil, fmt.Errorf("failed to create note: %w", err)
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
					Text: fmt.Sprintf("Successfully created note with ID: %d\n\n%s", n.ID, string(jsonData)),
				},
			},
		}, nil
	}
}