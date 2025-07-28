# Note Storage Implementation Design

## Overview

This document describes the implementation of the Note storage interface and SQLite storage implementation following Test-Driven Development (TDD) approach and exact patterns from the knowledgebase implementation.

## Architecture

The implementation follows the same domain-driven design approach as the knowledgebase:

### Directory Structure

```
internal/note/
├── model.go              # Domain models and DTOs
├── storage.go            # Storage interface (DAO)
└── sqlite/               # SQLite implementation
    ├── storage.go        # SQLite storage implementation
    └── storage_test.go   # Unit tests for SQLite storage
```

### Package Structure

- **Domain Layer**: `internal/note/`
  - `model.go`: Contains domain models (Note) and DTOs (CreateNoteRequest, UpdateNoteRequest, etc.)
  - `storage.go`: Defines storage interface (DAO) that works with domain models

- **Infrastructure Layer**: `internal/note/sqlite/`
  - `storage.go`: Implements the storage interface using SQLite
  - `storage_test.go`: Unit tests for the SQLite implementation

## Storage Interface

The storage interface follows the exact same pattern as knowledgebase storage:

```go
type Storage interface {
    Create(ctx context.Context, req CreateNoteRequest) (*Note, error)
    Get(ctx context.Context, id int64) (*Note, error)
    Update(ctx context.Context, id int64, req UpdateNoteRequest) (*Note, error)
    Delete(ctx context.Context, id int64) error
    List(ctx context.Context, req ListNotesRequest) (*ListNotesResponse, error)
}
```

## SQLite Implementation Features

### Core CRUD Operations
- **Create**: Insert new notes with JSON serialization for tags and metadata
- **Get**: Retrieve notes by ID with proper JSON deserialization
- **Update**: Dynamic update queries that only modify specified fields
- **Delete**: Remove notes with proper error handling for non-existent records
- **List**: Paginated listing with filtering and search capabilities

### Advanced Features
- **Full-Text Search**: Uses SQLite FTS5 virtual table for efficient content search
- **JSON Storage**: Proper handling of tags ([]string) and metadata (map[string]interface{})
- **Filtering**: Support for filtering by type, tags, and search terms
- **Ordering**: Configurable ordering by any field with ASC/DESC direction
- **Pagination**: Limit/offset based pagination with total count

### Database Schema
```sql
CREATE TABLE note (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    type TEXT NOT NULL,
    tags TEXT,
    metadata TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- FTS5 virtual table for full-text search
CREATE VIRTUAL TABLE note_fts USING fts5(
    title, content, content='note', content_rowid='id'
);
```

### Indexes
- `idx_note_title`: Index on title for efficient title-based queries
- `idx_note_type`: Index on type for efficient type filtering
- `idx_note_created_at`: Index on created_at for efficient ordering

### Triggers
- **FTS Sync Triggers**: Keep FTS table synchronized with main table
- **Update Timestamp Trigger**: Automatically update `updated_at` on record changes

## Test-Driven Development Approach

### Test Structure
The tests follow the exact same table-driven test pattern as knowledgebase:

1. **Create Tests**: Test all field combinations, edge cases, and validation
2. **Get Tests**: Test existing and non-existing record retrieval
3. **Update Tests**: Test partial and full updates, non-existing records
4. **Delete Tests**: Test successful deletion and error cases
5. **List Tests**: Test pagination, filtering, search, and ordering

### Test Coverage
- ✅ All CRUD operations
- ✅ JSON serialization/deserialization
- ✅ Full-text search functionality
- ✅ Filtering by type and tags
- ✅ Pagination and ordering
- ✅ Error handling for edge cases
- ✅ Database constraints and triggers

## Implementation Patterns

### Error Handling
- Consistent error wrapping with context
- Proper handling of `sql.ErrNoRows`
- Descriptive error messages with entity IDs

### JSON Handling
- Safe marshaling/unmarshaling of tags and metadata
- Proper handling of nil values (empty arrays vs null)
- Type preservation where possible

### Dynamic Queries
- Builder pattern for UPDATE queries
- Safe parameter binding to prevent SQL injection
- Conditional clause building for filtering

### Resource Management
- Proper database connection handling
- Deferred cleanup of resources
- Transaction safety (where applicable)

## Compliance with Project Standards

### Naming Conventions
- ✅ Package names: lowercase, no underscores
- ✅ File names: snake_case
- ✅ Types: PascalCase for exported, camelCase for unexported
- ✅ Functions: PascalCase for exported, camelCase for unexported

### Code Organization
- ✅ Clear separation between domain and infrastructure layers
- ✅ Interface definition at domain level
- ✅ Implementation in separate package
- ✅ Comprehensive test coverage

### Dependencies
- ✅ Uses same SQLite driver as knowledgebase (`github.com/ncruces/go-sqlite3`)
- ✅ Uses same testing framework (`github.com/stretchr/testify`)
- ✅ Follows same import organization

## Acceptance Criteria

- [x] Storage interface implemented with CRUDL operations
- [x] SQLite implementation following exact knowledgebase patterns
- [x] Comprehensive unit tests with table-driven approach
- [x] Full-text search capabilities using FTS5
- [x] JSON metadata storage with proper serialization
- [x] Foreign key relationships ready (for future connection to knowledge_base)
- [x] All tests passing
- [x] Error handling and validation implemented
- [x] Proper resource management and cleanup

## Future Enhancements

1. **Foreign Key Relationships**: Add knowledge_base_id foreign key when needed
2. **Batch Operations**: Implement batch create/update/delete operations
3. **Soft Deletes**: Add soft delete functionality with deleted_at timestamp
4. **Audit Trail**: Add created_by/updated_by fields for user tracking
5. **Versioning**: Implement note versioning for change history