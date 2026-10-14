package model

import "time"

// Team roles
const (
	RoleTeamAdmin  int = 0 // Admin: can add/remove members
	RoleTeamMember int = 1 // Member: can create/assign tasks
)

type Team struct {
	ID          string     `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description" db:"description"`
	OwnerID     string     `json:"owner_id" db:"owner_id"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

type TeamMember struct {
	TeamID    string     `json:"team_id" db:"team_id"`
	UserID    string     `json:"user_id" db:"user_id"`
	RoleType  int        `json:"role_type" db:"role_type"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`

	// Joined user fields
	Email    string `json:"email,omitempty" db:"email"`
	Username string `json:"username,omitempty" db:"username"`
}
