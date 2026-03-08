package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/osak/picoly/internal/db"
	"github.com/osak/picoly/internal/model"
)

// UserStore handles CRUD operations for users.
type UserStore struct {
	db *db.DB
}

// NewUserStore creates a new UserStore.
func NewUserStore(d *db.DB) *UserStore {
	return &UserStore{db: d}
}

// GetOrCreate returns the user with the given ID, creating one with the worker role if it does not exist.
func (s *UserStore) GetOrCreate(ctx context.Context, userID string) (*model.User, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, role FROM users WHERE id = ?`,
		userID,
	)
	var u model.User
	err := row.Scan(&u.ID, &u.Role)
	if err == nil {
		return &u, nil
	}
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("get user: %w", err)
	}

	// user not found; create with worker role
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO users (id, role) VALUES (?, ?)`,
		userID, string(model.RoleWorker),
	)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}
	return &model.User{ID: userID, Role: model.RoleWorker}, nil
}

// EnsureGodUser inserts the given user with the god role only if the users table is empty.
// This ensures that only the very first user to access a new database receives the god role.
func (s *UserStore) EnsureGodUser(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO users (id, role) SELECT ?, 'god' WHERE NOT EXISTS (SELECT 1 FROM users)`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("ensure god user: %w", err)
	}
	return nil
}
