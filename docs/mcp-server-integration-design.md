# MCP Server Integration Design

## Overview

This document outlines the integration of Note and Connection entities into the existing MCP server alongside the Knowledge Base entity. The integration follows the established patterns and maintains consistency with the current architecture.

## Current State

The MCP server currently supports:
- Knowledge Base entity with CRUD operations
- SQLite storage with migrations
- MCP tool registration pattern

## Integration Requirements

### 1. Storage Integration
- Initialize Note SQLite storage instance
- Initialize Connection SQLite storage instance
- Ensure proper resource cleanup (defer Close())
- Use the same database file for all entities

### 2. MCP Tools Registration
- Register all Note MCP tools (create_note, get_note, update_note, delete_note, list_notes)
- Register all Connection MCP tools (create_connection, get_connection, update_connection, delete_connection, list_connections, get_note_connections)
- Follow the same pattern as Knowledge Base tools registration

### 3. Error Handling
- Proper error handling for storage initialization
- Graceful failure if any component fails to initialize
- Consistent error messages and logging

## Implementation Plan

### Step 1: Update Imports
Add imports for Note and Connection packages:
```go
notemcp "github.com/red1r3ct/knowledge-graph-mcp/internal/note/mcp"
notestorage "github.com/red1r3ct/knowledge-graph-mcp/internal/note/sqlite"
connmcp "github.com/red1r3ct/knowledge-graph-mcp/internal/connection/mcp"
connstorage "github.com/red1r3ct/knowledge-graph-mcp/internal/connection/sqlite"
```

### Step 2: Initialize Storage Instances
Create storage instances for Note and Connection entities:
```go
// Initialize note storage
noteStorage, err := notestorage.NewStorage(dbPath)
if err != nil {
    log.Fatalf("Failed to initialize note storage: %v", err)
}
defer noteStorage.Close()

// Initialize connection storage
connStorage, err := connstorage.NewStorage(dbPath)
if err != nil {
    log.Fatalf("Failed to initialize connection storage: %v", err)
}
defer connStorage.Close()
```

### Step 3: Register MCP Tools
Register tools for all entities:
```go
// Register knowledgebase tools
if err := kbmcp.RegisterTools(s, storage); err != nil {
    log.Fatalf("Failed to register knowledgebase tools: %v", err)
}

// Register note tools
if err := notemcp.RegisterTools(s, noteStorage); err != nil {
    log.Fatalf("Failed to register note tools: %v", err)
}

// Register connection tools
if err := connmcp.RegisterTools(s, connStorage); err != nil {
    log.Fatalf("Failed to register connection tools: %v", err)
}
```

### Step 4: Update Server Metadata
Update server name and description to reflect all supported entities:
```go
s := server.NewMCPServer(
    "Knowledge Graph MCP Server",
    "1.0.0",
)
```

## Testing Strategy

### 1. Build Test
- Ensure the application builds without errors
- Verify all imports are resolved correctly

### 2. Integration Test
- Test that all storage instances initialize correctly
- Verify all MCP tools are registered
- Ensure the server starts without errors

### 3. Functional Test
- Test basic operations for each entity type
- Verify tool schemas are properly registered
- Test error handling scenarios

## Acceptance Criteria

- [ ] Application builds successfully
- [ ] All storage instances initialize without errors
- [ ] All MCP tools are registered (15 total: 5 KB + 5 Note + 6 Connection)
- [ ] Server starts and runs without errors
- [ ] All unit tests pass
- [ ] Integration follows existing patterns exactly
- [ ] Proper resource cleanup is implemented
- [ ] Error handling is consistent and informative

## Risk Mitigation

1. **Storage Initialization Failures**: Each storage initialization is wrapped with proper error handling
2. **Tool Registration Conflicts**: Tool names are unique across all entities
3. **Resource Leaks**: All storage instances have proper defer Close() calls
4. **Migration Dependencies**: Migrations are run before any storage initialization

## Dependencies

- Note entity implementation (completed)
- Connection entity implementation (completed)
- SQLite storage implementations (completed)
- MCP handlers and tools registration (completed)
- Database migrations (completed)