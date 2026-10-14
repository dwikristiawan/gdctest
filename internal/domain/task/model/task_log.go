package model

import "time"

type TaskLog struct {
	ID                 string    `json:"id" db:"id"`
	TaskID             string    `json:"task_id" db:"task_id"`
	PreviousAssigneeID *string   `json:"previous_assignee_id,omitempty" db:"previous_assignee_id"`
	NewAssigneeID      string    `json:"new_assignee_id" db:"new_assignee_id"`
	ChangedBy          string    `json:"changed_by" db:"changed_by"`
	Action             string    `json:"action" db:"action"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
}
