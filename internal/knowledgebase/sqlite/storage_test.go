package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/red1r3ct/knowledge-graph-mcp/internal/knowledgebase"
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
	err = runTestMigrations(storage.db)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		tests := []struct {
			name     string
			req      knowledgebase.CreateRequest
			wantErr  bool
			validate func(t *testing.T, kb *knowledgebase.KnowledgeBase)
		}{
			{
				name: "create with all fields",
				req: knowledgebase.CreateRequest{
					Name:        "Test Knowledge Base",
					Description: strPtr("Test description"),
					Tags:        []string{"tag1", "tag2"},
				},
				wantErr: false,
				validate: func(t *testing.T, kb *knowledgebase.KnowledgeBase) {
					assert.Equal(t, "Test Knowledge Base", kb.Name)
					assert.Equal(t, "Test description", *kb.Description)
					assert.Equal(t, []string{"tag1", "tag2"}, kb.Tags)
					assert.False(t, kb.CreatedAt.IsZero())
					assert.False(t, kb.UpdatedAt.IsZero())
				},
			},
			{
				name: "create with required fields only",
				req: knowledgebase.CreateRequest{
					Name: "Minimal Knowledge Base",
				},
				wantErr: false,
				validate: func(t *testing.T, kb *knowledgebase.KnowledgeBase) {
					assert.Equal(t, "Minimal Knowledge Base", kb.Name)
					assert.Nil(t, kb.Description)
					assert.Empty(t, kb.Tags)
				},
			},
			{
				name: "create with empty name",
				req: knowledgebase.CreateRequest{
					Name: "",
				},
				wantErr: false, // SQLite allows empty strings, this is expected behavior
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				kb, err := storage.Create(ctx, tt.req)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}

				require.NoError(t, err)
				assert.NotZero(t, kb.ID)
				if tt.validate != nil {
					tt.validate(t, kb)
				}
			})
		}
	})

	t.Run("Get", func(t *testing.T) {
		// Create test data
		kb, err := storage.Create(ctx, knowledgebase.CreateRequest{
			Name:        "Test Get",
			Description: strPtr("Test description"),
			Tags:        []string{"test"},
		})
		require.NoError(t, err)

		tests := []struct {
			name    string
			id      int64
			wantErr bool
			wantNil bool
		}{
			{
				name:    "get existing",
				id:      kb.ID,
				wantErr: false,
				wantNil: false,
			},
			{
				name:    "get non-existing",
				id:      99999,
				wantErr: true,
				wantNil: true,
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
			})
		}
	})

	t.Run("Update", func(t *testing.T) {
		// Create test data
		kb, err := storage.Create(ctx, knowledgebase.CreateRequest{
			Name:        "Original Name",
			Description: strPtr("Original description"),
			Tags:        []string{"original"},
		})
		require.NoError(t, err)

		tests := []struct {
			name     string
			id       int64
			req      knowledgebase.UpdateRequest
			wantErr  bool
			validate func(t *testing.T, kb *knowledgebase.KnowledgeBase)
		}{
			{
				name: "update name only",
				id:   kb.ID,
				req: knowledgebase.UpdateRequest{
					Name: strPtr("Updated Name"),
				},
				wantErr: false,
				validate: func(t *testing.T, kb *knowledgebase.KnowledgeBase) {
					assert.Equal(t, "Updated Name", kb.Name)
					assert.Equal(t, "Original description", *kb.Description)
					assert.Equal(t, []string{"original"}, kb.Tags)
				},
			},
			{
				name: "update all fields",
				id:   kb.ID,
				req: knowledgebase.UpdateRequest{
					Name:        strPtr("Fully Updated"),
					Description: strPtr("New description"),
					Tags:        []string{"new", "tags"},
				},
				wantErr: false,
				validate: func(t *testing.T, kb *knowledgebase.KnowledgeBase) {
					assert.Equal(t, "Fully Updated", kb.Name)
					assert.Equal(t, "New description", *kb.Description)
					assert.Equal(t, []string{"new", "tags"}, kb.Tags)
				},
			},
			{
				name: "update non-existing",
				id:   99999,
				req: knowledgebase.UpdateRequest{
					Name: strPtr("Should fail"),
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
		kb, err := storage.Create(ctx, knowledgebase.CreateRequest{
			Name: "To Delete",
		})
		require.NoError(t, err)

		tests := []struct {
			name    string
			id      int64
			wantErr bool
		}{
			{
				name:    "delete existing",
				id:      kb.ID,
				wantErr: false,
			},
			{
				name:    "delete non-existing",
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
		// Clean up existing data
		_, err := storage.db.Exec("DELETE FROM knowledge_base")
		require.NoError(t, err)

		// Create test data
		for i := 0; i < 5; i++ {
			_, err := storage.Create(ctx, knowledgebase.CreateRequest{
				Name:        fmt.Sprintf("Test KB %d", i+1),
				Description: strPtr(fmt.Sprintf("Description %d", i+1)),
				Tags:        []string{fmt.Sprintf("tag%d", i+1)},
			})
			require.NoError(t, err)
		}

		// Wait a bit to ensure different timestamps
		time.Sleep(10 * time.Millisecond)

		// Create one more with specific tags
		_, err = storage.Create(ctx, knowledgebase.CreateRequest{
			Name:        "Special KB",
			Description: strPtr("Special description"),
			Tags:        []string{"special", "test"},
		})
		require.NoError(t, err)

		tests := []struct {
			name      string
			req       knowledgebase.ListRequest
			wantTotal int
			wantItems int
			wantErr   bool
		}{
			{
				name: "list all",
				req: knowledgebase.ListRequest{
					Limit:  10,
					Offset: 0,
				},
				wantTotal: 6,
				wantItems: 6,
				wantErr:   false,
			},
			{
				name: "list with limit",
				req: knowledgebase.ListRequest{
					Limit:  3,
					Offset: 0,
				},
				wantTotal: 6,
				wantItems: 3,
				wantErr:   false,
			},
			{
				name: "list with offset",
				req: knowledgebase.ListRequest{
					Limit:  10,
					Offset: 3,
				},
				wantTotal: 6,
				wantItems: 3,
				wantErr:   false,
			},
			{
				name: "list with search",
				req: knowledgebase.ListRequest{
					Limit:  10,
					Offset: 0,
					Search: "Special",
				},
				wantTotal: 1,
				wantItems: 1,
				wantErr:   false,
			},
			{
				name: "list with tags",
				req: knowledgebase.ListRequest{
					Limit:  10,
					Offset: 0,
					Tags:   []string{"special"},
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

				// Verify order (newest first)
				for i := 1; i < len(response.Items); i++ {
					assert.True(t, response.Items[i-1].CreatedAt.After(response.Items[i].CreatedAt) ||
						response.Items[i-1].CreatedAt.Equal(response.Items[i].CreatedAt))
				}
			})
		}
	})
}

func runTestMigrations(db *sql.DB) error {
	// Create table
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
		return fmt.Errorf("failed to create table: %w", err)
	}

	// Create indexes
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_knowledge_base_name ON knowledge_base(name)`)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_knowledge_base_created_at ON knowledge_base(created_at DESC)`)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	// Create trigger
	_, err = db.Exec(`
		CREATE TRIGGER IF NOT EXISTS update_knowledge_base_updated_at 
		AFTER UPDATE ON knowledge_base
		FOR EACH ROW
		BEGIN
			UPDATE knowledge_base SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
		END
	`)
	if err != nil {
		return fmt.Errorf("failed to create trigger: %w", err)
	}

	return nil
}

func strPtr(s string) *string {
	return &s
}
