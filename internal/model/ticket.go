package model

import "time"

// Status はチケットのステータスを表す
type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
	StatusCancelled  Status = "cancelled"
)

// Valid はステータスが有効な値かどうかを返す
func (s Status) Valid() bool {
	switch s {
	case StatusTodo, StatusInProgress, StatusDone, StatusCancelled:
		return true
	}
	return false
}

// Ticket はチケットを表す
type Ticket struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Comments    []Comment `json:"comments,omitempty"`
}

// Comment はチケットに付くコメントを表す
type Comment struct {
	ID        int64     `json:"id"`
	TicketID  int64     `json:"ticket_id"`
	Body      string    `json:"body"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

// ListItem はチケット一覧用の軽量表現
type ListItem struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	Status       Status    `json:"status"`
	CommentCount int       `json:"comment_count"`
	UpdatedAt    time.Time `json:"updated_at"`
}
