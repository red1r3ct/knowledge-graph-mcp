# Initial Design: Knowledge Base MCP Server

## Overview

This document outlines the initial design for a Model Context Protocol (MCP) server implementation for managing a knowledge base. The server provides CRUDL (Create, Read, Update, Delete, List) operations for knowledge base entities.

## Architecture

### Domain Model

The system is built around the `KnowledgeBase` entity with the following structure:

- **Name**: Required string field
- **Description**: Optional string field
- **Tags**: Optional array of strings
- **CreatedAt**: Timestamp for creation
- **UpdatedAt**: Timestamp for last update

### Layered Architecture

1. **Domain Layer** (`internal/knowledge-base/`):
   - Contains domain models and DTOs
   - Defines storage interfaces
   - Business logic and validation

2. **Infrastructure Layer** (`internal/knowledge-base/sqlite/`):
   - SQLite database implementation
   - Database migrations
   - SQL queries and data persistence

3. **Application Layer** (`cmd/knowledge-base-stdin/`):
   - MCP server implementation
   - Tool definitions and handlers
   - Request/response mapping

## API Design

### MCP Tools

The server exposes the following tools:

1. **create_knowledge_base**
   - Creates a new knowledge base
   - Parameters: name (required), description (optional), tags (optional)

2. **get_knowledge_base**
   - Retrieves a knowledge base by ID
   - Parameters: id (required)

3. **update_knowledge_base**
   - Updates an existing knowledge base
   - Parameters: id (required), name (optional), description (optional), tags (optional)

4. **delete_knowledge_base**
   - Deletes a knowledge base by ID
   - Parameters: id (required)

5. **list_knowledge_bases**
   - Lists knowledge bases with optional filtering
   - Parameters: limit (optional), offset (optional), search (optional), tags (optional)

### Data Flow

```
MCP Client → MCP Server → Storage Interface → SQLite → Database
```

## Database Schema

### knowledge_base Table

```sql
CREATE TABLE knowledge_base (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    tags TEXT, -- JSON array of tags
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### Indexes

- `idx_knowledge_base_name` on name for faster lookups
- `idx_knowledge_base_created_at` on created_at for sorting

### Trigger

- `update_knowledge_base_updated_at` automatically updates the updated_at timestamp

## Testing Strategy

### Unit Tests

- **Storage Tests**: Comprehensive tests for all CRUDL operations
- **Table-driven tests**: Cover edge cases and multiple scenarios
- **Isolation**: Each test uses a fresh temporary database

### Test Categories

1. **Create Operations**
   - Valid creation with all fields
   - Creation with required fields only
   - Validation errors

2. **Read Operations**
   - Retrieval of existing records
   - Handling of non-existent records

3. **Update Operations**
   - Partial updates (patch-style)
   - Full updates
   - Non-existent record handling

4. **Delete Operations**
   - Successful deletion
   - Non-existent record handling

5. **List Operations**
   - Pagination
   - Search functionality
   - Tag filtering
   - Sorting (newest first)

## Implementation Notes

### Separation of Concerns

- **DTOs vs Models**: Clear separation between Data Transfer Objects (requests/responses) and domain models
- **Storage Interface**: Abstract interface allows for different storage implementations
- **Error Handling**: Consistent error handling across all layers

### Dependencies

- **mcp-go**: MCP protocol implementation
- **sqlite3**: SQLite database driver
- **testify**: Testing utilities and assertions
- **migrate**: Database migrations (future enhancement)

## Future Enhancements

1. **Advanced Search**: Full-text search capabilities
2. **Validation**: More sophisticated validation rules
3. **Caching**: Performance optimization with caching layer
4. **Metrics**: Observability and monitoring
5. **Configuration**: Environment-based configuration
6. **Migrations**: Use golang-migrate for production-grade migrations

## Getting Started

1. Initialize the database: `go run cmd/knowledge-base-stdin/main.go`
2. The server will create the database file `knowledge_base.db`
3. Use any MCP client to interact with the tools

## Example Usage

```json
// Create a knowledge base
{
  "tool": "create_knowledge_base",
  "arguments": {
    "name": "My Knowledge Base",
    "description": "A collection of useful information",
    "tags": ["personal", "reference"]
  }
}

// List knowledge bases
{
  "tool": "list_knowledge_bases",
  "arguments": {
    "limit": 10,
    "search": "knowledge"
  }
}