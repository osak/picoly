package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/osak/picoly/internal/lock"
	"github.com/osak/picoly/internal/store"
)

type addInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func runAdd(args []string) error {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	title := fs.String("title", "", "Ticket title (required)")
	desc := fs.String("desc", "", "Ticket description")
	jsonStr := fs.String("json", "", "JSON input (overrides --title and --desc)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// parse JSON input if provided
	input := addInput{Title: *title, Description: *desc}
	if *jsonStr != "" {
		if err := json.Unmarshal([]byte(*jsonStr), &input); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	if input.Title == "" {
		WriteError(fmt.Errorf("--title is required"))
		return fmt.Errorf("--title is required")
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
	ticket, err := app.Tickets.Create(ctx, input.Title, input.Description, cfg.UserID, "")
	if err != nil {
		WriteError(err)
		return err
	}

	// export Markdown files after mutation
	allTickets, _ := app.Tickets.List(ctx, store.ListOptions{})
	app.Exporter.ExportAll(allTickets, ticket)

	return WriteJSON(ticket)
}
