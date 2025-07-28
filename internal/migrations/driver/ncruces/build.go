//go:build sqlite_migrate

package ncruces

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"github.com/golang-migrate/migrate/v4/database"
)

// init registers the ncruces SQLite driver with go-migrate
func init() {
	database.Register("sqlite3", &Driver{})
}

// BuildURL constructs a database URL for the ncruces driver
func BuildURL(dbPath string, options ...Option) string {
	u := &url.URL{
		Scheme: "sqlite3",
		Path:   dbPath,
	}

	params := url.Values{}
	for _, opt := range options {
		opt(params)
	}

	if len(params) > 0 {
		u.RawQuery = params.Encode()
	}

	return u.String()
}

// Option configures the database URL
type Option func(params url.Values)

// WithMigrationsTable sets the migrations table name
func WithMigrationsTable(table string) Option {
	return func(params url.Values) {
		params.Set("x-migrations-table", table)
	}
}

// WithNoTxWrap disables transaction wrapping
func WithNoTxWrap() Option {
	return func(params url.Values) {
		params.Set("x-no-tx-wrap", "true")
	}
}

// WithTxMode sets the transaction mode
func WithTxMode(mode string) Option {
	return func(params url.Values) {
		params.Set("x-tx-mode", strings.ToUpper(mode))
	}
}

// RegisterDriver explicitly registers the ncruces driver
func RegisterDriver() {
	database.Register("sqlite3", &Driver{})
}

// MustRegisterDriver registers the driver and panics on error
func MustRegisterDriver() {
	RegisterDriver()
}

// OpenConnection opens a database connection using the ncruces driver
func OpenConnection(dbPath string, options ...Option) (*sql.DB, error) {
	url := BuildURL(dbPath, options...)
	return sql.Open("sqlite3", url)
}

// NewMigrator creates a new migrator instance with the ncruces driver
func NewMigrator(dbPath string, options ...Option) (*database.Driver, error) {
	url := BuildURL(dbPath, options...)
	
	driver := &Driver{}
	return driver.Open(url)
}