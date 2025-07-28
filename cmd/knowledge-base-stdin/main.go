package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/server"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"

	kbmcp "github.com/red1r3ct/knowledge-graph-mcp/internal/knowledgebase/mcp"
	kbstorage "github.com/red1r3ct/knowledge-graph-mcp/internal/knowledgebase/sqlite"
	"github.com/red1r3ct/knowledge-graph-mcp/internal/migrations"
	notemcp "github.com/red1r3ct/knowledge-graph-mcp/internal/note/mcp"
	notestorage "github.com/red1r3ct/knowledge-graph-mcp/internal/note/sqlite"
	connmcp "github.com/red1r3ct/knowledge-graph-mcp/internal/connection/mcp"
	connstorage "github.com/red1r3ct/knowledge-graph-mcp/internal/connection/sqlite"
)

const (
	defaultDBPath = "knowledge-base.db"
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

	// Initialize knowledgebase storage
	kbStorage, err := kbstorage.NewStorage(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize knowledgebase storage: %v", err)
	}
	defer kbStorage.Close()

	// Initialize note storage
	noteStorage, err := notestorage.NewStorage(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize note storage: %v", err)
	}
	defer noteStorage.Close()

	// Initialize connection storage
	connStorage, err := connstorage.NewStorage(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize connection storage: %v", err)
	}
	defer connStorage.Close()

	// Create MCP server
	s := server.NewMCPServer(
		"Knowledge Graph MCP Server",
		"1.0.0",
	)

	// Register all knowledgebase tools
	if err := kbmcp.RegisterTools(s, kbStorage); err != nil {
		log.Fatalf("Failed to register knowledgebase tools: %v", err)
	}

	// Register all note tools
	if err := notemcp.RegisterTools(s, noteStorage); err != nil {
		log.Fatalf("Failed to register note tools: %v", err)
	}

	// Register all connection tools
	if err := connmcp.RegisterTools(s, connStorage); err != nil {
		log.Fatalf("Failed to register connection tools: %v", err)
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
