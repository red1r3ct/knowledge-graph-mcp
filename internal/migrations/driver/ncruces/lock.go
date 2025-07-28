package ncruces

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// LockManager handles SQLite-based advisory locking for migrations
type LockManager struct {
	db     *sql.DB
	config *Config
}

// NewLockManager creates a new lock manager
func NewLockManager(db *sql.DB, config *Config) *LockManager {
	return &LockManager{
		db:     db,
		config: config,
	}
}

// Acquire attempts to acquire a lock
func (lm *LockManager) Acquire(ctx context.Context) error {
	// Create lock table if it doesn't exist
	if err := lm.createLockTable(); err != nil {
		return fmt.Errorf("failed to create lock table: %w", err)
	}

	// Try to acquire lock with timeout
	timeout := 15 * time.Second
	deadline := time.Now().Add(timeout)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if time.Now().After(deadline) {
			return ErrLockTimeout
		}

		acquired, err := lm.tryAcquire()
		if err != nil {
			return fmt.Errorf("failed to acquire lock: %w", err)
		}

		if acquired {
			return nil
		}

		// Wait before retry
		time.Sleep(100 * time.Millisecond)
	}
}

// Release releases the lock
func (lm *LockManager) Release(ctx context.Context) error {
	// Check if lock table exists before attempting to release
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name='%s_lock')", lm.config.MigrationsTable)
	err := lm.db.QueryRowContext(ctx, query).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check lock table existence: %w", err)
	}
	
	if !exists {
		// Lock table doesn't exist, consider lock already released
		return nil
	}

	query = fmt.Sprintf("DELETE FROM %s_lock WHERE id = 1", lm.config.MigrationsTable)
	result, err := lm.db.ExecContext(ctx, query)
	if err != nil {
		// If table doesn't exist, consider it already dropped
		if strings.Contains(err.Error(), "no such table") {
			return nil
		}
		return fmt.Errorf("failed to release lock: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check lock release: %w", err)
	}

	if rows == 0 {
		// Lock might have been already released or table dropped
		return nil
	}

	return nil
}

// IsLocked checks if the lock is currently held
func (lm *LockManager) IsLocked(ctx context.Context) (bool, error) {
	var locked bool
	query := fmt.Sprintf("SELECT locked FROM %s_lock WHERE id = 1", lm.config.MigrationsTable)
	err := lm.db.QueryRowContext(ctx, query).Scan(&locked)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check lock status: %w", err)
	}
	return locked, nil
}

// createLockTable creates the lock table if it doesn't exist
func (lm *LockManager) createLockTable() error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s_lock (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			locked BOOLEAN NOT NULL DEFAULT FALSE,
			owner TEXT,
			acquired_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			CHECK (id = 1)
		)`, lm.config.MigrationsTable)

	_, err := lm.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create lock table: %w", err)
	}

	return nil
}

// tryAcquire attempts to acquire the lock
func (lm *LockManager) tryAcquire() (bool, error) {
	tx, err := lm.db.Begin()
	if err != nil {
		return false, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Ensure the lock table has the initial row
	query := fmt.Sprintf(`
		INSERT OR IGNORE INTO %s_lock (id, locked, owner, acquired_at)
		VALUES (1, FALSE, '', CURRENT_TIMESTAMP)`, lm.config.MigrationsTable)
	
	_, err = tx.Exec(query)
	if err != nil {
		return false, fmt.Errorf("failed to initialize lock row: %w", err)
	}

	// Try to acquire lock using atomic update
	query = fmt.Sprintf(`
		UPDATE %s_lock
		SET locked = TRUE, owner = 'migration', acquired_at = CURRENT_TIMESTAMP
		WHERE id = 1 AND locked = FALSE`, lm.config.MigrationsTable)

	result, err := tx.Exec(query)
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to check lock acquisition: %w", err)
	}

	if rows == 0 {
		return false, nil
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("failed to commit lock acquisition: %w", err)
	}

	return true, nil
}

// ForceRelease forces release of the lock (for cleanup)
func (lm *LockManager) ForceRelease(ctx context.Context) error {
	query := fmt.Sprintf("DELETE FROM %s_lock", lm.config.MigrationsTable)
	_, err := lm.db.ExecContext(ctx, query)
	return err
}