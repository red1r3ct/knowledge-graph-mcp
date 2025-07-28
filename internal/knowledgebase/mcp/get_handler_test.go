package mcp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"

	"github.com/red1r3ct/knowledge-graph-mcp/internal/knowledgebase"
	"github.com/red1r3ct/knowledge-graph-mcp/internal/knowledgebase/mcp"
	"github.com/red1r3ct/knowledge-graph-mcp/internal/knowledgebase/mock"
)

func TestGetHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mock.NewMockStorage(ctrl)
	handler := mcp.NewGetHandler(mockStorage)

	now := time.Now()
	desc := "Test Description"

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
				"id": "123",
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					Get(gomock.Any(), int64(123)).
					Return(&knowledgebase.KnowledgeBase{
						ID:          123,
						Name:        "Test KB",
						Description: &desc,
						Tags:        []string{"tag1"},
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)
			},
			wantErr:     false,
			wantContent: `"name": "Test KB"`,
		},
		{
			name: "not found",
			args: map[string]interface{}{
				"id": "123",
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					Get(gomock.Any(), int64(123)).
					Return(nil, nil)
			},
			wantErr:     false,
			wantContent: "Knowledge base entry with ID 123 not found",
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
					Get(gomock.Any(), int64(123)).
					Return(nil, errors.New("storage error"))
			},
			wantErr:     true,
			wantContent: "failed to get knowledge base",
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