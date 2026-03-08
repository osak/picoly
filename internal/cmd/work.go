package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/osak/picoly/internal/lock"
	"github.com/osak/picoly/internal/model"
	"github.com/osak/picoly/internal/store"
)

func runWork(args []string) error {
	// The ticket ID is the first positional argument; remaining args are flags.
	if len(args) < 1 {
		WriteError(fmt.Errorf("usage: picoly work <id> [options]"))
		return fmt.Errorf("ticket ID required")
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		WriteError(fmt.Errorf("invalid ticket ID: %s", args[0]))
		return err
	}

	fs := flag.NewFlagSet("work", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	statusStr := fs.String("status", "", "New status (todo|in_progress|done|cancelled)")
	comment := fs.String("comment", "", "Comment to add")
	sinceStr := fs.String("since", "", "Expected updated_at (RFC3339) for optimistic locking")

	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	if *statusStr == "" && *comment == "" {
		WriteError(fmt.Errorf("at least one of --status or --comment is required"))
		return fmt.Errorf("no operation specified")
	}

	var status model.Status
	if *statusStr != "" {
		status = model.Status(*statusStr)
		if !status.Valid() {
			WriteError(model.ErrInvalidStatus)
			return model.ErrInvalidStatus
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

	if *sinceStr == "" && app.User.Role != model.RoleGod {
		err := fmt.Errorf("--since is required for non-god users to prevent race conditions")
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

	if status != "" {
		if _, err := app.Tickets.UpdateStatus(ctx, id, status, sinceAt); err != nil {
			WriteError(err)
			return err
		}
	}

	if *comment != "" {
		if _, err := app.Comments.Add(ctx, id, *comment, cfg.UserID); err != nil {
			WriteError(err)
			return err
		}
	}

	ticket, err := app.Tickets.GetByIDWithComments(ctx, id)
	if err != nil {
		WriteError(err)
		return err
	}

	allTickets, _ := app.Tickets.List(ctx, store.ListOptions{})
	app.Exporter.ExportAll(allTickets, ticket)

	return WriteJSON(ticket)
}
