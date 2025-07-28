package knowledgebase

import (
	"time"
)

// KnowledgeBase represents the domain model for a knowledge base entity
type KnowledgeBase struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateRequest represents the DTO for creating a knowledge base
type CreateRequest struct {
	Name        string   `json:"name"`
	Description *string  `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// UpdateRequest represents the DTO for updating a knowledge base
type UpdateRequest struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// ListRequest represents the DTO for listing knowledge bases
type ListRequest struct {
	Limit  int      `json:"limit,omitempty"`
	Offset int      `json:"offset,omitempty"`
	Search string   `json:"search,omitempty"`
	Tags   []string `json:"tags,omitempty"`
}

// ListResponse represents the DTO for listing response
type ListResponse struct {
	Items []KnowledgeBase `json:"items"`
	Total int64           `json:"total"`
}