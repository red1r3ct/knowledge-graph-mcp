package ncruces

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/golang-migrate/migrate/v4/database"
	_ "github.com/ncruces/go-sqlite3/driver"
)

// Driver implements the database.Driver interface for ncruces/go-sqlite3
type Driver struct {
	db     *sql.DB
	config *Config
	lock   *LockManager
}

// init registers the driver with go-migrate
func init() {
	database.Register("sqlite3", &Driver{})
}

// Open opens a new database connection
func (d *Driver) Open(url string) (database.Driver, error) {
	config, err := ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// The ncruces driver is already registered via the import

	// Open database connection
	db, err := sql.Open("sqlite3", config.DatabaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Initialize database schema
	if err := initializeDatabase(db, config); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	driver := &Driver{
		db:     db,
		config: config,
		lock:   NewLockManager(db, config),
	}

	return driver, nil
}

// Close closes the database connection
func (d *Driver) Close() error {
	if d.db == nil {
		return nil
	}
	return d.db.Close()
}

// Lock acquires a database lock for migrations
func (d *Driver) Lock() error {
	if d.db == nil {
		return ErrDatabaseClosed
	}
	return d.lock.Acquire(context.Background())
}

// Unlock releases the database lock
func (d *Driver) Unlock() error {
	if d.db == nil {
		return ErrDatabaseClosed
	}
	return d.lock.Release(context.Background())
}

// Run executes a migration
func (d *Driver) Run(migration io.Reader) error {
	if d.db == nil {
		return ErrDatabaseClosed
	}

	// Read migration content
	content, err := io.ReadAll(migration)
	if err != nil {
		return fmt.Errorf("failed to read migration: %w", err)
	}

	migrationStr := strings.TrimSpace(string(content))
	if migrationStr == "" {
		return nil // Empty migration is valid
	}

	// Execute migration
	if d.config.NoTxWrap {
		return d.executeWithoutTransaction(migrationStr)
	}
	return d.executeWithTransaction(migrationStr)
}

// executeWithTransaction executes migration within a transaction
func (d *Driver) executeWithTransaction(migration string) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Set transaction mode if specified
	if d.config.TxMode != "DEFERRED" {
		if _, err := tx.Exec(fmt.Sprintf("PRAGMA %s", d.config.TxMode)); err != nil {
			return fmt.Errorf("failed to set transaction mode: %w", err)
		}
	}

	// Execute migration
	if _, err := tx.Exec(migration); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// executeWithoutTransaction executes migration without transaction wrapping
func (d *Driver) executeWithoutTransaction(migration string) error {
	if _, err := d.db.Exec(migration); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}
	return nil
}

// SetVersion sets the migration version
func (d *Driver) SetVersion(version int, dirty bool) error {
	if d.db == nil {
		return ErrDatabaseClosed
	}

	if dirty {
		// When dirty=true: Insert a new version record (don't clear existing versions)
		query := fmt.Sprintf(`
			INSERT INTO %s (version, dirty)
			VALUES (?, ?)`, d.config.MigrationsTable)
		
		_, err := d.db.Exec(query, version, dirty)
		if err != nil {
			return fmt.Errorf("failed to set version (dirty=true): %w", err)
		}
	} else {
		// When dirty=false: Clear all versions from the table and insert the specified version
		tx, err := d.db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback()

		// Clear all existing versions
		clearQuery := fmt.Sprintf("DELETE FROM %s", d.config.MigrationsTable)
		if _, err := tx.Exec(clearQuery); err != nil {
			return fmt.Errorf("failed to clear versions: %w", err)
		}

		// Insert the specified version
		insertQuery := fmt.Sprintf(`
			INSERT INTO %s (version, dirty)
			VALUES (?, ?)`, d.config.MigrationsTable)
		
		if _, err := tx.Exec(insertQuery, version, dirty); err != nil {
			return fmt.Errorf("failed to insert version: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	}

	return nil
}

// Version returns the current migration version
func (d *Driver) Version() (int, bool, error) {
	if d.db == nil {
		return 0, false, ErrDatabaseClosed
	}

	var version int
	var dirty bool

	query := fmt.Sprintf("SELECT version, dirty FROM %s ORDER BY version DESC LIMIT 1", d.config.MigrationsTable)
	err := d.db.QueryRow(query).Scan(&version, &dirty)

	if err == sql.ErrNoRows {
		return -1, false, nil // No migrations applied
	}
	if err != nil {
		// Handle case where table doesn't exist
		if strings.Contains(err.Error(), "no such table") {
			return -1, false, nil // No migrations applied
		}
		return 0, false, fmt.Errorf("failed to get version: %w", err)
	}

	return version, dirty, nil
}

// Drop drops all tables in the database
func (d *Driver) Drop() error {
	if d.db == nil {
		return ErrDatabaseClosed
	}

	// Get list of all tables
	rows, err := d.db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'")
	if err != nil {
		return fmt.Errorf("failed to get table list: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, table)
	}

	// Drop all tables
	for _, table := range tables {
		query := fmt.Sprintf("DROP TABLE IF EXISTS %s", strconv.Quote(table))
		if _, err := d.db.Exec(query); err != nil {
			// Ignore errors for non-existent tables during drop
			if strings.Contains(err.Error(), "no such table") {
				continue
			}
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	return nil
}

// WithInstance creates a new driver instance with an existing database connection
func (d *Driver) WithInstance(instance interface{}, config *Config) (database.Driver, error) {
	db, ok := instance.(*sql.DB)
	if !ok {
		return nil, fmt.Errorf("expected *sql.DB instance, got %T", instance)
	}

	driverConfig := DefaultConfig()
	if config != nil {
		if config.MigrationsTable != "" {
			driverConfig.MigrationsTable = config.MigrationsTable
		}
	}

	// Initialize database schema
	if err := initializeDatabase(db, driverConfig); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return &Driver{
		db:     db,
		config: driverConfig,
		lock:   NewLockManager(db, driverConfig),
	}, nil
}

// initializeDatabase initializes the database schema
func initializeDatabase(db *sql.DB, config *Config) error {
	// Create migrations table
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			version INTEGER PRIMARY KEY,
			dirty BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`, config.MigrationsTable)

	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Create lock table
	query = fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s_lock (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			locked BOOLEAN NOT NULL DEFAULT FALSE,
			owner TEXT,
			acquired_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			CHECK (id = 1)
		)`, config.MigrationsTable)

	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("failed to create lock table: %w", err)
	}

	// Initialize lock row
	query = fmt.Sprintf(`
		INSERT OR IGNORE INTO %s_lock (id, locked, owner, acquired_at)
		VALUES (1, FALSE, '', CURRENT_TIMESTAMP)`, config.MigrationsTable)
	
	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("failed to initialize lock row: %w", err)
	}

	return nil
}
