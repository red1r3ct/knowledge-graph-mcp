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

func TestCreateHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mock.NewMockStorage(ctrl)
	handler := mcp.NewCreateHandler(mockStorage)

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
			name: "successful creation",
			args: map[string]interface{}{
				"name":        "Test KB",
				"description": "Test Description",
				"tags":        []interface{}{"tag1", "tag2"},
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					Create(gomock.Any(), knowledgebase.CreateRequest{
						Name:        "Test KB",
						Description: &desc,
						Tags:        []string{"tag1", "tag2"},
					}).
					Return(&knowledgebase.KnowledgeBase{
						ID:          1,
						Name:        "Test KB",
						Description: &desc,
						Tags:        []string{"tag1", "tag2"},
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)
			},
			wantErr:     false,
			wantContent: "Successfully created knowledge base entry with ID: 1",
		},
		{
			name: "missing name",
			args: map[string]interface{}{
				"description": desc,
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "name is required",
		},
		{
			name: "storage error",
			args: map[string]interface{}{
				"name": "Test KB",
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("storage error"))
			},
			wantErr:     true,
			wantContent: "failed to create knowledge base",
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
