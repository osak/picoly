package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/osak/picoly/internal/db"
	"github.com/osak/picoly/internal/model"
)

// UserStore はユーザーのCRUDを行う
type UserStore struct {
	db *db.DB
}

// NewUserStore はUserStoreを作成する
func NewUserStore(d *db.DB) *UserStore {
	return &UserStore{db: d}
}

// GetOrCreate はユーザーを取得、存在しなければworkerロールで作成する
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

	// 存在しなければworkerで作成
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO users (id, role) VALUES (?, ?)`,
		userID, string(model.RoleWorker),
	)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}
	return &model.User{ID: userID, Role: model.RoleWorker}, nil
}

// EnsureGodUser はユーザーが存在しない場合のみgodロールで作成する
func (s *UserStore) EnsureGodUser(ctx context.Context, userID string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO users (id, role) VALUES (?, ?)`,
		userID, string(model.RoleGod),
	)
	if err != nil {
		return fmt.Errorf("ensure god user: %w", err)
	}
	return nil
}
