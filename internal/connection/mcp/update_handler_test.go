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

func TestUpdateHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mock.NewMockStorage(ctrl)
	handler := mcp.NewUpdateHandler(mockStorage)

	now := time.Now()
	desc := "Updated connection description"
	newType := "supports"
	newStrength := 8
	metadata := map[string]interface{}{"updated": "value"}

	tests := []struct {
		name        string
		args        map[string]interface{}
		mockSetup   func()
		wantErr     bool
		wantContent string
	}{
		{
			name: "successful update all fields",
			args: map[string]interface{}{
				"id":          int64(1),
				"type":        "supports",
				"description": "Updated connection description",
				"strength":    8,
				"metadata":    metadata,
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					Update(gomock.Any(), int64(1), connection.UpdateConnectionRequest{
						Type:        &newType,
						Description: &desc,
						Strength:    &newStrength,
						Metadata:    metadata,
					}).
					Return(&connection.Connection{
						ID:          1,
						FromNoteID:  1,
						ToNoteID:    2,
						Type:        "supports",
						Description: &desc,
						Strength:    8,
						Metadata:    metadata,
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)
			},
			wantErr:     false,
			wantContent: "Successfully updated connection with ID: 1",
		},
		{
			name: "successful update type only",
			args: map[string]interface{}{
				"id":   int64(1),
				"type": "references",
			},
			mockSetup: func() {
				refType := "references"
				mockStorage.EXPECT().
					Update(gomock.Any(), int64(1), connection.UpdateConnectionRequest{
						Type: &refType,
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
			wantContent: "Successfully updated connection with ID: 1",
		},
		{
			name: "successful update strength only",
			args: map[string]interface{}{
				"id":       int64(1),
				"strength": 9,
			},
			mockSetup: func() {
				strength := 9
				mockStorage.EXPECT().
					Update(gomock.Any(), int64(1), connection.UpdateConnectionRequest{
						Strength: &strength,
					}).
					Return(&connection.Connection{
						ID:         1,
						FromNoteID: 1,
						ToNoteID:   2,
						Type:       "relates_to",
						Strength:   9,
						CreatedAt:  now,
						UpdatedAt:  now,
					}, nil)
			},
			wantErr:     false,
			wantContent: "Successfully updated connection with ID: 1",
		},
		{
			name: "successful update description only",
			args: map[string]interface{}{
				"id":          int64(1),
				"description": "New description",
			},
			mockSetup: func() {
				newDesc := "New description"
				mockStorage.EXPECT().
					Update(gomock.Any(), int64(1), connection.UpdateConnectionRequest{
						Description: &newDesc,
					}).
					Return(&connection.Connection{
						ID:          1,
						FromNoteID:  1,
						ToNoteID:    2,
						Type:        "relates_to",
						Description: &newDesc,
						Strength:    5,
						CreatedAt:   now,
						UpdatedAt:   now,
					}, nil)
			},
			wantErr:     false,
			wantContent: "Successfully updated connection with ID: 1",
		},
		{
			name: "missing id",
			args: map[string]interface{}{
				"type": "supports",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "id is required",
		},
		{
			name: "invalid id type",
			args: map[string]interface{}{
				"id":   "invalid",
				"type": "supports",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "invalid id",
		},
		{
			name: "zero id",
			args: map[string]interface{}{
				"id":   int64(0),
				"type": "supports",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "id must be a positive integer",
		},
		{
			name: "negative id",
			args: map[string]interface{}{
				"id":   int64(-1),
				"type": "supports",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "id must be a positive integer",
		},
		{
			name: "invalid connection type",
			args: map[string]interface{}{
				"id":   int64(1),
				"type": "invalid_type",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "invalid connection type",
		},
		{
			name: "invalid strength - too low",
			args: map[string]interface{}{
				"id":       int64(1),
				"strength": 0,
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "strength must be between 1 and 10",
		},
		{
			name: "invalid strength - too high",
			args: map[string]interface{}{
				"id":       int64(1),
				"strength": 11,
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "strength must be between 1 and 10",
		},
		{
			name: "invalid strength type",
			args: map[string]interface{}{
				"id":       int64(1),
				"strength": "invalid",
			},
			mockSetup:   func() {},
			wantErr:     true,
			wantContent: "invalid strength",
		},
		{
			name: "storage error",
			args: map[string]interface{}{
				"id":   int64(1),
				"type": "supports",
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					Update(gomock.Any(), int64(1), gomock.Any()).
					Return(nil, errors.New("storage error"))
			},
			wantErr:     true,
			wantContent: "failed to update connection",
		},
		{
			name: "connection not found",
			args: map[string]interface{}{
				"id":   int64(999),
				"type": "supports",
			},
			mockSetup: func() {
				mockStorage.EXPECT().
					Update(gomock.Any(), int64(999), gomock.Any()).
					Return(nil, errors.New("connection not found"))
			},
			wantErr:     true,
			wantContent: "failed to update connection",
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