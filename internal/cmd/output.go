package cmd

import (
	"encoding/json"
	"log/slog"
	"os"
)

// WriteJSON は v を JSON として stdout に書き出す（改行付き）
func WriteJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	return enc.Encode(v)
}

// errorResponse はエラーレスポンスのJSON形式
type errorResponse struct {
	Error string `json:"error"`
}

// WriteError はエラーを JSON として stdout に書き出し、人間向けメッセージを stderr に書く
func WriteError(err error) {
	slog.Error("command error", "err", err)
	resp := errorResponse{Error: err.Error()}
	enc := json.NewEncoder(os.Stdout)
	enc.Encode(resp)
}
