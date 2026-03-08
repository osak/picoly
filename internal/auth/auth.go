package auth

import "github.com/osak/picoly/internal/model"

// CanManageTickets reports whether the user is allowed to add or edit tickets (requires admin role or above).
func CanManageTickets(user *model.User) bool {
	return user.Role == model.RoleGod || user.Role == model.RoleAdmin
}

// CanWork reports whether the user is allowed to update ticket status or add comments (all roles permitted).
func CanWork(user *model.User) bool {
	return true
}
