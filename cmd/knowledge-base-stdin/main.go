package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/server"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"

	"github.com/red1r3ct/knowledge-graph-mcp/internal/knowledgebase/mcp"
	"github.com/red1r3ct/knowledge-graph-mcp/internal/knowledgebase/sqlite"
	"github.com/red1r3ct/knowledge-graph-mcp/internal/migrations"
)

const (
	defaultDBPath = "knowledge_base.db"
)

func main() {
	// Parse command line arguments
	var dbPath string
	flag.StringVar(&dbPath, "db", defaultDBPath, "Path to SQLite database file")
	flag.StringVar(&dbPath, "database", defaultDBPath, "Path to SQLite database file (shorthand)")
	flag.Parse()

	// Validate database path
	if dbPath == "" {
		fmt.Fprintf(os.Stderr, "Error: database path cannot be empty\n")
		flag.Usage()
		os.Exit(1)
	}

	// Check if file exists and is accessible
	if _, err := os.Stat(dbPath); err != nil && !os.IsNotExist(err) {
		log.Fatalf("Failed to access database file: %v", err)
	}

	// Run migrations before initializing storage
	if err := runMigrations(dbPath); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize storage
	storage, err := sqlite.NewStorage(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer storage.Close()

	// Create MCP server
	s := server.NewMCPServer(
		"Knowledge Base MCP Server",
		"1.0.0",
	)

	// Register all tools using the mcp package
	if err := mcp.RegisterTools(s, storage); err != nil {
		log.Fatalf("Failed to register tools: %v", err)
	}

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func runMigrations(dbPath string) error {
	migrationRunner := migrations.NewMigrationRunner(dbPath)
	return migrationRunner.RunMigrations()
}
