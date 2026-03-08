package store

import (
	"context"
	"fmt"
	"time"

	"github.com/osak/picoly/internal/db"
	"github.com/osak/picoly/internal/model"
)

// CommentStore はコメントのCRUDを行う
type CommentStore struct {
	db *db.DB
}

// NewCommentStore はCommentStoreを作成する
func NewCommentStore(d *db.DB) *CommentStore {
	return &CommentStore{db: d}
}

// Add はチケットにコメントを追加する
func (s *CommentStore) Add(ctx context.Context, ticketID int64, body, author string) (*model.Comment, error) {
	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339Nano)
	result, err := s.db.ExecContext(ctx,
		`INSERT INTO comments (ticket_id, body, author, created_at) VALUES (?, ?, ?, ?)`,
		ticketID, body, author, nowStr,
	)
	if err != nil {
		return nil, fmt.Errorf("insert comment: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("last insert id: %w", err)
	}
	return &model.Comment{
		ID:        id,
		TicketID:  ticketID,
		Body:      body,
		Author:    author,
		CreatedAt: now,
	}, nil
}

// ListByTicketID はチケットのコメントを時系列昇順で返す
func (s *CommentStore) ListByTicketID(ctx context.Context, ticketID int64) ([]model.Comment, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, ticket_id, body, author, created_at FROM comments WHERE ticket_id = ? ORDER BY created_at ASC`,
		ticketID,
	)
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	defer rows.Close()

	var comments []model.Comment
	for rows.Next() {
		var c model.Comment
		var createdAtStr string
		if err := rows.Scan(&c.ID, &c.TicketID, &c.Body, &c.Author, &createdAtStr); err != nil {
			return nil, fmt.Errorf("scan comment: %w", err)
		}
		t, err := time.Parse(time.RFC3339Nano, createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("parse created_at: %w", err)
		}
		c.CreatedAt = t
		comments = append(comments, c)
	}
	return comments, rows.Err()
}
