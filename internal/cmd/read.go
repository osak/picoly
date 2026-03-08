package cmd

import (
	"context"
	"fmt"
	"strconv"
)

func runRead(args []string) error {
	if len(args) < 1 {
		WriteError(fmt.Errorf("usage: picoly read <id>"))
		return fmt.Errorf("ticket ID required")
	}
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		WriteError(fmt.Errorf("invalid ticket ID: %s", args[0]))
		return err
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
	ticket, err := app.Tickets.GetByIDWithComments(ctx, id)
	if err != nil {
		WriteError(err)
		return err
	}

	return WriteJSON(ticket)
}
