package cmd

import (
	"errors"
	"fmt"
	"os"
)

// Config はpicolyの設定を保持する
type Config struct {
	DBPath    string
	ExportDir string
	UserID    string
}

// LoadConfig は環境変数から設定を読み込む
func LoadConfig() (*Config, error) {
	userID := os.Getenv("PICOLY_USER_ID")
	if userID == "" {
		return nil, errors.New("PICOLY_USER_ID environment variable is not set")
	}
	dbPath := os.Getenv("PICOLY_DB")
	if dbPath == "" {
		dbPath = "./picoly.db"
	}
	exportDir := os.Getenv("PICOLY_EXPORT")
	if exportDir == "" {
		exportDir = "./picoly"
	}
	return &Config{
		DBPath:    dbPath,
		ExportDir: exportDir,
		UserID:    userID,
	}, nil
}

// Run はサブコマンドをルーティングする
func Run(args []string) error {
	if len(args) == 0 {
		return printUsage()
	}
	subcommand := args[0]
	rest := args[1:]
	switch subcommand {
	case "add":
		return runAdd(rest)
	case "edit":
		return runEdit(rest)
	case "work":
		return runWork(rest)
	case "read":
		return runRead(rest)
	case "list":
		return runList(rest)
	default:
		return fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

func printUsage() error {
	fmt.Fprintln(os.Stderr, `Usage: picoly <command> [options]

Commands:
  add     Add a new ticket
  edit    Edit a ticket
  work    Update ticket status or add a comment
  read    Read a ticket
  list    List tickets`)
	return fmt.Errorf("no command specified")
}
