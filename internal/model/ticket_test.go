package model

import "testing"

func TestStatusValid(t *testing.T) {
	tests := []struct {
		status Status
		valid  bool
	}{
		{StatusTodo, true},
		{StatusInProgress, true},
		{StatusDone, true},
		{StatusCancelled, true},
		{Status("invalid"), false},
		{Status(""), false},
	}
	for _, tt := range tests {
		if got := tt.status.Valid(); got != tt.valid {
			t.Errorf("Status(%q).Valid() = %v, want %v", tt.status, got, tt.valid)
		}
	}
}

func TestRoleValid(t *testing.T) {
	tests := []struct {
		role  Role
		valid bool
	}{
		{RoleGod, true},
		{RoleAdmin, true},
		{RoleWorker, true},
		{Role("invalid"), false},
		{Role(""), false},
	}
	for _, tt := range tests {
		if got := tt.role.Valid(); got != tt.valid {
			t.Errorf("Role(%q).Valid() = %v, want %v", tt.role, got, tt.valid)
		}
	}
}
