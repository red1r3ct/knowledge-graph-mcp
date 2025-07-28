package ncruces

import (
	"errors"
	"fmt"
	"time"
)

var (
	// ErrLockTimeout is returned when lock acquisition times out
	ErrLockTimeout = errors.New("lock acquisition timeout")
	
	// ErrLockHeld is returned when trying to acquire a lock that's already held
	ErrLockHeld = errors.New("lock already held")
	
	// ErrInvalidConfig is returned when configuration is invalid
	ErrInvalidConfig = errors.New("invalid configuration")
	
	// ErrDatabaseClosed is returned when operations are attempted on a closed database
	ErrDatabaseClosed = errors.New("database connection is closed")
	
	// ErrInvalidMigration is returned when migration content is invalid
	ErrInvalidMigration = errors.New("invalid migration content")
	
	// ErrTransactionFailed is returned when a transaction fails
	ErrTransactionFailed = errors.New("transaction failed")
)

// LockError represents a lock-related error
type LockError struct {
	Operation string
	Owner     string
	Duration  time.Duration
}

func (e *LockError) Error() string {
	return fmt.Sprintf("lock %s failed: owner=%s, duration=%v", e.Operation, e.Owner, e.Duration)
}

// ConfigError represents a configuration-related error
type ConfigError struct {
	Field   string
	Value   string
	Message string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("configuration error: field=%s, value=%s, message=%s", e.Field, e.Value, e.Message)
}

// MigrationError represents a migration-related error
type MigrationError struct {
	Version int
	Dirty   bool
	Message string
}

func (e *MigrationError) Error() string {
	return fmt.Sprintf("migration error: version=%d, dirty=%v, message=%s", e.Version, e.Dirty, e.Message)
}

// DatabaseError represents a database-related error
type DatabaseError struct {
	Operation string
	Message   string
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("database error: operation=%s, message=%s", e.Operation, e.Message)
}