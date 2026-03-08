package auth

import "github.com/osak/picoly/internal/model"

// CanManageTickets はadd/edit操作の権限チェック（admin以上が必要）
func CanManageTickets(user *model.User) bool {
	return user.Role == model.RoleGod || user.Role == model.RoleAdmin
}

// CanWork はwork操作の権限チェック（全ロールが可能）
func CanWork(user *model.User) bool {
	return true
}
