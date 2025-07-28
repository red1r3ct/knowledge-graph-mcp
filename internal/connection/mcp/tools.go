package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/red1r3ct/knowledge-graph-mcp/internal/connection"
)

// RegisterTools registers all connection MCP tools with the server
func RegisterTools(s *server.MCPServer, storage connection.Storage) error {
	tools := []struct {
		name        string
		description string
		handler     server.ToolHandlerFunc
		schema      mcp.ToolInputSchema
	}{
		{
			name:        "create_connection",
			description: "Create a new connection between two notes",
			handler:     NewCreateHandler(storage),
			schema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"from_note_id": map[string]interface{}{
						"type":        "integer",
						"description": "ID of the source note",
					},
					"to_note_id": map[string]interface{}{
						"type":        "integer",
						"description": "ID of the target note",
					},
					"type": map[string]interface{}{
						"type":        "string",
						"description": "Type of connection (e.g., relates_to, references, supports, etc.)",
						"enum":        connection.ValidConnectionTypes(),
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Optional description of the connection",
					},
					"strength": map[string]interface{}{
						"type":        "integer",
						"description": "Strength of the connection (1-10, default: 5)",
						"minimum":     1,
						"maximum":     10,
					},
					"metadata": map[string]interface{}{
						"type":        "object",
						"description": "Optional metadata for the connection",
					},
				},
				Required: []string{"from_note_id", "to_note_id", "type"},
			},
		},
		{
			name:        "get_connection",
			description: "Get a connection by ID",
			handler:     NewGetHandler(storage),
			schema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "integer",
						"description": "Unique identifier of the connection",
					},
				},
				Required: []string{"id"},
			},
		},
		{
			name:        "update_connection",
			description: "Update an existing connection",
			handler:     NewUpdateHandler(storage),
			schema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "integer",
						"description": "Unique identifier of the connection",
					},
					"type": map[string]interface{}{
						"type":        "string",
						"description": "Updated type of connection",
						"enum":        connection.ValidConnectionTypes(),
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Updated description of the connection",
					},
					"strength": map[string]interface{}{
						"type":        "integer",
						"description": "Updated strength of the connection (1-10)",
						"minimum":     1,
						"maximum":     10,
					},
					"metadata": map[string]interface{}{
						"type":        "object",
						"description": "Updated metadata for the connection",
					},
				},
				Required: []string{"id"},
			},
		},
		{
			name:        "delete_connection",
			description: "Delete a connection by ID",
			handler:     NewDeleteHandler(storage),
			schema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "integer",
						"description": "Unique identifier of the connection to delete",
					},
				},
				Required: []string{"id"},
			},
		},
		{
			name:        "list_connections",
			description: "List connections with optional filtering and pagination",
			handler:     NewListHandler(storage),
			schema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of connections to return (default: 100)",
						"minimum":     1,
						"maximum":     1000,
					},
					"offset": map[string]interface{}{
						"type":        "integer",
						"description": "Number of connections to skip (default: 0)",
						"minimum":     0,
					},
					"from_note_id": map[string]interface{}{
						"type":        "integer",
						"description": "Filter by source note ID",
					},
					"to_note_id": map[string]interface{}{
						"type":        "integer",
						"description": "Filter by target note ID",
					},
					"type": map[string]interface{}{
						"type":        "string",
						"description": "Filter by connection type",
						"enum":        connection.ValidConnectionTypes(),
					},
					"strength": map[string]interface{}{
						"type":        "integer",
						"description": "Filter by connection strength",
						"minimum":     1,
						"maximum":     10,
					},
					"order_by": map[string]interface{}{
						"type":        "string",
						"description": "Field to order by (default: id)",
						"enum":        []string{"id", "created_at", "updated_at", "strength", "type"},
					},
					"order_dir": map[string]interface{}{
						"type":        "string",
						"description": "Order direction (default: asc)",
						"enum":        []string{"asc", "desc"},
					},
				},
			},
		},
		{
			name:        "get_note_connections",
			description: "Get all connections for a specific note (incoming and outgoing)",
			handler:     NewNoteConnectionsHandler(storage),
			schema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"note_id": map[string]interface{}{
						"type":        "integer",
						"description": "ID of the note to get connections for",
					},
					"type": map[string]interface{}{
						"type":        "string",
						"description": "Filter by connection type",
						"enum":        connection.ValidConnectionTypes(),
					},
					"strength": map[string]interface{}{
						"type":        "integer",
						"description": "Filter by connection strength",
						"minimum":     1,
						"maximum":     10,
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of connections to return (default: 100)",
						"minimum":     1,
						"maximum":     1000,
					},
					"offset": map[string]interface{}{
						"type":        "integer",
						"description": "Number of connections to skip (default: 0)",
						"minimum":     0,
					},
				},
				Required: []string{"note_id"},
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