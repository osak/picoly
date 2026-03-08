package model

// Role represents the permission level of a user.
type Role string

const (
	RoleGod    Role = "god"
	RoleAdmin  Role = "admin"
	RoleWorker Role = "worker"
)

// Valid reports whether the role is a valid value.
func (r Role) Valid() bool {
	switch r {
	case RoleGod, RoleAdmin, RoleWorker:
		return true
	}
	return false
}

// User represents a Picoly user with an assigned role.
type User struct {
	ID   string `json:"id"`
	Role Role   `json:"role"`
}
