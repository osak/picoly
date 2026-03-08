package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/osak/picoly/internal/db"
	"github.com/osak/picoly/internal/model"
)

// TicketStore handles CRUD operations for tickets.
type TicketStore struct {
	db *db.DB
}

// NewTicketStore creates a new TicketStore.
func NewTicketStore(d *db.DB) *TicketStore {
	return &TicketStore{db: d}
}

// Create inserts a new ticket and returns it with the assigned ID.
func (s *TicketStore) Create(ctx context.Context, title, description, createdBy string) (*model.Ticket, error) {
	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339Nano)
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO tickets (title, description, status, created_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		title, description, string(model.StatusTodo), createdBy, nowStr, nowStr,
	)
	if err != nil {
		return nil, fmt.Errorf("insert ticket: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("last insert id: %w", err)
	}
	return &model.Ticket{
		ID:          id,
		Title:       title,
		Description: description,
		Status:      model.StatusTodo,
		CreatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// GetByID retrieves a ticket by ID without its comments.
func (s *TicketStore) GetByID(ctx context.Context, id int64) (*model.Ticket, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, title, description, status, created_by, created_at, updated_at FROM tickets WHERE id = ?`,
		id,
	)
	return scanTicket(row)
}

// GetByIDWithComments retrieves a ticket together with all its comments.
func (s *TicketStore) GetByIDWithComments(ctx context.Context, id int64) (*model.Ticket, error) {
	ticket, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	cs := NewCommentStore(s.db)
	comments, err := cs.ListByTicketID(ctx, id)
	if err != nil {
		return nil, err
	}
	ticket.Comments = comments
	return ticket, nil
}

// Update modifies the title and description of a ticket.
// If sinceAt is non-zero, it returns ErrRaceCondition when updated_at does not match.
func (s *TicketStore) Update(ctx context.Context, id int64, title, description string, sinceAt time.Time) (*model.Ticket, error) {
	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339Nano)

	var result sql.Result
	var err error
	if sinceAt.IsZero() {
		result, err = s.db.ExecContext(ctx,
			`UPDATE tickets SET title = ?, description = ?, updated_at = ? WHERE id = ?`,
			title, description, nowStr, id,
		)
	} else {
		sinceAtStr := sinceAt.UTC().Format(time.RFC3339Nano)
		result, err = s.db.ExecContext(ctx,
			`UPDATE tickets SET title = ?, description = ?, updated_at = ? WHERE id = ? AND updated_at = ?`,
			title, description, nowStr, id, sinceAtStr,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("update ticket: %w", err)
	}
	return s.checkRowsAffected(ctx, result, id)
}

// UpdateStatus changes the status of a ticket.
// If sinceAt is non-zero, it returns ErrRaceCondition when updated_at does not match.
func (s *TicketStore) UpdateStatus(ctx context.Context, id int64, status model.Status, sinceAt time.Time) (*model.Ticket, error) {
	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339Nano)

	var result sql.Result
	var err error
	if sinceAt.IsZero() {
		result, err = s.db.ExecContext(ctx,
			`UPDATE tickets SET status = ?, updated_at = ? WHERE id = ?`,
			string(status), nowStr, id,
		)
	} else {
		sinceAtStr := sinceAt.UTC().Format(time.RFC3339Nano)
		result, err = s.db.ExecContext(ctx,
			`UPDATE tickets SET status = ?, updated_at = ? WHERE id = ? AND updated_at = ?`,
			string(status), nowStr, id, sinceAtStr,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("update status: %w", err)
	}
	return s.checkRowsAffected(ctx, result, id)
}

// checkRowsAffected inspects the result of an UPDATE and returns ErrRaceCondition or ErrNotFound when no rows were affected.
func (s *TicketStore) checkRowsAffected(ctx context.Context, result sql.Result, id int64) (*model.Ticket, error) {
	n, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		// verify whether the ticket exists at all
		_, err := s.GetByID(ctx, id)
		if err != nil {
			return nil, err // propagate ErrNotFound
		}
		return nil, model.ErrRaceCondition
	}
	return s.GetByID(ctx, id)
}

// ListOptions holds filter and sort parameters for listing tickets.
type ListOptions struct {
	StatusFilter model.Status
	SortBy       string // "id", "updated_at", or "status"
	Ascending    bool
}

// List returns tickets matching the given filter and sort options.
func (s *TicketStore) List(ctx context.Context, opts ListOptions) ([]model.ListItem, error) {
	query := `
		SELECT t.id, t.title, t.status, COUNT(c.id) as comment_count, t.updated_at
		FROM tickets t
		LEFT JOIN comments c ON c.ticket_id = t.id
	`
	args := []any{}
	if opts.StatusFilter != "" {
		query += " WHERE t.status = ?"
		args = append(args, string(opts.StatusFilter))
	}
	query += " GROUP BY t.id"

	// build ORDER BY clause
	sortCol := "t.id"
	switch opts.SortBy {
	case "updated_at":
		sortCol = "t.updated_at"
	case "status":
		sortCol = "t.status"
	}
	dir := "ASC"
	if !opts.Ascending && opts.SortBy != "" {
		dir = "DESC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s", sortCol, dir)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list tickets: %w", err)
	}
	defer rows.Close()

	var items []model.ListItem
	for rows.Next() {
		var item model.ListItem
		var updatedAtStr string
		if err := rows.Scan(&item.ID, &item.Title, &item.Status, &item.CommentCount, &updatedAtStr); err != nil {
			return nil, fmt.Errorf("scan list item: %w", err)
		}
		t, err := time.Parse(time.RFC3339Nano, updatedAtStr)
		if err != nil {
			return nil, fmt.Errorf("parse updated_at: %w", err)
		}
		item.UpdatedAt = t
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanTicket(row *sql.Row) (*model.Ticket, error) {
	var t model.Ticket
	var createdAtStr, updatedAtStr string
	err := row.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.CreatedBy, &createdAtStr, &updatedAtStr)
	if err == sql.ErrNoRows {
		return nil, model.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan ticket: %w", err)
	}
	createdAt, err := time.Parse(time.RFC3339Nano, createdAtStr)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	updatedAt, err := time.Parse(time.RFC3339Nano, updatedAtStr)
	if err != nil {
		return nil, fmt.Errorf("parse updated_at: %w", err)
	}
	t.CreatedAt = createdAt
	t.UpdatedAt = updatedAt
	return &t, nil
}
