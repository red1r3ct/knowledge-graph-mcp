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

func TestCreateHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mock.NewMockStorage(ctrl)
	handler := mcp.NewCreateHandler(mockStorage)

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
			name: "successful creation",
			args: map[string]interface{}{
				"title":    "Test Note",
				"content":  "Test Content",
				"type":     "markdown",
				"tags":     []interface{}{"tag1", "tag2"},
				"metadata": metadata,
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					Create(gomock.Any(), note.CreateNoteRequest{
						Title:    "Test Note",
						Content:  "Test Content",
						Type:     "markdown",
						Tags:     []string{"tag1", "tag2"},
						Metadata: metadata,
					}).
					Return(&note.Note{
						ID:        1,
						Title:     "Test Note",
						Content:   "Test Content",
						Type:      "markdown",
						Tags:      []string{"tag1", "tag2"},
						Metadata:  metadata,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
			},
			wantErr:     false,
			wantContent: "Successfully created note with ID: 1",
		},
		{
			name: "successful creation with defaults",
			args: map[string]interface{}{
				"title":   "Test Note",
				"content": "Test Content",
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					Create(gomock.Any(), note.CreateNoteRequest{
						Title:   "Test Note",
						Content: "Test Content",
						Type:    "text",
						Tags:    nil,
					}).
					Return(&note.Note{
						ID:        2,
						Title:     "Test Note",
						Content:   "Test Content",
						Type:      "text",
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
			},
			wantErr:     false,
			wantContent: "Successfully created note with ID: 2",
		},
		{
			name: "missing title",
			args: map[string]interface{}{
				"content": "Test Content",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "title is required",
		},
		{
			name: "missing content",
			args: map[string]interface{}{
				"title": "Test Note",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "content is required",
		},
		{
			name: "storage error",
			args: map[string]interface{}{
				"title":   "Test Note",
				"content": "Test Content",
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("storage error"))
			},
			wantErr:     true,
			wantContent: "failed to create note",
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