package cmd

import (
	"encoding/json"
	"log/slog"
	"os"
	"time"
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

// parseSince parses a --since timestamp, accepting both RFC3339Nano and RFC3339 formats.
func parseSince(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, s)
}

// WriteError writes the error as a JSON object to stdout and logs it to stderr.
func WriteError(err error) {
	slog.Error("command error", "err", err)
	resp := errorResponse{Error: err.Error()}
	enc := json.NewEncoder(os.Stdout)
	enc.Encode(resp)
}
