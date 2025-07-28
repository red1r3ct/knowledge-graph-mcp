//go:build integration

package ncruces

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestSuite provides integration tests for the ncruces driver with go-migrate
type IntegrationTestSuite struct {
	suite.Suite
	tempDir    string
	dbPath     string
	migrations string
}

// TestIntegrationTestSuite runs the integration test suite
func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// SetupSuite creates test environment
func (s *IntegrationTestSuite) SetupSuite() {
	var err error
	s.tempDir, err = os.MkdirTemp("", "ncruces-integration-test-*")
	s.Require().NoError(err)

	s.dbPath = filepath.Join(s.tempDir, "test.db")
	s.migrations = filepath.Join(s.tempDir, "migrations")

	// Create migrations directory
	err = os.MkdirAll(s.migrations, 0755)
	s.Require().NoError(err)

	// Create test migration files
	s.createTestMigrations()
}

// TearDownSuite cleans up test environment
func (s *IntegrationTestSuite) TearDownSuite() {
	if s.tempDir != "" {
		os.RemoveAll(s.tempDir)
	}
}

// createTestMigrations creates test migration files
func (s *IntegrationTestSuite) createTestMigrations() {
	// 001_create_users_table.up.sql
	up1 := `CREATE TABLE users (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	err := os.WriteFile(filepath.Join(s.migrations, "001_create_users_table.up.sql"), []byte(up1), 0644)
	s.Require().NoError(err)

	// 001_create_users_table.down.sql
	down1 := "DROP TABLE IF EXISTS users;"
	err = os.WriteFile(filepath.Join(s.migrations, "001_create_users_table.down.sql"), []byte(down1), 0644)
	s.Require().NoError(err)

	// 002_create_posts_table.up.sql
	up2 := `CREATE TABLE posts (
		id INTEGER PRIMARY KEY,
		user_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		content TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);`
	err = os.WriteFile(filepath.Join(s.migrations, "002_create_posts_table.up.sql"), []byte(up2), 0644)
	s.Require().NoError(err)

	// 002_create_posts_table.down.sql
	down2 := "DROP TABLE IF EXISTS posts;"
	err = os.WriteFile(filepath.Join(s.migrations, "002_create_posts_table.down.sql"), []byte(down2), 0644)
	s.Require().NoError(err)

	// 003_add_index_to_users.up.sql
	up3 := "CREATE INDEX idx_users_email ON users(email);"
	err = os.WriteFile(filepath.Join(s.migrations, "003_add_index_to_users.up.sql"), []byte(up3), 0644)
	s.Require().NoError(err)

	// 003_add_index_to_users.down.sql
	down3 := "DROP INDEX IF EXISTS idx_users_email;"
	err = os.WriteFile(filepath.Join(s.migrations, "003_add_index_to_users.down.sql"), []byte(down3), 0644)
	s.Require().NoError(err)
}

// TestDriverRegistration tests that the driver is properly registered
func (s *IntegrationTestSuite) TestDriverRegistration() {
	driver := &Driver{}
	result, err := driver.Open("sqlite3://" + s.dbPath)
	s.Require().NoError(err)
	s.Require().NotNil(result)

	err = result.Close()
	s.Require().NoError(err)
}

// TestMigrationUp tests migrating up
func (s *IntegrationTestSuite) TestMigrationUp() {
	// Use a fresh database for this test
	dbPath := filepath.Join(s.tempDir, "test_up.db")
	m, err := migrate.New(
		fmt.Sprintf("file://%s", s.migrations),
		"sqlite3://"+dbPath,
	)
	s.Require().NoError(err)
	defer m.Close()

	// Migrate up to latest version
	err = m.Up()
	s.Require().NoError(err)

	// Verify version
	version, dirty, err := m.Version()
	s.Require().NoError(err)
	s.Require().False(dirty)
	s.Require().Equal(uint(3), version)

	// Verify tables exist
	db, err := sql.Open("sqlite3", dbPath)
	s.Require().NoError(err)
	defer db.Close()

	var exists bool
	err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name='users')").Scan(&exists)
	s.Require().NoError(err)
	s.Require().True(exists)

	err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name='posts')").Scan(&exists)
	s.Require().NoError(err)
	s.Require().True(exists)

	err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='index' AND name='idx_users_email')").Scan(&exists)
	s.Require().NoError(err)
	s.Require().True(exists)
}

// TestMigrationDown tests migrating down
func (s *IntegrationTestSuite) TestMigrationDown() {
	// Use a fresh database for this test
	dbPath := filepath.Join(s.tempDir, "test_down.db")
	m, err := migrate.New(
		fmt.Sprintf("file://%s", s.migrations),
		"sqlite3://"+dbPath,
	)
	s.Require().NoError(err)
	defer m.Close()

	// Migrate up first
	err = m.Up()
	s.Require().NoError(err)

	// Migrate down one step
	err = m.Steps(-1)
	s.Require().NoError(err)

	// Verify version - should be 2 after migrating down from 3
	version, dirty, err := m.Version()
	s.Require().NoError(err)
	s.Require().False(dirty)
	s.Require().Equal(uint(2), version)

	// Verify index was dropped
	db, err := sql.Open("sqlite3", dbPath)
	s.Require().NoError(err)
	defer db.Close()

	var exists bool
	err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='index' AND name='idx_users_email')").Scan(&exists)
	s.Require().NoError(err)
	s.Require().False(exists)
}

// TestMigrationToVersion tests migrating to a specific version
func (s *IntegrationTestSuite) TestMigrationToVersion() {
	// Use a fresh database for this test
	dbPath := filepath.Join(s.tempDir, "test_version.db")
	m, err := migrate.New(
		fmt.Sprintf("file://%s", s.migrations),
		"sqlite3://"+dbPath,
	)
	s.Require().NoError(err)
	defer m.Close()

	// Migrate to version 2
	err = m.Migrate(2)
	s.Require().NoError(err)

	version, dirty, err := m.Version()
	s.Require().NoError(err)
	s.Require().False(dirty)
	s.Require().Equal(uint(2), version)

	// Verify tables
	db, err := sql.Open("sqlite3", dbPath)
	s.Require().NoError(err)
	defer db.Close()

	var exists bool
	err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name='users')").Scan(&exists)
	s.Require().NoError(err)
	s.Require().True(exists)

	err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name='posts')").Scan(&exists)
	s.Require().NoError(err)
	s.Require().True(exists)

	err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='index' AND name='idx_users_email')").Scan(&exists)
	s.Require().NoError(err)
	s.Require().False(exists)
}

// TestForceVersion tests forcing a version
func (s *IntegrationTestSuite) TestForceVersion() {
	// Use a fresh database for this test
	dbPath := filepath.Join(s.tempDir, "test_force.db")
	m, err := migrate.New(
		fmt.Sprintf("file://%s", s.migrations),
		"sqlite3://"+dbPath,
	)
	s.Require().NoError(err)
	defer m.Close()

	// Migrate up first
	err = m.Up()
	s.Require().NoError(err)

	// Force version to 2
	err = m.Force(2)
	s.Require().NoError(err)

	version, dirty, err := m.Version()
	s.Require().NoError(err)
	s.Require().False(dirty)
	s.Require().Equal(uint(2), version)
}

// TestDropDatabase tests dropping all tables
func (s *IntegrationTestSuite) TestDropDatabase() {
	// Use a fresh database for this test
	dbPath := filepath.Join(s.tempDir, "test_drop.db")
	m, err := migrate.New(
		fmt.Sprintf("file://%s", s.migrations),
		"sqlite3://"+dbPath,
	)
	s.Require().NoError(err)
	defer m.Close()

	// Migrate up first
	err = m.Up()
	s.Require().NoError(err)

	// Drop everything
	err = m.Drop()
	s.Require().NoError(err)

	// Verify all tables are dropped
	db, err := sql.Open("sqlite3", dbPath)
	s.Require().NoError(err)
	defer db.Close()

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'").Scan(&count)
	s.Require().NoError(err)
	s.Require().Equal(0, count)
}

// TestConcurrentMigrations tests concurrent migration attempts
func (s *IntegrationTestSuite) TestConcurrentMigrations() {
	m, err := s.createMigrator()
	s.Require().NoError(err)
	defer m.Close()

	// Migrate up
	err = m.Up()
	s.Require().NoError(err)

	// Create another migrator instance
	m2, err := s.createMigrator()
	s.Require().NoError(err)
	defer m2.Close()

	// Both should see the same version
	version1, dirty1, err := m.Version()
	s.Require().NoError(err)
	s.Require().False(dirty1)

	version2, dirty2, err := m2.Version()
	s.Require().NoError(err)
	s.Require().False(dirty2)

	s.Require().Equal(version1, version2)
}

// createMigrator creates a new migrator instance
func (s *IntegrationTestSuite) createMigrator() (*migrate.Migrate, error) {
	return migrate.New(
		fmt.Sprintf("file://%s", s.migrations),
		"sqlite3://"+s.dbPath,
	)
}

// TestIntegration is a standalone integration test
func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tempDir, err := os.MkdirTemp("", "integration-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	migrations := filepath.Join(tempDir, "migrations")

	// Create migrations directory
	err = os.MkdirAll(migrations, 0755)
	require.NoError(t, err)

	// Create simple migration
	up := "CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT);"
	err = os.WriteFile(filepath.Join(migrations, "001_test.up.sql"), []byte(up), 0644)
	require.NoError(t, err)

	down := "DROP TABLE IF EXISTS test_table;"
	err = os.WriteFile(filepath.Join(migrations, "001_test.down.sql"), []byte(down), 0644)
	require.NoError(t, err)

	// Test migration
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrations),
		"sqlite3://"+dbPath,
	)
	require.NoError(t, err)
	defer m.Close()

	// Migrate up
	err = m.Up()
	require.NoError(t, err)

	// Verify
	version, dirty, err := m.Version()
	require.NoError(t, err)
	assert.False(t, dirty)
	assert.Equal(t, uint(1), version)

	// Verify table exists
	db, err := sql.Open("sqlite3", dbPath)
	require.NoError(t, err)
	defer db.Close()

	var exists bool
	err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name='test_table')").Scan(&exists)
	require.NoError(t, err)
	assert.True(t, exists)
}
