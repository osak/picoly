package model

// Role はユーザーのロールを表す
type Role string

const (
	RoleGod    Role = "god"
	RoleAdmin  Role = "admin"
	RoleWorker Role = "worker"
)

// Valid はロールが有効な値かどうかを返す
func (r Role) Valid() bool {
	switch r {
	case RoleGod, RoleAdmin, RoleWorker:
		return true
	}
	return false
}

// User はユーザーを表す
type User struct {
	ID   string `json:"id"`
	Role Role   `json:"role"`
}
