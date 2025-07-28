package mcp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"

	"github.com/red1r3ct/knowledge-graph-mcp/internal/note"
	"github.com/red1r3ct/knowledge-graph-mcp/internal/note/mcp"
	"github.com/red1r3ct/knowledge-graph-mcp/internal/note/mock"
)

func TestListHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mock.NewMockStorage(ctrl)
	handler := mcp.NewListHandler(mockStorage)

	now := time.Now()
	metadata := map[string]interface{}{"key": "value"}

	tests := []struct {
		name        string
		args        map[string]interface{}
		mockSetup   func()
		wantErr     bool
		wantContent string
	}{
		{
			name: "successful list with results",
			args: map[string]interface{}{
				"limit":  float64(10),
				"offset": float64(0),
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					List(gomock.Any(), note.ListNotesRequest{
						Limit:  10,
						Offset: 0,
					}).
					Return(&note.ListNotesResponse{
						Items: []note.Note{
							{
								ID:        1,
								Title:     "Note 1",
								Content:   "Content 1",
								Type:      "text",
								Tags:      []string{"tag1"},
								Metadata:  metadata,
								CreatedAt: now,
								UpdatedAt: now,
							},
							{
								ID:        2,
								Title:     "Note 2",
								Content:   "Content 2",
								Type:      "markdown",
								Tags:      []string{"tag2"},
								CreatedAt: now,
								UpdatedAt: now,
							},
						},
						Total: 2,
					}, nil)
			},
			wantErr:     false,
			wantContent: "Found 2 notes (total: 2)",
		},
		{
			name: "list with filtering",
			args: map[string]interface{}{
				"search":    "test",
				"type":      "markdown",
				"tags":      []interface{}{"tag1", "tag2"},
				"order_by":  "created_at",
				"order_dir": "desc",
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					List(gomock.Any(), note.ListNotesRequest{
						Limit:    100, // default
						Offset:   0,   // default
						Search:   "test",
						Type:     "markdown",
						Tags:     []string{"tag1", "tag2"},
						OrderBy:  "created_at",
						OrderDir: "desc",
					}).
					Return(&note.ListNotesResponse{
						Items: []note.Note{
							{
								ID:        1,
								Title:     "Test Note",
								Content:   "Test Content",
								Type:      "markdown",
								Tags:      []string{"tag1", "tag2"},
								CreatedAt: now,
								UpdatedAt: now,
							},
						},
						Total: 1,
					}, nil)
			},
			wantErr:     false,
			wantContent: "Found 1 notes (total: 1)",
		},
		{
			name: "empty results",
			args: map[string]interface{}{},
			mockSetup: func() {
				mockStorage.EXPECT().
					List(gomock.Any(), note.ListNotesRequest{
						Limit:  100, // default
						Offset: 0,   // default
					}).
					Return(&note.ListNotesResponse{
						Items: []note.Note{},
						Total: 0,
					}, nil)
			},
			wantErr:     false,
			wantContent: "No notes found",
		},
		{
			name: "storage error",
			args: map[string]interface{}{},
			mockSetup: func() {
				mockStorage.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("storage error"))
			},
			wantErr:     true,
			wantContent: "failed to list notes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := gomcp.CallToolRequest{
				Params: gomcp.CallToolParams{
					Arguments: tt.args,
				},
			}

			result, err := handler(context.Background(), req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantContent)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Contains(t, result.Content[0].(gomcp.TextContent).Text, tt.wantContent)
			}
		})
	}
}