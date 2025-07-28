package mcp_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"

	"github.com/red1r3ct/knowledge-graph-mcp/internal/knowledgebase/mcp"
	"github.com/red1r3ct/knowledge-graph-mcp/internal/knowledgebase/mock"
)

func TestDeleteHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mock.NewMockStorage(ctrl)
	handler := mcp.NewDeleteHandler(mockStorage)

	tests := []struct {
		name        string
		args        map[string]interface{}
		mockSetup   func()
		wantErr     bool
		wantContent string
	}{
		{
			name: "successful delete",
			args: map[string]interface{}{
				"id": "123",
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					Delete(gomock.Any(), int64(123)).
					Return(nil)
			},
			wantErr:     false,
			wantContent: "Successfully deleted knowledge base entry with ID: 123",
		},
		{
			name: "missing id",
			args: map[string]interface{}{},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "id is required",
		},
		{
			name: "invalid id",
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
				"id": "123",
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					Delete(gomock.Any(), int64(123)).
					Return(errors.New("storage error"))
			},
			wantErr:     true,
			wantContent: "failed to delete knowledge base",
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