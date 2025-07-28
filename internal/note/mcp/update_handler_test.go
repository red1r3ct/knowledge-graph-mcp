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

func TestUpdateHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mock.NewMockStorage(ctrl)
	handler := mcp.NewUpdateHandler(mockStorage)

	now := time.Now()
	metadata := map[string]interface{}{"key": "updated_value"}
	title := "Updated Title"
	content := "Updated Content"
	noteType := "code"

	tests := []struct {
		name        string
		args        map[string]interface{}
		mockSetup   func()
		wantErr     bool
		wantContent string
	}{
		{
			name: "successful update",
			args: map[string]interface{}{
				"id":       "1",
				"title":    "Updated Title",
				"content":  "Updated Content",
				"type":     "code",
				"tags":     []interface{}{"tag1", "tag3"},
				"metadata": metadata,
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					Update(gomock.Any(), int64(1), note.UpdateNoteRequest{
						Title:    &title,
						Content:  &content,
						Type:     &noteType,
						Tags:     []string{"tag1", "tag3"},
						Metadata: metadata,
					}).
					Return(&note.Note{
						ID:        1,
						Title:     "Updated Title",
						Content:   "Updated Content",
						Type:      "code",
						Tags:      []string{"tag1", "tag3"},
						Metadata:  metadata,
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
			},
			wantErr:     false,
			wantContent: "Successfully updated note with ID: 1",
		},
		{
			name: "partial update",
			args: map[string]interface{}{
				"id":    "1",
				"title": "Updated Title Only",
			},
			mockSetup: func() {
				partialTitle := "Updated Title Only"
				mockStorage.EXPECT().
					Update(gomock.Any(), int64(1), note.UpdateNoteRequest{
						Title: &partialTitle,
					}).
					Return(&note.Note{
						ID:        1,
						Title:     "Updated Title Only",
						Content:   "Original Content",
						Type:      "text",
						CreatedAt: now,
						UpdatedAt: now,
					}, nil)
			},
			wantErr:     false,
			wantContent: "Successfully updated note with ID: 1",
		},
		{
			name: "note not found",
			args: map[string]interface{}{
				"id":    "999",
				"title": "Updated Title",
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					Update(gomock.Any(), int64(999), gomock.Any()).
					Return(nil, nil)
			},
			wantErr:     false,
			wantContent: "Note with ID 999 not found",
		},
		{
			name: "missing id",
			args: map[string]interface{}{
				"title": "Updated Title",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "id is required",
		},
		{
			name: "invalid id format",
			args: map[string]interface{}{
				"id":    "invalid",
				"title": "Updated Title",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "invalid id format",
		},
		{
			name: "storage error",
			args: map[string]interface{}{
				"id":    "1",
				"title": "Updated Title",
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					Update(gomock.Any(), int64(1), gomock.Any()).
					Return(nil, errors.New("storage error"))
			},
			wantErr:     true,
			wantContent: "failed to update note",
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