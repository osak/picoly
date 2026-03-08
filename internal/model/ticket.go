package model

import "time"

// Status represents the status of a ticket.
type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
	StatusCancelled  Status = "cancelled"
	StatusProblem    Status = "problem"
)

// Valid reports whether the status is a valid value.
func (s Status) Valid() bool {
	switch s {
	case StatusTodo, StatusInProgress, StatusDone, StatusCancelled, StatusProblem:
		return true
	}
	return false
}

// Ticket represents a task in the Kanban board.
type Ticket struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	Assignee    string    `json:"assignee,omitempty"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Comments    []Comment `json:"comments,omitempty"`
}

// Comment represents a progress note on a ticket.
type Comment struct {
	ID        int64     `json:"id"`
	TicketID  int64     `json:"ticket_id"`
	Body      string    `json:"body"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

// ListItem is a lightweight representation of a ticket for list views.
type ListItem struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	Status       Status    `json:"status"`
	Assignee     string    `json:"assignee,omitempty"`
	CommentCount int       `json:"comment_count"`
	UpdatedAt    time.Time `json:"updated_at"`
}
