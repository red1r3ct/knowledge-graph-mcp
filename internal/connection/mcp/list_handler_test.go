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

func TestListHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mock.NewMockStorage(ctrl)
	handler := mcp.NewListHandler(mockStorage)

	now := time.Now()
	desc := "Test connection description"

	tests := []struct {
		name        string
		args        map[string]interface{}
		mockSetup   func()
		wantErr     bool
		wantContent string
	}{
		{
			name: "successful list with defaults",
			args: map[string]interface{}{},
			mockSetup: func() {
				mockStorage.EXPECT().
					List(gomock.Any(), connection.ListConnectionsRequest{
						Limit:    100,
						Offset:   0,
						OrderBy:  "id",
						OrderDir: "asc",
					}).
					Return(&connection.ListConnectionsResponse{
						Items: []connection.Connection{
							{
								ID:         1,
								FromNoteID: 1,
								ToNoteID:   2,
								Type:       "relates_to",
								Strength:   5,
								CreatedAt:  now,
								UpdatedAt:  now,
							},
							{
								ID:          2,
								FromNoteID:  2,
								ToNoteID:    3,
								Type:        "supports",
								Description: &desc,
								Strength:    7,
								CreatedAt:   now,
								UpdatedAt:   now,
							},
						},
						Total: 2,
					}, nil)
			},
			wantErr:     false,
			wantContent: "Found 2 connections",
		},
		{
			name: "successful list with pagination",
			args: map[string]interface{}{
				"limit":  50,
				"offset": 10,
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					List(gomock.Any(), connection.ListConnectionsRequest{
						Limit:    50,
						Offset:   10,
						OrderBy:  "id",
						OrderDir: "asc",
					}).
					Return(&connection.ListConnectionsResponse{
						Items: []connection.Connection{
							{
								ID:         11,
								FromNoteID: 1,
								ToNoteID:   2,
								Type:       "relates_to",
								Strength:   5,
								CreatedAt:  now,
								UpdatedAt:  now,
							},
						},
						Total: 100,
					}, nil)
			},
			wantErr:     false,
			wantContent: "Found 1 connections (showing 11-11 of 100 total)",
		},
		{
			name: "successful list with filters",
			args: map[string]interface{}{
				"from_note_id": int64(1),
				"type":         "relates_to",
				"strength":     5,
			},
			mockSetup: func() {
				fromNoteID := int64(1)
				connectionType := "relates_to"
				strength := 5
				mockStorage.EXPECT().
					List(gomock.Any(), connection.ListConnectionsRequest{
						Limit:      100,
						Offset:     0,
						FromNoteID: &fromNoteID,
						Type:       &connectionType,
						Strength:   &strength,
						OrderBy:    "id",
						OrderDir:   "asc",
					}).
					Return(&connection.ListConnectionsResponse{
						Items: []connection.Connection{
							{
								ID:         1,
								FromNoteID: 1,
								ToNoteID:   2,
								Type:       "relates_to",
								Strength:   5,
								CreatedAt:  now,
								UpdatedAt:  now,
							},
						},
						Total: 1,
					}, nil)
			},
			wantErr:     false,
			wantContent: "Found 1 connections",
		},
		{
			name: "successful list with ordering",
			args: map[string]interface{}{
				"order_by":  "created_at",
				"order_dir": "desc",
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					List(gomock.Any(), connection.ListConnectionsRequest{
						Limit:    100,
						Offset:   0,
						OrderBy:  "created_at",
						OrderDir: "desc",
					}).
					Return(&connection.ListConnectionsResponse{
						Items: []connection.Connection{},
						Total: 0,
					}, nil)
			},
			wantErr:     false,
			wantContent: "Found 0 connections",
		},
		{
			name: "invalid limit - too low",
			args: map[string]interface{}{
				"limit": 0,
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "limit must be between 1 and 1000",
		},
		{
			name: "invalid limit - too high",
			args: map[string]interface{}{
				"limit": 1001,
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "limit must be between 1 and 1000",
		},
		{
			name: "invalid limit type",
			args: map[string]interface{}{
				"limit": "invalid",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "invalid limit",
		},
		{
			name: "invalid offset - negative",
			args: map[string]interface{}{
				"offset": -1,
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "offset must be non-negative",
		},
		{
			name: "invalid offset type",
			args: map[string]interface{}{
				"offset": "invalid",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "invalid offset",
		},
		{
			name: "invalid from_note_id type",
			args: map[string]interface{}{
				"from_note_id": "invalid",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "invalid from_note_id",
		},
		{
			name: "invalid to_note_id type",
			args: map[string]interface{}{
				"to_note_id": "invalid",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "invalid to_note_id",
		},
		{
			name: "invalid connection type",
			args: map[string]interface{}{
				"type": "invalid_type",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "invalid connection type",
		},
		{
			name: "invalid strength - too low",
			args: map[string]interface{}{
				"strength": 0,
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "strength must be between 1 and 10",
		},
		{
			name: "invalid strength - too high",
			args: map[string]interface{}{
				"strength": 11,
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "strength must be between 1 and 10",
		},
		{
			name: "invalid order_by",
			args: map[string]interface{}{
				"order_by": "invalid_field",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "invalid order_by",
		},
		{
			name: "invalid order_dir",
			args: map[string]interface{}{
				"order_dir": "invalid_dir",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "invalid order_dir",
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
			wantContent: "failed to list connections",
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