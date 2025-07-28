package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/red1r3ct/knowledge-graph-mcp/internal/knowledgebase"
)

// RegisterTools registers all knowledge base MCP tools with the server
func RegisterTools(s *server.MCPServer, storage knowledgebase.Storage) error {
	tools := []struct {
		name        string
		description string
		handler     server.ToolHandlerFunc
		schema      mcp.ToolInputSchema
	}{
		{
			name:        "create_knowledge_base",
			description: "Create a new knowledge base entry",
			handler:     NewCreateHandler(storage),
			schema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the knowledge base entry",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Description of the knowledge base entry",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"description": "Tags associated with the knowledge base entry",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
				},
				Required: []string{"name"},
			},
		},
		{
			name:        "get_knowledge_base",
			description: "Get a knowledge base entry by ID",
			handler:     NewGetHandler(storage),
			schema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Unique identifier of the knowledge base entry",
					},
				},
				Required: []string{"id"},
			},
		},
		{
			name:        "update_knowledge_base",
			description: "Update an existing knowledge base entry",
			handler:     NewUpdateHandler(storage),
			schema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Unique identifier of the knowledge base entry",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Updated name of the knowledge base entry",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Updated description of the knowledge base entry",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"description": "Updated tags associated with the knowledge base entry",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
				},
				Required: []string{"id"},
			},
		},
		{
			name:        "delete_knowledge_base",
			description: "Delete a knowledge base entry by ID",
			handler:     NewDeleteHandler(storage),
			schema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Unique identifier of the knowledge base entry to delete",
					},
				},
				Required: []string{"id"},
			},
		},
		{
			name:        "list_knowledge_bases",
			description: "List all knowledge base entries with optional filtering",
			handler:     NewListHandler(storage),
			schema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of entries to return (default: 100)",
						"minimum":     1,
						"maximum":     1000,
					},
					"offset": map[string]interface{}{
						"type":        "integer",
						"description": "Number of entries to skip (default: 0)",
						"minimum":     0,
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"description": "Filter by tags (returns entries that have any of the specified tags)",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"search": map[string]interface{}{
						"type":        "string",
						"description": "Search term to filter entries by name or description",
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
