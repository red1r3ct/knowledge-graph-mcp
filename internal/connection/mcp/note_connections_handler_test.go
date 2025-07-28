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

func TestNoteConnectionsHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mock.NewMockStorage(ctrl)
	handler := mcp.NewNoteConnectionsHandler(mockStorage)

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
			name: "successful get note connections",
			args: map[string]interface{}{
				"note_id": int64(1),
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					GetNoteConnections(gomock.Any(), connection.NoteConnectionsRequest{
						NoteID: 1,
						Limit:  100,
						Offset: 0,
					}).
					Return(&connection.NoteConnectionsResponse{
						NoteID: 1,
						Outgoing: []connection.Connection{
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
								FromNoteID:  1,
								ToNoteID:    3,
								Type:        "supports",
								Description: &desc,
								Strength:    7,
								CreatedAt:   now,
								UpdatedAt:   now,
							},
						},
						Incoming: []connection.Connection{
							{
								ID:         3,
								FromNoteID: 4,
								ToNoteID:   1,
								Type:       "references",
								Strength:   6,
								CreatedAt:  now,
								UpdatedAt:  now,
							},
						},
						TotalCount: 3,
						TypesCount: map[string]int64{
							"relates_to":  1,
							"supports":    1,
							"references":  1,
						},
					}, nil)
			},
			wantErr:     false,
			wantContent: "Found connections for note 1:\n- 2 outgoing connections\n- 1 incoming connections\n- 3 total connections",
		},
		{
			name: "successful get with filters",
			args: map[string]interface{}{
				"note_id":  int64(1),
				"type":     "relates_to",
				"strength": 5,
				"limit":    50,
				"offset":   10,
			},
			mockSetup: func() {
				connectionType := "relates_to"
				strength := 5
				mockStorage.EXPECT().
					GetNoteConnections(gomock.Any(), connection.NoteConnectionsRequest{
						NoteID:   1,
						Type:     &connectionType,
						Strength: &strength,
						Limit:    50,
						Offset:   10,
					}).
					Return(&connection.NoteConnectionsResponse{
						NoteID:   1,
						Outgoing: []connection.Connection{},
						Incoming: []connection.Connection{},
						TotalCount: 0,
						TypesCount: map[string]int64{},
					}, nil)
			},
			wantErr:     false,
			wantContent: "Found connections for note 1:\n- 0 outgoing connections\n- 0 incoming connections\n- 0 total connections",
		},
		{
			name: "successful get with string note_id",
			args: map[string]interface{}{
				"note_id": "1",
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					GetNoteConnections(gomock.Any(), connection.NoteConnectionsRequest{
						NoteID: 1,
						Limit:  100,
						Offset: 0,
					}).
					Return(&connection.NoteConnectionsResponse{
						NoteID:     1,
						Outgoing:   []connection.Connection{},
						Incoming:   []connection.Connection{},
						TotalCount: 0,
						TypesCount: map[string]int64{},
					}, nil)
			},
			wantErr:     false,
			wantContent: "Found connections for note 1",
		},
		{
			name: "missing note_id",
			args: map[string]interface{}{},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "note_id is required",
		},
		{
			name: "invalid note_id type",
			args: map[string]interface{}{
				"note_id": "invalid",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "invalid note_id",
		},
		{
			name: "zero note_id",
			args: map[string]interface{}{
				"note_id": int64(0),
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "note_id must be a positive integer",
		},
		{
			name: "negative note_id",
			args: map[string]interface{}{
				"note_id": int64(-1),
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "note_id must be a positive integer",
		},
		{
			name: "invalid connection type",
			args: map[string]interface{}{
				"note_id": int64(1),
				"type":    "invalid_type",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "invalid connection type",
		},
		{
			name: "invalid strength - too low",
			args: map[string]interface{}{
				"note_id":  int64(1),
				"strength": 0,
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "strength must be between 1 and 10",
		},
		{
			name: "invalid strength - too high",
			args: map[string]interface{}{
				"note_id":  int64(1),
				"strength": 11,
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "strength must be between 1 and 10",
		},
		{
			name: "invalid strength type",
			args: map[string]interface{}{
				"note_id":  int64(1),
				"strength": "invalid",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "invalid strength",
		},
		{
			name: "invalid limit - too low",
			args: map[string]interface{}{
				"note_id": int64(1),
				"limit":   0,
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "limit must be between 1 and 1000",
		},
		{
			name: "invalid limit - too high",
			args: map[string]interface{}{
				"note_id": int64(1),
				"limit":   1001,
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "limit must be between 1 and 1000",
		},
		{
			name: "invalid limit type",
			args: map[string]interface{}{
				"note_id": int64(1),
				"limit":   "invalid",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "invalid limit",
		},
		{
			name: "invalid offset - negative",
			args: map[string]interface{}{
				"note_id": int64(1),
				"offset":  -1,
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "offset must be non-negative",
		},
		{
			name: "invalid offset type",
			args: map[string]interface{}{
				"note_id": int64(1),
				"offset":  "invalid",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "invalid offset",
		},
		{
			name: "storage error",
			args: map[string]interface{}{
				"note_id": int64(1),
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					GetNoteConnections(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("storage error"))
			},
			wantErr:     true,
			wantContent: "failed to get note connections",
		},
		{
			name: "note not found",
			args: map[string]interface{}{
				"note_id": int64(999),
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					GetNoteConnections(gomock.Any(), connection.NoteConnectionsRequest{
						NoteID: 999,
						Limit:  100,
						Offset: 0,
					}).
					Return(nil, errors.New("note not found"))
			},
			wantErr:     true,
			wantContent: "failed to get note connections",
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