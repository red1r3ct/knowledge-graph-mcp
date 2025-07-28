# Agent Rules Standard (AGENTS.md)
## Rules

- Start from creating short design document with key changes in markdown format in docs/
- Follow test driven development. First write design doc with acceptance criteria, then write table tests with test cases, only then start to implement changes
- Unit test must be written in _test packages and use tested package using imports
- Each design document must have short but easy to read name
- After changes are ready run unit-tests and fix error. Make sure you are following test cases and design doc
- Check acceptance criteria after changes are finished
- If you stuck with same command more the 3 times or stuck in the loop, ask user for help
- Install dependencies only using golang cli

## Project Structure

The project follows a domain-driven design approach with clear separation of concerns:

### Directory Structure

```
knowledge-graph-mcp/
├── cmd/                          # Entry points for different applications
│   └── knowledge-base-stdin/     # MCP server using stdio transport
│       └── main.go            # Main entry point for MCP server
├── internal/                    # Internal packages (not importable)
│   └── knowledgebase/          # Knowledge base domain
│       ├── model.go             # Domain models and DTOs
│       ├── storage.go         # Storage interface (DAO)
│       └── sqlite/              # SQLite implementation
│           ├── storage.go       # SQLite storage implementation
│           └── storage_test.go  # Unit tests for SQLite storage
├── docs/                      # Design documents and specifications
├── go.mod                     # Go module file
├── go.sum                     # Go checksum file
└── README.md                  # Project documentation
```

### Package Structure

- **Domain Layer**: `internal/knowledge-base/`
  - `model.go`: Contains domain models (KnowledgeBase) and DTOs (CreateRequest, UpdateRequest, etc.)
  - `storage.go`: Defines storage interface (DAO) that works with domain models

- **Infrastructure Layer**: `internal/knowledge-base/sqlite/`
  - `storage.go`: Implements the storage interface using SQLite
  - `storage_test.go`: Unit tests for the SQLite implementation

- **Application Layer**: `cmd/knowledge-base-stdin/`
  - `main.go`: MCP server implementation that uses the domain and infrastructure layers

### Key Design Principles

1. **Separation of Concerns**: Clear separation between DTOs (Data Transfer Objects) and domain models
2. **Interface Segregation**: Storage interface is defined at the domain level, implementations are in separate packages
3. **Dependency Inversion**: Domain layer depends on abstractions (interfaces), not concrete implementations
4. **Testability**: Each layer can be tested independently with table-driven tests
5. **Migrations**: Database schema changes are managed through versioned migration files

### Naming Conventions

- **Packages**: Use lowercase, no underscores (e.g., `knowledgebase`, not `knowledge_base`)
- **Files**: Use snake_case for file names (e.g., `storage_test.go`)
- **Types**: Use PascalCase for exported types (e.g., `KnowledgeBase`)
- **Functions**: Use PascalCase for exported functions, camelCase for unexported
- **Variables**: Use camelCase for variables

### Testing Strategy

- **Unit Tests**: Focus on individual components (storage layer)
- **Table-Driven Tests**: Use table-driven tests for comprehensive test coverage
- **Test Isolation**: Each test uses a fresh temporary database
- **Assertions**: Use testify for clear, readable assertions