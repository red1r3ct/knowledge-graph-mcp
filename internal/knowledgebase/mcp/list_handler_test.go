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

func TestListHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mock.NewMockStorage(ctrl)
	handler := mcp.NewListHandler(mockStorage)

	now := time.Now()
	desc1 := "Description 1"
	desc2 := "Description 2"
	descTest := "Test Description"

	tests := []struct {
		name        string
		args        map[string]interface{}
		mockSetup   func()
		wantErr     bool
		wantContent string
	}{
		{
			name: "successful list",
			args: map[string]interface{}{},
			mockSetup: func() {
				mockStorage.EXPECT().
					List(gomock.Any(), knowledgebase.ListRequest{
						Limit:  100,
						Offset: 0,
					}).
					Return(&knowledgebase.ListResponse{
						Items: []knowledgebase.KnowledgeBase{
							{
								ID:          1,
								Name:        "Test KB 1",
								Description: &desc1,
								Tags:        []string{"tag1"},
								CreatedAt:   now,
								UpdatedAt:   now,
							},
							{
								ID:          2,
								Name:        "Test KB 2",
								Description: &desc2,
								Tags:        []string{"tag2"},
								CreatedAt:   now,
								UpdatedAt:   now,
							},
						},
						Total: 2,
					}, nil)
			},
			wantErr:     false,
			wantContent: "Found 2 knowledge base entries",
		},
		{
			name: "list with search",
			args: map[string]interface{}{
				"search": "test",
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					List(gomock.Any(), knowledgebase.ListRequest{
						Search: "test",
						Limit:  100,
						Offset: 0,
					}).
					Return(&knowledgebase.ListResponse{
						Items: []knowledgebase.KnowledgeBase{
							{
								ID:          1,
								Name:        "Test KB",
								Description: &descTest,
								Tags:        []string{"tag1"},
								CreatedAt:   now,
								UpdatedAt:   now,
							},
						},
						Total: 1,
					}, nil)
			},
			wantErr:     false,
			wantContent: "Found 1 knowledge base entries",
		},
		{
			name: "empty list",
			args: map[string]interface{}{},
			mockSetup: func() {
				mockStorage.EXPECT().
					List(gomock.Any(), knowledgebase.ListRequest{
						Limit:  100,
						Offset: 0,
					}).
					Return(&knowledgebase.ListResponse{
						Items: []knowledgebase.KnowledgeBase{},
						Total: 0,
					}, nil)
			},
			wantErr:     false,
			wantContent: "No knowledge base entries found",
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
			wantContent: "failed to list knowledge bases",
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