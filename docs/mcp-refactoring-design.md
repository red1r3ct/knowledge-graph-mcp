# MCP Tools Refactoring Design

## Overview
This document outlines the refactoring of MCP tools to use the Storage interface instead of concrete SQLite storage, enabling better testability and adherence to dependency inversion principle.

## Goals
1. Decouple MCP handlers from concrete SQLite storage
2. Enable comprehensive unit testing with mocked storage
3. Maintain backward compatibility
4. Improve testability and maintainability

## Changes Required

### 1. Storage Interface Enhancement
- Add go:generate directive to `internal/knowledgebase/storage.go`
- Generate mock implementation using gomock
- Ensure Storage interface is properly defined for all CRUD operations

### 2. MCP Tools Refactoring
- Update all MCP handlers to accept Storage interface instead of concrete SQLite storage
- Modify handler signatures to use dependency injection
- Update tools registration to pass Storage interface

### 3. Testing Strategy
- Create comprehensive table-driven tests for all 5 MCP tools
- Use gomock-generated mocks for Storage interface
- Test success and error scenarios
- Verify proper error handling and edge cases

### 4. Dependencies
- Add gomock and mockgen to go.mod
- Ensure proper version compatibility

## Files to Modify
- `internal/knowledgebase/storage.go` - Add go:generate directive
- `internal/knowledgebase/mcp/tools.go` - Update to use Storage interface
- `internal/knowledgebase/mcp/create_handler.go` - Refactor to use Storage
- `internal/knowledgebase/mcp/get_handler.go` - Refactor to use Storage
- `internal/knowledgebase/mcp/update_handler.go` - Refactor to use Storage
- `internal/knowledgebase/mcp/delete_handler.go` - Refactor to use Storage
- `internal/knowledgebase/mcp/list_handler.go` - Refactor to use Storage
- `cmd/knowledge-base-stdin/main.go` - Update initialization

## New Files
- `internal/knowledgebase/mock/storage.go` - Generated mock storage
- `internal/knowledgebase/mcp/tools_test.go` - Comprehensive tests

## Testing Approach
- Use table-driven tests for all scenarios
- Mock storage responses for both success and error cases
- Test edge cases like empty results, not found errors, etc.
- Verify proper error propagation to MCP clients