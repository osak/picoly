package auth

import (
	"testing"

	"github.com/osak/picoly/internal/model"
)

func TestCanManageTickets(t *testing.T) {
	tests := []struct {
		role model.Role
		can  bool
	}{
		{model.RoleGod, true},
		{model.RoleAdmin, true},
		{model.RoleWorker, false},
	}
	for _, tt := range tests {
		user := &model.User{ID: "test", Role: tt.role}
		if got := CanManageTickets(user); got != tt.can {
			t.Errorf("CanManageTickets(%q) = %v, want %v", tt.role, got, tt.can)
		}
	}
}

func TestCanWork(t *testing.T) {
	roles := []model.Role{model.RoleGod, model.RoleAdmin, model.RoleWorker}
	for _, role := range roles {
		user := &model.User{ID: "test", Role: role}
		if !CanWork(user) {
			t.Errorf("CanWork(%q) = false, want true", role)
		}
	}
}

func TestCanChangeStatusOfInProgress(t *testing.T) {
	inProgressTicket := &model.Ticket{Status: model.StatusInProgress, Assignee: "alice"}
	todoTicket       := &model.Ticket{Status: model.StatusTodo, Assignee: "alice"}

	tests := []struct {
		name   string
		actor  *model.User
		ticket *model.Ticket
		can    bool
	}{
		{"todo ticket allows anyone",      &model.User{ID: "bob",   Role: model.RoleWorker}, todoTicket,       true},
		{"god can always change",          &model.User{ID: "god",   Role: model.RoleGod},   inProgressTicket, true},
		{"admin can always change",        &model.User{ID: "admin", Role: model.RoleAdmin},  inProgressTicket, true},
		{"assignee can change",            &model.User{ID: "alice", Role: model.RoleWorker}, inProgressTicket, true},
		{"non-assignee worker blocked",    &model.User{ID: "bob",   Role: model.RoleWorker}, inProgressTicket, false},
		{"worker blocked if no assignee",  &model.User{ID: "bob",   Role: model.RoleWorker}, &model.Ticket{Status: model.StatusInProgress, Assignee: ""}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanChangeStatusOfInProgress(tt.actor, tt.ticket); got != tt.can {
				t.Errorf("CanChangeStatusOfInProgress() = %v, want %v", got, tt.can)
			}
		})
	}
}
