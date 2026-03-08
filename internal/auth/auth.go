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

// CanChangeStatusOfInProgress reports whether actor is allowed to change the status
// of a ticket that is currently in_progress.
// God and admin may always change it; other users may only change it if they are the assignee.
// If the ticket is not in_progress, the check always passes.
func CanChangeStatusOfInProgress(actor *model.User, ticket *model.Ticket) bool {
	if ticket.Status != model.StatusInProgress {
		return true
	}
	if actor.Role == model.RoleGod || actor.Role == model.RoleAdmin {
		return true
	}
	return ticket.Assignee != "" && actor.ID == ticket.Assignee
}
