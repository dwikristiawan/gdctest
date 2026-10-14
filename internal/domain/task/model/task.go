package model

import "time"

// Task status iota
const (
	StatusBacklog    int = 0
	StatusTodo       int = 1
	StatusInProgress int = 2
	StatusInReview   int = 3
	StatusDone       int = 4
	StatusArchived   int = 5
)

// Task collaborator roles
const (
	RoleAssignee int = 0
	RoleReviewer int = 1
	RoleWatcher  int = 2
)

type Task struct {
	ID          string     `json:"id" db:"id"`
	TeamID      string     `json:"team_id" db:"team_id"`
	Title       string     `json:"title" db:"title"`
	Description string     `json:"description" db:"description"`
	Status      int        `json:"status" db:"status"`
	OwnerID     string     `json:"owner_id" db:"owner_id"`
	UpdatedBy   *string    `json:"updated_by,omitempty" db:"updated_by"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`

	Collaborators []*TaskCollaborator `json:"collaborators,omitempty"`
}

type TaskCollaborator struct {
	TaskID    string    `json:"task_id" db:"task_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	RoleType  int       `json:"role_type" db:"role_type"`
	UpdatedBy *string   `json:"updated_by,omitempty" db:"updated_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Joined user fields
	Email    string `json:"email,omitempty" db:"email"`
	Username string `json:"username,omitempty" db:"username"`
}
