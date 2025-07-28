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

// NewUpdateHandler creates a new handler for updating notes
func NewUpdateHandler(storage note.Storage) server.ToolHandlerFunc {
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

		updateReq := note.UpdateNoteRequest{}

		if title, ok := arguments["title"].(string); ok && title != "" {
			updateReq.Title = &title
		}

		if content, ok := arguments["content"].(string); ok {
			updateReq.Content = &content
		}

		if noteType, ok := arguments["type"].(string); ok && noteType != "" {
			updateReq.Type = &noteType
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

		if metadataRaw, ok := arguments["metadata"].(map[string]interface{}); ok {
			updateReq.Metadata = metadataRaw
		}

		n, err := storage.Update(ctx, id, updateReq)
		if err != nil {
			return nil, fmt.Errorf("failed to update note: %w", err)
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
					Text: fmt.Sprintf("Successfully updated note with ID: %d\n\n%s", n.ID, string(jsonData)),
				},
			},
		}, nil
	}
}