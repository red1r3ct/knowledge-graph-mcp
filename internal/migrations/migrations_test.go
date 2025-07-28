package migrations_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/red1r3ct/knowledge-graph-mcp/internal/migrations"
)

func TestMigrationRunner(t *testing.T) {
	// Create a temporary directory for the test database
	tempDir, err := os.MkdirTemp("", "migrations-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")

	t.Run("RunMigrations on fresh database", func(t *testing.T) {
		runner := migrations.NewMigrationRunner(dbPath)

		// Run migrations
		err := runner.RunMigrations()
		assert.NoError(t, err)

		// Verify migrations ran successfully
		version, dirty, err := runner.GetVersion()
		assert.NoError(t, err)
		assert.False(t, dirty)
		assert.Greater(t, version, uint(0))

		// Verify table was created
		db, err := sql.Open("sqlite3", dbPath)
		require.NoError(t, err)
		defer db.Close()

		var tableName string
		err = db.QueryRow(`
			SELECT name FROM sqlite_master 
			WHERE type='table' AND name='knowledge_base'
		`).Scan(&tableName)
		assert.NoError(t, err)
		assert.Equal(t, "knowledge_base", tableName)
	})

	t.Run("RunMigrations on existing database", func(t *testing.T) {
		runner := migrations.NewMigrationRunner(dbPath)

		// Run migrations again - should be idempotent
		err := runner.RunMigrations()
		assert.NoError(t, err)

		// Verify version is still the same
		version, dirty, err := runner.GetVersion()
		assert.NoError(t, err)
		assert.False(t, dirty)
		assert.Greater(t, version, uint(0))
	})
}

func TestMigrationRunner_InvalidPath(t *testing.T) {
	// Test with invalid database path
	runner := migrations.NewMigrationRunner("/invalid/path/to/database.db")

	err := runner.RunMigrations()
	assert.Error(t, err)
}

func TestGetVersion_NoMigrations(t *testing.T) {
	// Create a temporary directory for the test database
	tempDir, err := os.MkdirTemp("", "migrations-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "empty.db")

	// Create empty database
	db, err := sql.Open("sqlite3", dbPath)
	require.NoError(t, err)
	db.Close()

	runner := migrations.NewMigrationRunner(dbPath)

	// Get version on empty database
	version, dirty, err := runner.GetVersion()
	assert.NoError(t, err)
	assert.False(t, dirty)
	assert.Equal(t, uint(0), version)
}

func TestMigrationRunner_Concurrent(t *testing.T) {
	// Create a temporary directory for the test database
	tempDir, err := os.MkdirTemp("", "migrations-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "concurrent.db")

	// Test concurrent migration runs
	done := make(chan bool)
	errors := make(chan error, 2)

	for i := 0; i < 2; i++ {
		go func() {
			runner := migrations.NewMigrationRunner(dbPath)
			err := runner.RunMigrations()
			if err != nil {
				errors <- err
			}
			done <- true
		}()
	}

	// Wait for both goroutines to complete
	<-done
	<-done

	// Check for any errors
	select {
	case err := <-errors:
		t.Errorf("Concurrent migration failed: %v", err)
	default:
		// No errors, migrations completed successfully
	}

	// Verify migrations ran successfully
	runner := migrations.NewMigrationRunner(dbPath)
	version, dirty, err := runner.GetVersion()
	assert.NoError(t, err)
	assert.False(t, dirty)
	assert.Greater(t, version, uint(0))
}
