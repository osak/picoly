package main

import (
	"log/slog"
	"os"

	"github.com/osak/picoly/internal/cmd"
)

func main() {
	if err := cmd.Run(os.Args[1:]); err != nil {
		slog.Error("command failed", "err", err)
		os.Exit(1)
	}
}
