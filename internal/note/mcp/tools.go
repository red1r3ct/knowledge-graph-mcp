package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/red1r3ct/knowledge-graph-mcp/internal/note"
)

// RegisterTools registers all note MCP tools with the server
func RegisterTools(s *server.MCPServer, storage note.Storage) error {
	tools := []struct {
		name        string
		description string
		handler     server.ToolHandlerFunc
		schema      mcp.ToolInputSchema
	}{
		{
			name:        "create_note",
			description: "Create a new note",
			handler:     NewCreateHandler(storage),
			schema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Title of the note",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Content of the note",
					},
					"type": map[string]interface{}{
						"type":        "string",
						"description": "Type of the note (text, markdown, code, link, image)",
						"enum":        []string{"text", "markdown", "code", "link", "image"},
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"description": "Tags associated with the note",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"metadata": map[string]interface{}{
						"type":        "object",
						"description": "Additional metadata for the note",
					},
				},
				Required: []string{"title", "content"},
			},
		},
		{
			name:        "get_note",
			description: "Get a note by ID",
			handler:     NewGetHandler(storage),
			schema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Unique identifier of the note",
					},
				},
				Required: []string{"id"},
			},
		},
		{
			name:        "update_note",
			description: "Update an existing note",
			handler:     NewUpdateHandler(storage),
			schema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Unique identifier of the note",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Updated title of the note",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Updated content of the note",
					},
					"type": map[string]interface{}{
						"type":        "string",
						"description": "Updated type of the note (text, markdown, code, link, image)",
						"enum":        []string{"text", "markdown", "code", "link", "image"},
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"description": "Updated tags associated with the note",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"metadata": map[string]interface{}{
						"type":        "object",
						"description": "Updated metadata for the note",
					},
				},
				Required: []string{"id"},
			},
		},
		{
			name:        "delete_note",
			description: "Delete a note by ID",
			handler:     NewDeleteHandler(storage),
			schema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Unique identifier of the note to delete",
					},
				},
				Required: []string{"id"},
			},
		},
		{
			name:        "list_notes",
			description: "List all notes with optional filtering and pagination",
			handler:     NewListHandler(storage),
			schema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of notes to return (default: 100)",
						"minimum":     1,
						"maximum":     1000,
					},
					"offset": map[string]interface{}{
						"type":        "integer",
						"description": "Number of notes to skip (default: 0)",
						"minimum":     0,
					},
					"search": map[string]interface{}{
						"type":        "string",
						"description": "Search term to filter notes by title or content",
					},
					"type": map[string]interface{}{
						"type":        "string",
						"description": "Filter by note type",
						"enum":        []string{"text", "markdown", "code", "link", "image"},
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"description": "Filter by tags (returns notes that have any of the specified tags)",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"order_by": map[string]interface{}{
						"type":        "string",
						"description": "Field to order by (created_at, updated_at, title)",
						"enum":        []string{"created_at", "updated_at", "title"},
					},
					"order_dir": map[string]interface{}{
						"type":        "string",
						"description": "Order direction (asc, desc)",
						"enum":        []string{"asc", "desc"},
					},
				},
			},
		},
	}

	for _, tool := range tools {
		t := mcp.Tool{
			Name:        tool.name,
			Description: tool.description,
			InputSchema: tool.schema,
		}
		s.AddTool(t, tool.handler)
	}

	return nil
}