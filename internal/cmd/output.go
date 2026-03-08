package cmd

import (
	"encoding/json"
	"log/slog"
	"os"
)

// WriteJSON encodes v as JSON and writes it to stdout.
func WriteJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	return enc.Encode(v)
}

// errorResponse is the JSON shape returned on error.
type errorResponse struct {
	Error string `json:"error"`
}

// WriteError writes the error as a JSON object to stdout and logs it to stderr.
func WriteError(err error) {
	slog.Error("command error", "err", err)
	resp := errorResponse{Error: err.Error()}
	enc := json.NewEncoder(os.Stdout)
	enc.Encode(resp)
}
