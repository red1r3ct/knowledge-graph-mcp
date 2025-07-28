# Custom go-migrate SQLite Driver for ncruces/go-sqlite3

## Overview

This document outlines the design for implementing a custom database driver for [go-migrate](https://github.com/golang-migrate/migrate) that uses the [ncruces/go-sqlite3](https://github.com/ncruces/go-sqlite3) SQLite driver instead of the traditional mattn/go-sqlite3 driver. This implementation provides better performance, modern SQLite features, and improved compatibility with the latest SQLite versions.

## Background

The current project uses go-migrate for database migrations but relies on the standard SQLite driver. The ncruces/go-sqlite3 driver offers several advantages:
- Pure Go implementation with cgo-free builds
- Support for modern SQLite features (3.45+)
- Better performance characteristics
- Improved JSON and extension support
- Active maintenance and development

## Architecture Overview

### Driver Structure

The custom driver will be implemented as a go-migrate `database.Driver` interface implementation:

```go
type Driver interface {
    Open(url string) (Driver, error)
    Close() error
    Lock() error
    Unlock() error
    Run(migration io.Reader) error
    SetVersion(version int, dirty bool) error
    Version() (version int, dirty bool, err error)
    Drop() error
    WithInstance(instance interface{}, config *Config) (Driver, error)
}
```

### Key Components

1. **SQLite Driver**: Wraps ncruces/go-sqlite3 database connection
2. **Migration Table Manager**: Handles schema_migrations table operations
3. **Lock Manager**: Implements SQLite-based locking for concurrent migrations
4. **Statement Executor**: Processes migration SQL with proper transaction handling

## Key Differences from Reference Implementation

### 1. Driver Registration
- **Reference**: Uses `sqlite3` driver name
- **Custom**: Uses `ncruces-sqlite3` driver name to avoid conflicts
- **Impact**: Requires explicit driver selection in migration URLs

### 2. Connection String Format
- **Reference**: `sqlite3:///path/to/db.sqlite3`
- **Custom**: `ncruces-sqlite3:///path/to/db.sqlite3`
- **Impact**: Backward compatible with existing migration files

### 3. Transaction Handling
- **Reference**: Uses standard database/sql transactions
- **Custom**: Leverages ncruces/go-sqlite3's enhanced transaction support
- **Impact**: Better performance for large migrations

### 4. Locking Mechanism
- **Reference**: File-based locking with timeout
- **Custom**: SQLite table-based locking with better concurrency
- **Impact**: More reliable in containerized environments

### 5. Error Handling
- **Reference**: Generic SQLite error codes
- **Custom**: Rich error types with context
- **Impact**: Better debugging and troubleshooting

## Architecture Decisions

### Decision 1: Driver Location
**Choice**: Place driver in `internal/migrations/driver/ncruces/`
**Rationale**: Keeps migration-specific code isolated from domain logic

### Decision 2: Configuration Structure
```go
type Config struct {
    DatabaseName    string
    MigrationsTable string
    NoTxWrap        bool
    TxMode          string // "DEFERRED", "IMMEDIATE", "EXCLUSIVE"
}
```

### Decision 3: Migration Table Schema
```sql
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    dirty BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Decision 4: Locking Strategy
**Approach**: SQLite table-based advisory locks
**Implementation**:
```sql
CREATE TABLE IF NOT EXISTS schema_migrations_lock (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    locked BOOLEAN NOT NULL DEFAULT FALSE,
    owner TEXT,
    acquired_at TIMESTAMP,
    CHECK (id = 1)
);
```

## File Structure

```
internal/migrations/
├── driver/
│   └── ncruces/
│       ├── driver.go           # Main driver implementation
│       ├── driver_test.go      # Unit tests
│       ├── config.go          # Configuration structures
│       ├── errors.go          # Custom error types
│       └── lock.go            # Locking mechanism
├── sqlite/
│   ├── 000001_create_knowledge_base_table.up.sql
│   └── 000001_create_knowledge_base_table.down.sql
├── embed.go                   # Migration embedding
└── migrations.go              # Migration orchestration
```

## Implementation Details

### Driver Registration
```go
func init() {
    database.Register("ncruces-sqlite3", &Driver{})
}
```

### Connection Handling
```go
func (d *Driver) Open(url string) (database.Driver, error) {
    // Parse URL and extract database path
    // Register ncruces/go-sqlite3 driver
    // Initialize connection with proper pragmas
}
```

### Migration Execution
```go
func (d *Driver) Run(migration io.Reader) error {
    // Begin transaction based on TxMode
    // Execute migration SQL
    // Handle errors with rollback
    // Commit on success
}
```

### Lock Management
```go
func (d *Driver) Lock() error {
    // Attempt to acquire advisory lock
    // Use INSERT ... ON CONFLICT for atomicity
    // Implement exponential backoff for retries
}

func (d *Driver) Unlock() error {
    // Release advisory lock
    // Clean up lock table
}
```

## Acceptance Criteria

### Functional Requirements
- [ ] Successfully registers as go-migrate driver with name "ncruces-sqlite3"
- [ ] Supports all standard go-migrate operations (up, down, drop, version)
- [ ] Handles concurrent migrations safely with proper locking
- [ ] Supports both transactional and non-transactional migrations
- [ ] Provides clear error messages for common failure scenarios

### Performance Requirements
- [ ] Migration execution time within 10% of reference implementation
- [ ] Lock acquisition timeout configurable (default: 15 seconds)
- [ ] Supports batch operations for large datasets

### Compatibility Requirements
- [ ] Compatible with SQLite 3.45+
- [ ] Supports Windows, macOS, and Linux
- [ ] Works with existing migration files without modification
- [ ] Supports both relative and absolute database paths

### Testing Requirements
- [ ] Unit test coverage > 90% for driver package
- [ ] Integration tests with actual SQLite database
- [ ] Concurrent migration tests
- [ ] Error handling and recovery tests
- [ ] Performance benchmarks against reference implementation

## Testing Strategy

### Unit Tests
- Driver initialization and configuration
- Lock acquisition and release
- Migration execution and rollback
- Error handling scenarios
- Configuration validation

### Integration Tests
- End-to-end migration workflows
- Concurrent migration safety
- Database file permissions
- Migration table schema validation

### Performance Tests
- Migration execution benchmarks
- Lock contention scenarios
- Memory usage profiling
- Large dataset handling

## Migration Path

### Phase 1: Driver Implementation
1. Create basic driver structure
2. Implement core Driver interface methods
3. Add configuration support

### Phase 2: Locking and Safety
1. Implement SQLite-based locking
2. Add concurrent migration tests
3. Performance optimization

### Phase 3: Integration
1. Update migration orchestration
2. Add configuration validation
3. Documentation and examples

### Phase 4: Validation
1. Comprehensive testing
2. Performance benchmarking
3. Production readiness review

## Configuration Examples

### Basic Usage
```go
m, err := migrate.New(
    "file://migrations",
    "ncruces-sqlite3:///path/to/database.sqlite3",
)
```

### Advanced Configuration
```go
m, err := migrate.New(
    "file://migrations",
    "ncruces-sqlite3:///path/to/database.sqlite3?x-migrations-table=custom_migrations&x-no-tx-wrap=true",
)
```

## Risk Mitigation

### Risk 1: Driver Compatibility
**Mitigation**: Extensive testing with various SQLite versions and platforms

### Risk 2: Performance Regression
**Mitigation**: Benchmarking against reference implementation with performance budgets

### Risk 3: Locking Deadlocks
**Mitigation**: Comprehensive testing of concurrent scenarios with timeout handling

### Risk 4: Migration Corruption
**Mitigation**: Transactional safety with rollback capabilities and integrity checks

## Future Enhancements

- Support for SQLite extensions (JSON, FTS5)
- WAL mode optimization
- Backup and restore capabilities
- Migration checksum validation
- Dry-run mode for validation

## Dependencies

- github.com/ncruces/go-sqlite3 v0.19.0+
- github.com/golang-migrate/migrate/v4 v4.17.0+
- github.com/stretchr/testify v1.8.0+ (testing)

## Success Metrics

- All acceptance criteria met
- Zero data corruption incidents
- Performance within 10% of reference
- 100% test pass rate
- Successful production deployment