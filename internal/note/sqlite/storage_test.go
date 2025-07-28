package sqlite

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/red1r3ct/knowledge-graph-mcp/internal/migrations"
	"github.com/red1r3ct/knowledge-graph-mcp/internal/note"
)

func TestStorage(t *testing.T) {
	// Create temporary database file
	tempFile, err := os.CreateTemp("", "test-note-*.db")
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

	t.Run("Create", func(t *testing.T) {
		tests := []struct {
			name     string
			req      note.CreateNoteRequest
			wantErr  bool
			validate func(t *testing.T, n *note.Note)
		}{
			{
				name: "create with all fields",
				req: note.CreateNoteRequest{
					Title:    "Test Note",
					Content:  "This is a test note content",
					Type:     "text",
					Tags:     []string{"tag1", "tag2"},
					Metadata: map[string]interface{}{"key": "value", "number": 42},
				},
				wantErr: false,
				validate: func(t *testing.T, n *note.Note) {
					assert.Equal(t, "Test Note", n.Title)
					assert.Equal(t, "This is a test note content", n.Content)
					assert.Equal(t, "text", n.Type)
					assert.Equal(t, []string{"tag1", "tag2"}, n.Tags)
					assert.Equal(t, map[string]interface{}{"key": "value", "number": float64(42)}, n.Metadata)
					assert.False(t, n.CreatedAt.IsZero())
					assert.False(t, n.UpdatedAt.IsZero())
				},
			},
			{
				name: "create with required fields only",
				req: note.CreateNoteRequest{
					Title:   "Minimal Note",
					Content: "Minimal content",
					Type:    "markdown",
				},
				wantErr: false,
				validate: func(t *testing.T, n *note.Note) {
					assert.Equal(t, "Minimal Note", n.Title)
					assert.Equal(t, "Minimal content", n.Content)
					assert.Equal(t, "markdown", n.Type)
					assert.Empty(t, n.Tags)
					assert.Nil(t, n.Metadata)
					assert.False(t, n.CreatedAt.IsZero())
					assert.False(t, n.UpdatedAt.IsZero())
				},
			},
			{
				name: "create with empty title",
				req: note.CreateNoteRequest{
					Title:   "",
					Content: "Content with empty title",
					Type:    "text",
				},
				wantErr: false,
			},
			{
				name: "create with empty content",
				req: note.CreateNoteRequest{
					Title:   "Empty Content Note",
					Content: "",
					Type:    "text",
				},
				wantErr: false,
			},
			{
				name: "create with empty type",
				req: note.CreateNoteRequest{
					Title:   "Empty Type Note",
					Content: "Content",
					Type:    "",
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				n, err := storage.Create(ctx, tt.req)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}

				require.NoError(t, err)
				assert.NotZero(t, n.ID)
				if tt.validate != nil {
					tt.validate(t, n)
				}
			})
		}
	})

	t.Run("Get", func(t *testing.T) {
		// Create test data
		n, err := storage.Create(ctx, note.CreateNoteRequest{
			Title:    "Test Get",
			Content:  "Test content",
			Type:     "text",
			Tags:     []string{"test"},
			Metadata: map[string]interface{}{"test": true},
		})
		require.NoError(t, err)

		tests := []struct {
			name    string
			id      int64
			wantErr bool
		}{
			{
				name:    "get existing",
				id:      n.ID,
				wantErr: false,
			},
			{
				name:    "get non-existing",
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
			})
		}
	})

	t.Run("Update", func(t *testing.T) {
		// Create test data
		n, err := storage.Create(ctx, note.CreateNoteRequest{
			Title:    "Original Title",
			Content:  "Original content",
			Type:     "text",
			Tags:     []string{"original"},
			Metadata: map[string]interface{}{"original": true},
		})
		require.NoError(t, err)

		tests := []struct {
			name     string
			id       int64
			req      note.UpdateNoteRequest
			wantErr  bool
			validate func(t *testing.T, n *note.Note)
		}{
			{
				name: "update title only",
				id:   n.ID,
				req: note.UpdateNoteRequest{
					Title: strPtr("Updated Title"),
				},
				wantErr: false,
				validate: func(t *testing.T, n *note.Note) {
					assert.Equal(t, "Updated Title", n.Title)
					assert.Equal(t, "Original content", n.Content)
					assert.Equal(t, []string{"original"}, n.Tags)
					assert.Equal(t, map[string]interface{}{"original": true}, n.Metadata)
				},
			},
			{
				name: "update all fields",
				id:   n.ID,
				req: note.UpdateNoteRequest{
					Title:    strPtr("Fully Updated"),
					Content:  strPtr("New content"),
					Type:     strPtr("markdown"),
					Tags:     []string{"new", "tags"},
					Metadata: map[string]interface{}{"new": "value"},
				},
				wantErr: false,
				validate: func(t *testing.T, n *note.Note) {
					assert.Equal(t, "Fully Updated", n.Title)
					assert.Equal(t, "New content", n.Content)
					assert.Equal(t, "markdown", n.Type)
					assert.Equal(t, []string{"new", "tags"}, n.Tags)
					assert.Equal(t, map[string]interface{}{"new": "value"}, n.Metadata)
				},
			},
			{
				name: "update non-existing",
				id:   99999,
				req: note.UpdateNoteRequest{
					Title: strPtr("Should fail"),
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
		n, err := storage.Create(ctx, note.CreateNoteRequest{
			Title:   "To Delete",
			Content: "Content to delete",
			Type:    "text",
		})
		require.NoError(t, err)

		tests := []struct {
			name    string
			id      int64
			wantErr bool
		}{
			{
				name:    "delete existing",
				id:      n.ID,
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
		_, err := storage.db.Exec("DELETE FROM notes")
		require.NoError(t, err)

		// Create test data
		for i := 0; i < 5; i++ {
			_, err := storage.Create(ctx, note.CreateNoteRequest{
				Title:    fmt.Sprintf("Test Note %d", i+1),
				Content:  fmt.Sprintf("Content %d", i+1),
				Type:     "text",
				Tags:     []string{fmt.Sprintf("tag%d", i+1)},
				Metadata: map[string]interface{}{"index": i + 1},
			})
			require.NoError(t, err)
		}

		// Wait a bit to ensure different timestamps
		time.Sleep(10 * time.Millisecond)

		// Create one more with specific tags and type
		_, err = storage.Create(ctx, note.CreateNoteRequest{
			Title:    "Special Note",
			Content:  "Special content",
			Type:     "markdown",
			Tags:     []string{"special", "test"},
			Metadata: map[string]interface{}{"special": true},
		})
		require.NoError(t, err)

		tests := []struct {
			name      string
			req       note.ListNotesRequest
			wantTotal int
			wantItems int
			wantErr   bool
		}{
			{
				name: "list all",
				req: note.ListNotesRequest{
					Limit:  10,
					Offset: 0,
				},
				wantTotal: 6,
				wantItems: 6,
				wantErr:   false,
			},
			{
				name: "list with limit",
				req: note.ListNotesRequest{
					Limit:  3,
					Offset: 0,
				},
				wantTotal: 6,
				wantItems: 3,
				wantErr:   false,
			},
			{
				name: "list with offset",
				req: note.ListNotesRequest{
					Limit:  10,
					Offset: 3,
				},
				wantTotal: 6,
				wantItems: 3,
				wantErr:   false,
			},
			{
				name: "list with search",
				req: note.ListNotesRequest{
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
				req: note.ListNotesRequest{
					Limit:  10,
					Offset: 0,
					Tags:   []string{"special"},
				},
				wantTotal: 1,
				wantItems: 1,
				wantErr:   false,
			},
			{
				name: "list with type filter",
				req: note.ListNotesRequest{
					Limit:  10,
					Offset: 0,
					Type:   "markdown",
				},
				wantTotal: 1,
				wantItems: 1,
				wantErr:   false,
			},
			{
				name: "list with order by created_at desc",
				req: note.ListNotesRequest{
					Limit:    10,
					Offset:   0,
					OrderBy:  "created_at",
					OrderDir: "desc",
				},
				wantTotal: 6,
				wantItems: 6,
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
				if tt.req.OrderBy == "" || (tt.req.OrderBy == "created_at" && tt.req.OrderDir != "asc") {
					for i := 1; i < len(response.Items); i++ {
						assert.True(t, response.Items[i-1].CreatedAt.After(response.Items[i].CreatedAt) ||
							response.Items[i-1].CreatedAt.Equal(response.Items[i].CreatedAt))
					}
				}
			})
		}
	})
}

func strPtr(s string) *string {
	return &s
}
