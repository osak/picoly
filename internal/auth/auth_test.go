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
