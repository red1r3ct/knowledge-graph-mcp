package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/red1r3ct/knowledge-graph-mcp/internal/connection"
	"github.com/red1r3ct/knowledge-graph-mcp/internal/migrations"
)

func TestStorage(t *testing.T) {
	// Create temporary database file
	tempFile, err := os.CreateTemp("", "test-*.db")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Initialize storage
	storage, err := NewStorage(tempFile.Name())
	require.NoError(t, err)
	defer storage.Close()

	// Run migrations
	migrationRunner := migrations.NewMigrationRunner(tempFile.Name())
	err = migrationRunner.RunMigrations()
	require.NoError(t, err)

	ctx := context.Background()

	// Create test notes for foreign key relationships
	note1ID, note2ID, note3ID := createTestNotes(t, storage.db)

	t.Run("Create", func(t *testing.T) {
		tests := []struct {
			name     string
			req      connection.CreateConnectionRequest
			wantErr  bool
			validate func(t *testing.T, conn *connection.Connection)
		}{
			{
				name: "create with all fields",
				req: connection.CreateConnectionRequest{
					FromNoteID:  note1ID,
					ToNoteID:    note2ID,
					Type:        "relates_to",
					Description: strPtr("Test connection description"),
					Strength:    8,
					Metadata:    map[string]interface{}{"key": "value"},
				},
				wantErr: false,
				validate: func(t *testing.T, conn *connection.Connection) {
					assert.Equal(t, note1ID, conn.FromNoteID)
					assert.Equal(t, note2ID, conn.ToNoteID)
					assert.Equal(t, "relates_to", conn.Type)
					assert.Equal(t, "Test connection description", *conn.Description)
					assert.Equal(t, 8, conn.Strength)
					assert.Equal(t, "value", conn.Metadata["key"])
					assert.False(t, conn.CreatedAt.IsZero())
					assert.False(t, conn.UpdatedAt.IsZero())
				},
			},
			{
				name: "create with required fields only",
				req: connection.CreateConnectionRequest{
					FromNoteID: note1ID,
					ToNoteID:   note3ID,
					Type:       "references",
					Strength:   5,
				},
				wantErr: false,
				validate: func(t *testing.T, conn *connection.Connection) {
					assert.Equal(t, note1ID, conn.FromNoteID)
					assert.Equal(t, note3ID, conn.ToNoteID)
					assert.Equal(t, "references", conn.Type)
					assert.Nil(t, conn.Description)
					assert.Equal(t, 5, conn.Strength)
					assert.Empty(t, conn.Metadata)
				},
			},
			{
				name: "create with invalid connection type",
				req: connection.CreateConnectionRequest{
					FromNoteID: note1ID,
					ToNoteID:   note2ID,
					Type:       "invalid_type",
					Strength:   5,
				},
				wantErr: true,
			},
			{
				name: "create with invalid strength (too low)",
				req: connection.CreateConnectionRequest{
					FromNoteID: note1ID,
					ToNoteID:   note2ID,
					Type:       "relates_to",
					Strength:   0,
				},
				wantErr: true,
			},
			{
				name: "create with invalid strength (too high)",
				req: connection.CreateConnectionRequest{
					FromNoteID: note1ID,
					ToNoteID:   note2ID,
					Type:       "relates_to",
					Strength:   11,
				},
				wantErr: true,
			},
			{
				name: "create with non-existing from_note_id",
				req: connection.CreateConnectionRequest{
					FromNoteID: 99999,
					ToNoteID:   note2ID,
					Type:       "relates_to",
					Strength:   5,
				},
				wantErr: true,
			},
			{
				name: "create with non-existing to_note_id",
				req: connection.CreateConnectionRequest{
					FromNoteID: note1ID,
					ToNoteID:   99999,
					Type:       "relates_to",
					Strength:   5,
				},
				wantErr: true,
			},
			{
				name: "create self-connection",
				req: connection.CreateConnectionRequest{
					FromNoteID: note1ID,
					ToNoteID:   note1ID,
					Type:       "relates_to",
					Strength:   5,
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				conn, err := storage.Create(ctx, tt.req)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}

				require.NoError(t, err)
				assert.NotZero(t, conn.ID)
				if tt.validate != nil {
					tt.validate(t, conn)
				}
			})
		}
	})

	t.Run("Get", func(t *testing.T) {
		// Create test data
		conn, err := storage.Create(ctx, connection.CreateConnectionRequest{
			FromNoteID:  note1ID,
			ToNoteID:    note2ID,
			Type:        "supports",
			Description: strPtr("Test get connection"),
			Strength:    7,
			Metadata:    map[string]interface{}{"test": "get"},
		})
		require.NoError(t, err)

		tests := []struct {
			name    string
			id      int64
			wantErr bool
		}{
			{
				name:    "get existing connection",
				id:      conn.ID,
				wantErr: false,
			},
			{
				name:    "get non-existing connection",
				id:      99999,
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := storage.Get(ctx, tt.id)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}

				require.NoError(t, err)
				assert.Equal(t, tt.id, got.ID)
				assert.Equal(t, conn.FromNoteID, got.FromNoteID)
				assert.Equal(t, conn.ToNoteID, got.ToNoteID)
				assert.Equal(t, conn.Type, got.Type)
			})
		}
	})

	t.Run("Update", func(t *testing.T) {
		// Create test data
		conn, err := storage.Create(ctx, connection.CreateConnectionRequest{
			FromNoteID:  note2ID,
			ToNoteID:    note3ID,
			Type:        "influences",
			Description: strPtr("Original description"),
			Strength:    6,
			Metadata:    map[string]interface{}{"original": "value"},
		})
		require.NoError(t, err)

		tests := []struct {
			name     string
			id       int64
			req      connection.UpdateConnectionRequest
			wantErr  bool
			validate func(t *testing.T, conn *connection.Connection)
		}{
			{
				name: "update type only",
				id:   conn.ID,
				req: connection.UpdateConnectionRequest{
					Type: strPtr("depends_on"),
				},
				wantErr: false,
				validate: func(t *testing.T, conn *connection.Connection) {
					assert.Equal(t, "depends_on", conn.Type)
					assert.Equal(t, "Original description", *conn.Description)
					assert.Equal(t, 6, conn.Strength)
				},
			},
			{
				name: "update all fields",
				id:   conn.ID,
				req: connection.UpdateConnectionRequest{
					Type:        strPtr("contradicts"),
					Description: strPtr("Updated description"),
					Strength:    intPtr(9),
					Metadata:    map[string]interface{}{"updated": "value"},
				},
				wantErr: false,
				validate: func(t *testing.T, conn *connection.Connection) {
					assert.Equal(t, "contradicts", conn.Type)
					assert.Equal(t, "Updated description", *conn.Description)
					assert.Equal(t, 9, conn.Strength)
					assert.Equal(t, "value", conn.Metadata["updated"])
				},
			},
			{
				name: "update with invalid type",
				id:   conn.ID,
				req: connection.UpdateConnectionRequest{
					Type: strPtr("invalid_type"),
				},
				wantErr: true,
			},
			{
				name: "update with invalid strength",
				id:   conn.ID,
				req: connection.UpdateConnectionRequest{
					Strength: intPtr(15),
				},
				wantErr: true,
			},
			{
				name: "update non-existing connection",
				id:   99999,
				req: connection.UpdateConnectionRequest{
					Type: strPtr("relates_to"),
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				updated, err := storage.Update(ctx, tt.id, tt.req)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}

				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, updated)
				}
			})
		}
	})

	t.Run("Delete", func(t *testing.T) {
		// Create test data
		conn, err := storage.Create(ctx, connection.CreateConnectionRequest{
			FromNoteID: note1ID,
			ToNoteID:   note3ID,
			Type:       "cites",
			Strength:   4,
		})
		require.NoError(t, err)

		tests := []struct {
			name    string
			id      int64
			wantErr bool
		}{
			{
				name:    "delete existing connection",
				id:      conn.ID,
				wantErr: false,
			},
			{
				name:    "delete non-existing connection",
				id:      99999,
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := storage.Delete(ctx, tt.id)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}

				require.NoError(t, err)

				// Verify deletion
				_, err = storage.Get(ctx, tt.id)
				assert.Error(t, err)
			})
		}
	})

	t.Run("List", func(t *testing.T) {
		// Clean up existing connections
		_, err := storage.db.Exec("DELETE FROM connections")
		require.NoError(t, err)

		// Create test data
		connections := []connection.CreateConnectionRequest{
			{FromNoteID: note1ID, ToNoteID: note2ID, Type: "relates_to", Strength: 5},
			{FromNoteID: note1ID, ToNoteID: note3ID, Type: "references", Strength: 7},
			{FromNoteID: note2ID, ToNoteID: note3ID, Type: "supports", Strength: 8},
			{FromNoteID: note2ID, ToNoteID: note1ID, Type: "contradicts", Strength: 3},
			{FromNoteID: note3ID, ToNoteID: note1ID, Type: "influences", Strength: 6},
		}

		for _, req := range connections {
			_, err := storage.Create(ctx, req)
			require.NoError(t, err)
		}

		tests := []struct {
			name      string
			req       connection.ListConnectionsRequest
			wantTotal int
			wantItems int
			wantErr   bool
		}{
			{
				name: "list all connections",
				req: connection.ListConnectionsRequest{
					Limit:  10,
					Offset: 0,
				},
				wantTotal: 5,
				wantItems: 5,
				wantErr:   false,
			},
			{
				name: "list with limit",
				req: connection.ListConnectionsRequest{
					Limit:  3,
					Offset: 0,
				},
				wantTotal: 5,
				wantItems: 3,
				wantErr:   false,
			},
			{
				name: "list with offset",
				req: connection.ListConnectionsRequest{
					Limit:  10,
					Offset: 3,
				},
				wantTotal: 5,
				wantItems: 2,
				wantErr:   false,
			},
			{
				name: "filter by from_note_id",
				req: connection.ListConnectionsRequest{
					Limit:      10,
					Offset:     0,
					FromNoteID: &note1ID,
				},
				wantTotal: 2,
				wantItems: 2,
				wantErr:   false,
			},
			{
				name: "filter by to_note_id",
				req: connection.ListConnectionsRequest{
					Limit:    10,
					Offset:   0,
					ToNoteID: &note3ID,
				},
				wantTotal: 2,
				wantItems: 2,
				wantErr:   false,
			},
			{
				name: "filter by type",
				req: connection.ListConnectionsRequest{
					Limit:  10,
					Offset: 0,
					Type:   strPtr("supports"),
				},
				wantTotal: 1,
				wantItems: 1,
				wantErr:   false,
			},
			{
				name: "filter by strength",
				req: connection.ListConnectionsRequest{
					Limit:    10,
					Offset:   0,
					Strength: intPtr(7),
				},
				wantTotal: 1,
				wantItems: 1,
				wantErr:   false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				response, err := storage.List(ctx, tt.req)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}

				require.NoError(t, err)
				assert.Equal(t, int64(tt.wantTotal), response.Total)
				assert.Len(t, response.Items, tt.wantItems)

				// Verify order (newest first by default)
				for i := 1; i < len(response.Items); i++ {
					assert.True(t, response.Items[i-1].CreatedAt.After(response.Items[i].CreatedAt) ||
						response.Items[i-1].CreatedAt.Equal(response.Items[i].CreatedAt))
				}
			})
		}
	})

	t.Run("GetNoteConnections", func(t *testing.T) {
		// Clean up existing connections
		_, err := storage.db.Exec("DELETE FROM connections")
		require.NoError(t, err)

		// Create test connections for note1
		outgoingConns := []connection.CreateConnectionRequest{
			{FromNoteID: note1ID, ToNoteID: note2ID, Type: "relates_to", Strength: 5},
			{FromNoteID: note1ID, ToNoteID: note3ID, Type: "references", Strength: 7},
		}
		incomingConns := []connection.CreateConnectionRequest{
			{FromNoteID: note2ID, ToNoteID: note1ID, Type: "supports", Strength: 8},
			{FromNoteID: note3ID, ToNoteID: note1ID, Type: "influences", Strength: 6},
		}

		for _, req := range append(outgoingConns, incomingConns...) {
			_, err := storage.Create(ctx, req)
			require.NoError(t, err)
		}

		tests := []struct {
			name            string
			req             connection.NoteConnectionsRequest
			wantOutgoing    int
			wantIncoming    int
			wantTotalCount  int64
			wantErr         bool
		}{
			{
				name: "get all connections for note1",
				req: connection.NoteConnectionsRequest{
					NoteID: note1ID,
					Limit:  10,
					Offset: 0,
				},
				wantOutgoing:   2,
				wantIncoming:   2,
				wantTotalCount: 4,
				wantErr:        false,
			},
			{
				name: "filter by type",
				req: connection.NoteConnectionsRequest{
					NoteID: note1ID,
					Type:   strPtr("relates_to"),
					Limit:  10,
					Offset: 0,
				},
				wantOutgoing:   1,
				wantIncoming:   0,
				wantTotalCount: 1,
				wantErr:        false,
			},
			{
				name: "filter by strength",
				req: connection.NoteConnectionsRequest{
					NoteID:   note1ID,
					Strength: intPtr(7),
					Limit:    10,
					Offset:   0,
				},
				wantOutgoing:   1,
				wantIncoming:   0,
				wantTotalCount: 1,
				wantErr:        false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				response, err := storage.GetNoteConnections(ctx, tt.req)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}

				require.NoError(t, err)
				assert.Equal(t, note1ID, response.NoteID)
				assert.Len(t, response.Outgoing, tt.wantOutgoing)
				assert.Len(t, response.Incoming, tt.wantIncoming)
				assert.Equal(t, tt.wantTotalCount, response.TotalCount)
				assert.NotEmpty(t, response.TypesCount)
			})
		}
	})

	t.Run("GetConnectionsByType", func(t *testing.T) {
		// Clean up existing connections
		_, err := storage.db.Exec("DELETE FROM connections")
		require.NoError(t, err)

		// Create test connections
		connections := []connection.CreateConnectionRequest{
			{FromNoteID: note1ID, ToNoteID: note2ID, Type: "relates_to", Strength: 5},
			{FromNoteID: note1ID, ToNoteID: note3ID, Type: "relates_to", Strength: 7},
			{FromNoteID: note2ID, ToNoteID: note3ID, Type: "supports", Strength: 8},
		}

		for _, req := range connections {
			_, err := storage.Create(ctx, req)
			require.NoError(t, err)
		}

		tests := []struct {
			name           string
			connectionType string
			req            connection.ListConnectionsRequest
			wantTotal      int
			wantItems      int
			wantErr        bool
		}{
			{
				name:           "get relates_to connections",
				connectionType: "relates_to",
				req: connection.ListConnectionsRequest{
					Limit:  10,
					Offset: 0,
				},
				wantTotal: 2,
				wantItems: 2,
				wantErr:   false,
			},
			{
				name:           "get supports connections",
				connectionType: "supports",
				req: connection.ListConnectionsRequest{
					Limit:  10,
					Offset: 0,
				},
				wantTotal: 1,
				wantItems: 1,
				wantErr:   false,
			},
			{
				name:           "get non-existing type",
				connectionType: "non_existing",
				req: connection.ListConnectionsRequest{
					Limit:  10,
					Offset: 0,
				},
				wantTotal: 0,
				wantItems: 0,
				wantErr:   false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				response, err := storage.GetConnectionsByType(ctx, tt.connectionType, tt.req)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}

				require.NoError(t, err)
				assert.Equal(t, int64(tt.wantTotal), response.Total)
				assert.Len(t, response.Items, tt.wantItems)

				// Verify all connections have the correct type
				for _, conn := range response.Items {
					assert.Equal(t, tt.connectionType, conn.Type)
				}
			})
		}
	})

	t.Run("GetBidirectionalConnections", func(t *testing.T) {
		// Clean up existing connections
		_, err := storage.db.Exec("DELETE FROM connections")
		require.NoError(t, err)

		// Create bidirectional connections
		connections := []connection.CreateConnectionRequest{
			{FromNoteID: note1ID, ToNoteID: note2ID, Type: "relates_to", Strength: 5},
			{FromNoteID: note2ID, ToNoteID: note1ID, Type: "relates_to", Strength: 5},
			{FromNoteID: note1ID, ToNoteID: note3ID, Type: "references", Strength: 7},
		}

		for _, req := range connections {
			_, err := storage.Create(ctx, req)
			require.NoError(t, err)
		}

		response, err := storage.GetBidirectionalConnections(ctx, note1ID)
		require.NoError(t, err)

		assert.Equal(t, note1ID, response.NoteID)
		assert.Len(t, response.Outgoing, 2) // note1 -> note2, note1 -> note3
		assert.Len(t, response.Incoming, 1) // note2 -> note1
		assert.Equal(t, int64(3), response.TotalCount)
	})

	t.Run("GetConnectionStats", func(t *testing.T) {
		// Clean up existing connections
		_, err := storage.db.Exec("DELETE FROM connections")
		require.NoError(t, err)

		// Create test connections with various types and strengths
		connections := []connection.CreateConnectionRequest{
			{FromNoteID: note1ID, ToNoteID: note2ID, Type: "relates_to", Strength: 5},
			{FromNoteID: note1ID, ToNoteID: note3ID, Type: "relates_to", Strength: 7},
			{FromNoteID: note2ID, ToNoteID: note3ID, Type: "supports", Strength: 8},
			{FromNoteID: note2ID, ToNoteID: note1ID, Type: "contradicts", Strength: 3},
			{FromNoteID: note3ID, ToNoteID: note1ID, Type: "influences", Strength: 6},
			{FromNoteID: note3ID, ToNoteID: note2ID, Type: "supports", Strength: 9},
		}

		for _, req := range connections {
			_, err := storage.Create(ctx, req)
			require.NoError(t, err)
		}

		tests := []struct {
			name     string
			validate func(t *testing.T, stats *connection.ConnectionStats)
		}{
			{
				name: "get connection statistics",
				validate: func(t *testing.T, stats *connection.ConnectionStats) {
					// Verify total connections
					assert.Equal(t, int64(6), stats.TotalConnections)

					// Verify connections by type
					assert.Equal(t, int64(2), stats.ConnectionsByType["relates_to"])
					assert.Equal(t, int64(2), stats.ConnectionsByType["supports"])
					assert.Equal(t, int64(1), stats.ConnectionsByType["contradicts"])
					assert.Equal(t, int64(1), stats.ConnectionsByType["influences"])

					// Verify connections by strength
					assert.Equal(t, int64(1), stats.ConnectionsByStrength[3])
					assert.Equal(t, int64(1), stats.ConnectionsByStrength[5])
					assert.Equal(t, int64(1), stats.ConnectionsByStrength[6])
					assert.Equal(t, int64(1), stats.ConnectionsByStrength[7])
					assert.Equal(t, int64(1), stats.ConnectionsByStrength[8])
					assert.Equal(t, int64(1), stats.ConnectionsByStrength[9])

					// Verify most connected notes (should have at least some notes)
					assert.NotEmpty(t, stats.MostConnectedNotes)
					
					// Verify that note counts are correct
					for _, noteConn := range stats.MostConnectedNotes {
						assert.True(t, noteConn.TotalCount > 0)
						assert.Equal(t, noteConn.IncomingCount+noteConn.OutgoingCount, noteConn.TotalCount)
					}
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				stats, err := storage.GetConnectionStats(ctx)
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, stats)
				}
			})
		}
	})

	t.Run("FindConnectionPaths", func(t *testing.T) {
		// Clean up existing connections
		_, err := storage.db.Exec("DELETE FROM connections")
		require.NoError(t, err)

		// Create test connections for path finding
		connections := []connection.CreateConnectionRequest{
			{FromNoteID: note1ID, ToNoteID: note2ID, Type: "relates_to", Strength: 5},
			{FromNoteID: note1ID, ToNoteID: note2ID, Type: "supports", Strength: 8}, // Multiple connections between same notes
			{FromNoteID: note2ID, ToNoteID: note3ID, Type: "references", Strength: 7},
		}

		for _, req := range connections {
			_, err := storage.Create(ctx, req)
			require.NoError(t, err)
		}

		tests := []struct {
			name         string
			fromNoteID   int64
			toNoteID     int64
			maxDepth     int
			wantPaths    int
			wantErr      bool
			validate     func(t *testing.T, paths []connection.ConnectionPath)
		}{
			{
				name:       "find direct connections",
				fromNoteID: note1ID,
				toNoteID:   note2ID,
				maxDepth:   1,
				wantPaths:  2, // Two direct connections between note1 and note2
				wantErr:    false,
				validate: func(t *testing.T, paths []connection.ConnectionPath) {
					for _, path := range paths {
						assert.Equal(t, note1ID, path.FromNoteID)
						assert.Equal(t, note2ID, path.ToNoteID)
						assert.Equal(t, 1, path.Length)
						assert.Len(t, path.Path, 1)
						assert.True(t, path.Strength == 5 || path.Strength == 8)
					}
				},
			},
			{
				name:       "find no direct connection",
				fromNoteID: note1ID,
				toNoteID:   note3ID,
				maxDepth:   1,
				wantPaths:  0, // No direct connection between note1 and note3
				wantErr:    false,
				validate: func(t *testing.T, paths []connection.ConnectionPath) {
					assert.Empty(t, paths)
				},
			},
			{
				name:       "find connection with non-existing notes",
				fromNoteID: 99999,
				toNoteID:   note2ID,
				maxDepth:   1,
				wantPaths:  0,
				wantErr:    false,
				validate: func(t *testing.T, paths []connection.ConnectionPath) {
					assert.Empty(t, paths)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				paths, err := storage.FindConnectionPaths(ctx, tt.fromNoteID, tt.toNoteID, tt.maxDepth)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}

				require.NoError(t, err)
				assert.Len(t, paths, tt.wantPaths)
				if tt.validate != nil {
					tt.validate(t, paths)
				}
			})
		}
	})
}

func runTestMigrations(db *sql.DB) error {
	// Create knowledge_base table (required for foreign keys)
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS knowledge_base (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			tags TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create knowledge_base table: %w", err)
	}

	// Create notes table (required for foreign keys)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS notes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			type TEXT NOT NULL,
			tags TEXT,
			metadata TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create notes table: %w", err)
	}

	// Create connections table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS connections (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			from_note_id INTEGER NOT NULL,
			to_note_id INTEGER NOT NULL,
			type TEXT NOT NULL CHECK (type IN (
				'relates_to', 'references', 'supports', 'contradicts', 'influences',
				'depends_on', 'similar_to', 'part_of', 'cites', 'follows', 'precedes'
			)),
			description TEXT,
			strength INTEGER NOT NULL CHECK (strength >= 1 AND strength <= 10) DEFAULT 5,
			metadata TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (from_note_id) REFERENCES notes(id) ON DELETE CASCADE,
			FOREIGN KEY (to_note_id) REFERENCES notes(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create connections table: %w", err)
	}

	// Create indexes
	indexes := []string{
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_connections_unique ON connections(from_note_id, to_note_id, type)",
		"CREATE INDEX IF NOT EXISTS idx_connections_from_note_id ON connections(from_note_id)",
		"CREATE INDEX IF NOT EXISTS idx_connections_to_note_id ON connections(to_note_id)",
		"CREATE INDEX IF NOT EXISTS idx_connections_type ON connections(type)",
		"CREATE INDEX IF NOT EXISTS idx_connections_strength ON connections(strength)",
		"CREATE INDEX IF NOT EXISTS idx_connections_created_at ON connections(created_at DESC)",
	}

	for _, indexSQL := range indexes {
		if _, err := db.Exec(indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// Create triggers
	_, err = db.Exec(`
		CREATE TRIGGER IF NOT EXISTS update_connections_updated_at 
		AFTER UPDATE ON connections
		FOR EACH ROW
		BEGIN
			UPDATE connections SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END
	`)
	if err != nil {
		return fmt.Errorf("failed to create update trigger: %w", err)
	}

	_, err = db.Exec(`
		CREATE TRIGGER IF NOT EXISTS prevent_self_connection
		BEFORE INSERT ON connections
		FOR EACH ROW
		WHEN NEW.from_note_id = NEW.to_note_id
		BEGIN
			SELECT RAISE(ABORT, 'Self-connections are not allowed');
		END
	`)
	if err != nil {
		return fmt.Errorf("failed to create self-connection prevention trigger: %w", err)
	}

	return nil
}

func createTestNotes(t *testing.T, db *sql.DB) (note1ID, note2ID, note3ID int64) {
	// Create test notes
	notes := []struct {
		title   string
		content string
		noteType string
	}{
		{"Test Note 1", "Content of test note 1", "text"},
		{"Test Note 2", "Content of test note 2", "markdown"},
		{"Test Note 3", "Content of test note 3", "text"},
	}

	var ids []int64
	for _, note := range notes {
		result, err := db.Exec(
			"INSERT INTO notes (title, content, type, tags, metadata) VALUES (?, ?, ?, ?, ?)",
			note.title, note.content, note.noteType, "[]", "{}",
		)
		require.NoError(t, err)

		id, err := result.LastInsertId()
		require.NoError(t, err)
		ids = append(ids, id)
	}

	return ids[0], ids[1], ids[2]
}

func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}