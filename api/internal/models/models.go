package models

import "time"

type User struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type Label struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type Issue struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`   // Backlog, Todo, In Progress, Done, Canceled
	Priority    string    `json:"priority"` // Low, Medium, High, Critical
	AssigneeID  *string   `json:"assignee_id"`
	Assignee    *User     `json:"assignee,omitempty"` // For response population
	Labels      []Label   `json:"labels,omitempty"`   // For response population
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	OrderIndex  float64   `json:"order_index"`
}

type CreateIssueRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Status      string   `json:"status"`
	Priority    string   `json:"priority"`
	AssigneeID  *string  `json:"assignee_id"`
	LabelIDs    []string `json:"label_ids"`
}

type UpdateIssueRequest struct {
	Title       *string  `json:"title"`
	Description *string  `json:"description"`
	Status      *string  `json:"status"`
	Priority    *string  `json:"priority"`
	AssigneeID  *string  `json:"assignee_id"`
	LabelIDs    []string `json:"label_ids"`
	OrderIndex  *float64 `json:"order_index"`
}
