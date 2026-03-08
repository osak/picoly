package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/osak/picoly/internal/model"
	"github.com/osak/picoly/internal/store"
)

type listResponse struct {
	Tickets []model.ListItem `json:"tickets"`
	Total   int              `json:"total"`
}

func runList(args []string) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	statusStr := fs.String("status", "", "Filter by status (todo|in_progress|done|cancelled)")
	sortBy := fs.String("sort", "", "Sort by field (id|updated_at|status)")
	asc := fs.Bool("asc", false, "Sort ascending")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// statusフィルタのバリデーション
	var statusFilter model.Status
	if *statusStr != "" {
		statusFilter = model.Status(*statusStr)
		if !statusFilter.Valid() {
			WriteError(model.ErrInvalidStatus)
			return model.ErrInvalidStatus
		}
	}

	// sortByのバリデーション
	if *sortBy != "" {
		switch *sortBy {
		case "id", "updated_at", "status":
			// OK
		default:
			err := fmt.Errorf("invalid sort field: %s (must be id, updated_at, or status)", *sortBy)
			WriteError(err)
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

	ctx := context.Background()
	opts := store.ListOptions{
		StatusFilter: statusFilter,
		SortBy:       *sortBy,
		Ascending:    *asc,
	}
	tickets, err := app.Tickets.List(ctx, opts)
	if err != nil {
		WriteError(err)
		return err
	}
	if tickets == nil {
		tickets = []model.ListItem{}
	}

	return WriteJSON(listResponse{
		Tickets: tickets,
		Total:   len(tickets),
	})
}
