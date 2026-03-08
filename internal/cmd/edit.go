package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/osak/picoly/internal/lock"
	"github.com/osak/picoly/internal/store"
)

type editInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func runEdit(args []string) error {
	// The ticket ID is the first positional argument; remaining args are flags.
	if len(args) < 1 {
		WriteError(fmt.Errorf("usage: picoly edit <id> [options]"))
		return fmt.Errorf("ticket ID required")
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		WriteError(fmt.Errorf("invalid ticket ID: %s", args[0]))
		return err
	}

	fs := flag.NewFlagSet("edit", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	title := fs.String("title", "", "New ticket title")
	desc := fs.String("desc", "", "New ticket description")
	jsonStr := fs.String("json", "", "JSON input (overrides --title and --desc)")
	sinceStr := fs.String("since", "", "Expected updated_at (RFC3339) for optimistic locking")

	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	input := editInput{Title: *title, Description: *desc}
	if *jsonStr != "" {
		if err := json.Unmarshal([]byte(*jsonStr), &input); err != nil {
			WriteError(fmt.Errorf("parse --json: %w", err))
			return err
		}
	}

	var sinceAt time.Time
	if *sinceStr != "" {
		sinceAt, err = time.Parse(time.RFC3339, *sinceStr)
		if err != nil {
			WriteError(fmt.Errorf("parse --since: %w", err))
			return err
		}
	}

	cfg, err := LoadConfig()
	if err != nil {
		WriteError(err)
		return err
	}
	app, err := NewAppContext(cfg)
	if err != nil {
		WriteError(err)
		return err
	}
	defer app.DB.Close()

	if err := requireAdmin(app.User); err != nil {
		WriteError(err)
		return err
	}

	l, err := lock.Acquire(cfg.DBPath, 30*time.Second)
	if err != nil {
		WriteError(err)
		return err
	}
	defer l.Release()

	ctx := context.Background()
	ticket, err := app.Tickets.Update(ctx, id, input.Title, input.Description, sinceAt)
	if err != nil {
		WriteError(err)
		return err
	}

	allTickets, _ := app.Tickets.List(ctx, store.ListOptions{})
	app.Exporter.ExportAll(allTickets, ticket)

	return WriteJSON(ticket)
}
