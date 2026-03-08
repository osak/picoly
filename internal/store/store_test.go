package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/osak/picoly/internal/db"
	"github.com/osak/picoly/internal/model"
)

func newTestStores(t *testing.T) (*TicketStore, *CommentStore, *UserStore) {
	t.Helper()
	d := db.NewTestDB(t)
	return NewTicketStore(d), NewCommentStore(d), NewUserStore(d)
}

// --- TicketStore tests ---

func TestTicketCreate(t *testing.T) {
	ts, _, _ := newTestStores(t)
	ctx := context.Background()

	ticket, err := ts.Create(ctx, "Test Ticket", "Description", "alice")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if ticket.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if ticket.Title != "Test Ticket" {
		t.Errorf("Title = %q, want %q", ticket.Title, "Test Ticket")
	}
	if ticket.Status != model.StatusTodo {
		t.Errorf("Status = %q, want %q", ticket.Status, model.StatusTodo)
	}
}

func TestTicketGetByID_NotFound(t *testing.T) {
	ts, _, _ := newTestStores(t)
	ctx := context.Background()

	_, err := ts.GetByID(ctx, 999)
	if !errors.Is(err, model.ErrNotFound) {
		t.Errorf("GetByID(999) error = %v, want ErrNotFound", err)
	}
}

func TestTicketUpdateStatus_RaceCondition(t *testing.T) {
	ts, _, _ := newTestStores(t)
	ctx := context.Background()

	ticket, err := ts.Create(ctx, "T1", "", "alice")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// 古いタイムスタンプで更新しようとするとレースコンディションエラー
	oldTime := ticket.UpdatedAt.Add(-1 * time.Second)
	_, err = ts.UpdateStatus(ctx, ticket.ID, model.StatusInProgress, oldTime)
	if !errors.Is(err, model.ErrRaceCondition) {
		t.Errorf("UpdateStatus with old since = %v, want ErrRaceCondition", err)
	}
}

func TestTicketList(t *testing.T) {
	ts, _, _ := newTestStores(t)
	ctx := context.Background()

	ts.Create(ctx, "T1", "", "alice")
	ts.Create(ctx, "T2", "", "alice")
	ts.UpdateStatus(ctx, 1, model.StatusInProgress, time.Time{})

	items, err := ts.List(ctx, ListOptions{StatusFilter: model.StatusInProgress})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("len(items) = %d, want 1", len(items))
	}
}

// --- CommentStore tests ---

func TestCommentAdd(t *testing.T) {
	ts, cs, _ := newTestStores(t)
	ctx := context.Background()

	ticket, _ := ts.Create(ctx, "T1", "", "alice")
	comment, err := cs.Add(ctx, ticket.ID, "Hello", "bob")
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if comment.ID == 0 {
		t.Error("expected non-zero comment ID")
	}
}

func TestCommentListOrder(t *testing.T) {
	ts, cs, _ := newTestStores(t)
	ctx := context.Background()

	ticket, _ := ts.Create(ctx, "T1", "", "alice")
	cs.Add(ctx, ticket.ID, "first", "alice")
	time.Sleep(time.Millisecond)
	cs.Add(ctx, ticket.ID, "second", "alice")

	comments, err := cs.ListByTicketID(ctx, ticket.ID)
	if err != nil {
		t.Fatalf("ListByTicketID: %v", err)
	}
	if len(comments) != 2 {
		t.Fatalf("len(comments) = %d, want 2", len(comments))
	}
	if comments[0].Body != "first" {
		t.Errorf("comments[0].Body = %q, want %q", comments[0].Body, "first")
	}
}

// --- UserStore tests ---

func TestUserGetOrCreate(t *testing.T) {
	_, _, us := newTestStores(t)
	ctx := context.Background()

	u1, err := us.GetOrCreate(ctx, "alice")
	if err != nil {
		t.Fatalf("GetOrCreate: %v", err)
	}
	if u1.Role != model.RoleWorker {
		t.Errorf("Role = %q, want %q", u1.Role, model.RoleWorker)
	}

	// 再度呼んでも同じユーザーが返る
	u2, err := us.GetOrCreate(ctx, "alice")
	if err != nil {
		t.Fatalf("GetOrCreate (2nd): %v", err)
	}
	if u2.ID != u1.ID || u2.Role != u1.Role {
		t.Errorf("second GetOrCreate returned different user")
	}
}

func TestUserEnsureGodUser(t *testing.T) {
	_, _, us := newTestStores(t)
	ctx := context.Background()

	if err := us.EnsureGodUser(ctx, "admin"); err != nil {
		t.Fatalf("EnsureGodUser: %v", err)
	}
	u, err := us.GetOrCreate(ctx, "admin")
	if err != nil {
		t.Fatalf("GetOrCreate: %v", err)
	}
	// EnsureGodUser後はgodロール
	// GetOrCreateはSELECTを先に試みるので、既存ユーザーはそのまま返す
	if u.Role != model.RoleGod {
		t.Errorf("Role = %q, want %q", u.Role, model.RoleGod)
	}
}
