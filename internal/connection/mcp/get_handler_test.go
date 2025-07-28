package mcp_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"

	"github.com/red1r3ct/knowledge-graph-mcp/internal/connection"
	"github.com/red1r3ct/knowledge-graph-mcp/internal/connection/mcp"
	"github.com/red1r3ct/knowledge-graph-mcp/internal/connection/mock"
)

func TestGetHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mock.NewMockStorage(ctrl)
	handler := mcp.NewGetHandler(mockStorage)

	now := time.Now()
	desc := "Test connection description"
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
				"id": int64(1),
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					Get(gomock.Any(), int64(1)).
					Return(&connection.Connection{
						ID:          1,
						FromNoteID:  1,
						ToNoteID:    2,
						Type:        "relates_to",
						Description: &desc,
						Strength:    7,
						Metadata:    metadata,
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)
			},
			wantErr:     false,
			wantContent: "Connection found:",
		},
		{
			name: "successful get with string id",
			args: map[string]interface{}{
				"id": "1",
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					Get(gomock.Any(), int64(1)).
					Return(&connection.Connection{
						ID:         1,
						FromNoteID: 1,
						ToNoteID:   2,
						Type:       "relates_to",
						Strength:   5,
						CreatedAt:  now,
						UpdatedAt:  now,
					}, nil)
			},
			wantErr:     false,
			wantContent: "Connection found:",
		},
		{
			name: "successful get with float id",
			args: map[string]interface{}{
				"id": float64(1),
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					Get(gomock.Any(), int64(1)).
					Return(&connection.Connection{
						ID:         1,
						FromNoteID: 1,
						ToNoteID:   2,
						Type:       "relates_to",
						Strength:   5,
						CreatedAt:  now,
						UpdatedAt:  now,
					}, nil)
			},
			wantErr:     false,
			wantContent: "Connection found:",
		},
		{
			name: "missing id",
			args: map[string]interface{}{},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "id is required",
		},
		{
			name: "invalid id type",
			args: map[string]interface{}{
				"id": "invalid",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "invalid id",
		},
		{
			name: "zero id",
			args: map[string]interface{}{
				"id": int64(0),
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "id must be a positive integer",
		},
		{
			name: "negative id",
			args: map[string]interface{}{
				"id": int64(-1),
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "id must be a positive integer",
		},
		{
			name: "storage error",
			args: map[string]interface{}{
				"id": int64(1),
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					Get(gomock.Any(), int64(1)).
					Return(nil, errors.New("storage error"))
			},
			wantErr:     true,
			wantContent: "failed to get connection",
		},
		{
			name: "connection not found",
			args: map[string]interface{}{
				"id": int64(999),
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					Get(gomock.Any(), int64(999)).
					Return(nil, errors.New("connection not found"))
			},
			wantErr:     true,
			wantContent: "failed to get connection",
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