# Connection MCP Handlers Design

## Overview

This document outlines the design for implementing MCP (Model Context Protocol) handlers for Connection CRUDL operations. The handlers will follow the established patterns from the knowledgebase and note MCP handlers, providing consistent API interfaces for connection management.

## Acceptance Criteria

1. **Create Handler**: Accept CreateConnectionRequest and return created Connection
2. **Get Handler**: Accept connection ID and return Connection or appropriate error
3. **Update Handler**: Accept connection ID and UpdateConnectionRequest, return updated Connection
4. **Delete Handler**: Accept connection ID and return success/error status
5. **List Handler**: Accept ListConnectionsRequest with filtering and return paginated results
6. **Note Connections Handler**: Accept NoteConnectionsRequest and return all connections for a note
7. **Tool Registration**: Register all handlers as MCP tools with proper schemas
8. **Error Handling**: Consistent error responses following established patterns
9. **Validation**: Input validation for all requests
10. **Testing**: Comprehensive table-driven tests for all handlers

## Handler Specifications

### 1. Create Handler
- **Tool Name**: `create_connection`
- **Input**: CreateConnectionRequest (from_note_id, to_note_id, type, description, strength, metadata)
- **Output**: Created Connection with generated ID and timestamps
- **Validation**: 
  - Valid connection type
  - Strength between 1-10
  - from_note_id != to_note_id
  - Both notes exist

### 2. Get Handler
- **Tool Name**: `get_connection`
- **Input**: Connection ID (int64)
- **Output**: Connection or error if not found
- **Validation**: ID must be positive integer

### 3. Update Handler
- **Tool Name**: `update_connection`
- **Input**: Connection ID + UpdateConnectionRequest (optional fields)
- **Output**: Updated Connection
- **Validation**: 
  - Connection exists
  - Valid connection type if provided
  - Valid strength if provided

### 4. Delete Handler
- **Tool Name**: `delete_connection`
- **Input**: Connection ID (int64)
- **Output**: Success confirmation
- **Validation**: Connection exists

### 5. List Handler
- **Tool Name**: `list_connections`
- **Input**: ListConnectionsRequest (limit, offset, filters, ordering)
- **Output**: ListConnectionsResponse with items and total count
- **Features**:
  - Pagination support
  - Filtering by from_note_id, to_note_id, type, strength
  - Ordering by id, created_at, updated_at, strength

### 6. Note Connections Handler
- **Tool Name**: `get_note_connections`
- **Input**: NoteConnectionsRequest (note_id, optional filters)
- **Output**: NoteConnectionsResponse with incoming/outgoing connections
- **Features**:
  - Separate incoming and outgoing connections
  - Type and strength filtering
  - Connection count statistics

## File Structure

```
internal/connection/mcp/
├── create_handler.go              # Create connection handler
├── get_handler.go                 # Get connection by ID handler
├── update_handler.go              # Update connection handler
├── delete_handler.go              # Delete connection handler
├── list_handler.go                # List connections with filtering
├── note_connections_handler.go    # Get all connections for a note
├── tools.go                       # Tool registration and factory
├── create_handler_test.go         # Tests for create handler
├── get_handler_test.go            # Tests for get handler
├── update_handler_test.go         # Tests for update handler
├── delete_handler_test.go         # Tests for delete handler
├── list_handler_test.go           # Tests for list handler
└── note_connections_handler_test.go # Tests for note connections handler
```

## Implementation Patterns

### Handler Structure
Each handler follows the pattern:
```go
func HandleToolName(storage connection.Storage) mcp.ToolHandler {
    return func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
        // 1. Parse and validate input
        // 2. Call storage layer
        // 3. Handle errors consistently
        // 4. Return formatted response
    }
}
```

### Error Handling
- Use consistent error messages and codes
- Map storage errors to appropriate MCP errors
- Validate input parameters before storage calls
- Return structured error responses

### Testing Strategy
- Table-driven tests for each handler
- Mock storage interface for isolated testing
- Test success cases and error scenarios
- Validate input parsing and output formatting
- Test edge cases and boundary conditions

## Dependencies

- `internal/connection` package for models and storage interface
- `github.com/mark3labs/mcp-go/mcp` for MCP types
- `github.com/mark3labs/mcp-go/server` for tool registration
- Mock storage for testing

## Integration

The handlers will be registered in the main MCP server alongside existing knowledgebase and note handlers, providing a unified API for all knowledge graph operations.