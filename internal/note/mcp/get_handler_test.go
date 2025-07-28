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

func TestGetHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mock.NewMockStorage(ctrl)
	handler := mcp.NewGetHandler(mockStorage)

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
			name: "successful get",
			args: map[string]interface{}{
				"id": "1",
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					Get(gomock.Any(), int64(1)).
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
			wantContent: "Test Note",
		},
		{
			name: "note not found",
			args: map[string]interface{}{
				"id": "999",
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					Get(gomock.Any(), int64(999)).
					Return(nil, nil)
			},
			wantErr:     false,
			wantContent: "Note with ID 999 not found",
		},
		{
			name: "missing id",
			args: map[string]interface{}{},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "id is required",
		},
		{
			name: "invalid id format",
			args: map[string]interface{}{
				"id": "invalid",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "invalid id format",
		},
		{
			name: "storage error",
			args: map[string]interface{}{
				"id": "1",
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					Get(gomock.Any(), int64(1)).
					Return(nil, errors.New("storage error"))
			},
			wantErr:     true,
			wantContent: "failed to get note",
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