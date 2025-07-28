package migrations

import (
	"fmt"
	// Import the ncruces SQLite driver for go-migrate

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	_ "github.com/red1r3ct/knowledge-graph-mcp/internal/migrations/driver/ncruces"
)

// MigrationRunner handles database migrations using golang-migrate
type MigrationRunner struct {
	dbPath string
}

// NewMigrationRunner creates a new migration runner instance
func NewMigrationRunner(dbPath string) *MigrationRunner {
	return &MigrationRunner{
		dbPath: dbPath,
	}
}

// RunMigrations runs all pending migrations up to the latest version
func (mr *MigrationRunner) RunMigrations() error {
	// Create database URL for SQLite
	dbURL := fmt.Sprintf("sqlite3://%s", mr.dbPath)

	// Create source driver from embedded filesystem
	sourceDriver, err := iofs.New(MigrationsFS, "sqlite")
	if err != nil {
		return fmt.Errorf("failed to create source driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, dbURL)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// GetVersion returns the current migration version
func (mr *MigrationRunner) GetVersion() (uint, bool, error) {
	// Create database URL for SQLite
	dbURL := fmt.Sprintf("sqlite3://%s", mr.dbPath)

	// Create source driver from embedded filesystem
	sourceDriver, err := iofs.New(MigrationsFS, "sqlite")
	if err != nil {
		return 0, false, fmt.Errorf("failed to create source driver: %w", err)
	}

	// Create migrate instance
	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, dbURL)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	version, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("failed to get version: %w", err)
	}

	return version, dirty, nil
}
