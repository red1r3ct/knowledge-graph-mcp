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

func TestCreateHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mock.NewMockStorage(ctrl)
	handler := mcp.NewCreateHandler(mockStorage)

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
			name: "successful creation",
			args: map[string]interface{}{
				"from_note_id": int64(1),
				"to_note_id":   int64(2),
				"type":         "relates_to",
				"description":  "Test connection description",
				"strength":     7,
				"metadata":     metadata,
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					Create(gomock.Any(), connection.CreateConnectionRequest{
						FromNoteID:  1,
						ToNoteID:    2,
						Type:        "relates_to",
						Description: &desc,
						Strength:    7,
						Metadata:    metadata,
					}).
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
			wantContent: "Successfully created connection with ID: 1",
		},
		{
			name: "successful creation with defaults",
			args: map[string]interface{}{
				"from_note_id": int64(1),
				"to_note_id":   int64(2),
				"type":         "references",
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					Create(gomock.Any(), connection.CreateConnectionRequest{
						FromNoteID:  1,
						ToNoteID:    2,
						Type:        "references",
						Description: nil,
						Strength:    5, // default
						Metadata:    nil,
					}).
					Return(&connection.Connection{
						ID:         1,
						FromNoteID: 1,
						ToNoteID:   2,
						Type:       "references",
						Strength:   5,
						CreatedAt:  now,
						UpdatedAt:  now,
					}, nil)
			},
			wantErr:     false,
			wantContent: "Successfully created connection with ID: 1",
		},
		{
			name: "missing from_note_id",
			args: map[string]interface{}{
				"to_note_id": int64(2),
				"type":       "relates_to",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "from_note_id is required",
		},
		{
			name: "missing to_note_id",
			args: map[string]interface{}{
				"from_note_id": int64(1),
				"type":         "relates_to",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "to_note_id is required",
		},
		{
			name: "missing type",
			args: map[string]interface{}{
				"from_note_id": int64(1),
				"to_note_id":   int64(2),
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "type is required",
		},
		{
			name: "same from and to note",
			args: map[string]interface{}{
				"from_note_id": int64(1),
				"to_note_id":   int64(1),
				"type":         "relates_to",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "from_note_id and to_note_id cannot be the same",
		},
		{
			name: "invalid connection type",
			args: map[string]interface{}{
				"from_note_id": int64(1),
				"to_note_id":   int64(2),
				"type":         "invalid_type",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "invalid connection type",
		},
		{
			name: "invalid strength - too low",
			args: map[string]interface{}{
				"from_note_id": int64(1),
				"to_note_id":   int64(2),
				"type":         "relates_to",
				"strength":     0,
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "strength must be between 1 and 10",
		},
		{
			name: "invalid strength - too high",
			args: map[string]interface{}{
				"from_note_id": int64(1),
				"to_note_id":   int64(2),
				"type":         "relates_to",
				"strength":     11,
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "strength must be between 1 and 10",
		},
		{
			name: "invalid from_note_id type",
			args: map[string]interface{}{
				"from_note_id": "invalid",
				"to_note_id":   int64(2),
				"type":         "relates_to",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "invalid from_note_id",
		},
		{
			name: "invalid to_note_id type",
			args: map[string]interface{}{
				"from_note_id": int64(1),
				"to_note_id":   "invalid",
				"type":         "relates_to",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "invalid to_note_id",
		},
		{
			name: "storage error",
			args: map[string]interface{}{
				"from_note_id": int64(1),
				"to_note_id":   int64(2),
				"type":         "relates_to",
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("storage error"))
			},
			wantErr:     true,
			wantContent: "failed to create connection",
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