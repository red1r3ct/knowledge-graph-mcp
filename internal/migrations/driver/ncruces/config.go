package ncruces

import (
	"fmt"
	"net/url"
	"strconv"
)

// Config holds the configuration for the ncruces SQLite driver
type Config struct {
	DatabaseName    string
	MigrationsTable string
	NoTxWrap        bool
	TxMode          string // "DEFERRED", "IMMEDIATE", "EXCLUSIVE"
}

// DefaultConfig returns a new Config with default values
func DefaultConfig() *Config {
	return &Config{
		MigrationsTable: "schema_migrations",
		NoTxWrap:        false,
		TxMode:          "DEFERRED",
	}
}

// ParseConfig parses configuration from a URL string
func ParseConfig(rawurl string) (*Config, error) {
	config := DefaultConfig()

	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if u.Scheme != "sqlite3" && u.Scheme != "ncruces-sqlite3" {
		return nil, fmt.Errorf("invalid sqlite3 scheme: %s", u.Scheme)
	}

	config.DatabaseName = u.Host + u.Path
	if config.DatabaseName == "" && u.Host == "" {
		return nil, fmt.Errorf("empty database path")
	}

	// Parse query parameters
	if u.RawQuery != "" {
		values, err := url.ParseQuery(u.RawQuery)
		if err != nil {
			return nil, fmt.Errorf("invalid query parameters: %w", err)
		}

		if table := values.Get("x-migrations-table"); table != "" {
			config.MigrationsTable = table
		}

		if noTxWrap := values.Get("x-no-tx-wrap"); noTxWrap != "" {
			config.NoTxWrap, err = strconv.ParseBool(noTxWrap)
			if err != nil {
				return nil, fmt.Errorf("invalid x-no-tx-wrap value: %w", err)
			}
		}

		if txMode := values.Get("x-tx-mode"); txMode != "" {
			switch txMode {
			case "DEFERRED", "IMMEDIATE", "EXCLUSIVE":
				config.TxMode = txMode
			default:
				return nil, fmt.Errorf("invalid transaction mode: %s", txMode)
			}
		}
	}

	return config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.DatabaseName == "" {
		return fmt.Errorf("database name is required")
	}
	if c.MigrationsTable == "" {
		return fmt.Errorf("migrations table name is required")
	}
	if c.TxMode != "DEFERRED" && c.TxMode != "IMMEDIATE" && c.TxMode != "EXCLUSIVE" {
		return fmt.Errorf("invalid transaction mode: %s", c.TxMode)
	}
	return nil
}