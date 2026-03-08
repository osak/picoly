package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/osak/picoly/internal/auth"
	"github.com/osak/picoly/internal/db"
	"github.com/osak/picoly/internal/export"
	"github.com/osak/picoly/internal/model"
	"github.com/osak/picoly/internal/store"
)

// Config holds the runtime configuration for picoly.
type Config struct {
	DBPath    string
	ExportDir string
	UserID    string
}

// AppContext holds all dependencies needed to execute a command.
type AppContext struct {
	Config   *Config
	DB       *db.DB
	Tickets  *store.TicketStore
	Comments *store.CommentStore
	Users    *store.UserStore
	Exporter *export.Exporter
	User     *model.User
}

// LoadConfig reads configuration from environment variables.
func LoadConfig() (*Config, error) {
	userID := os.Getenv("PICOLY_USER_ID")
	if userID == "" {
		return nil, model.ErrMissingUserID
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

// NewAppContext initializes all dependencies from the given configuration.
func NewAppContext(cfg *Config) (*AppContext, error) {
	d, err := db.Open(cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	us := store.NewUserStore(d)
	// Register the current user as god on first run.
	if err := us.EnsureGodUser(context.Background(), cfg.UserID); err != nil {
		d.Close()
		return nil, fmt.Errorf("ensure god user: %w", err)
	}

	user, err := us.GetOrCreate(context.Background(), cfg.UserID)
	if err != nil {
		d.Close()
		return nil, fmt.Errorf("get user: %w", err)
	}

	exp, err := export.NewExporter(cfg.ExportDir)
	if err != nil {
		d.Close()
		return nil, fmt.Errorf("new exporter: %w", err)
	}

	return &AppContext{
		Config:   cfg,
		DB:       d,
		Tickets:  store.NewTicketStore(d),
		Comments: store.NewCommentStore(d),
		Users:    us,
		Exporter: exp,
		User:     user,
	}, nil
}

// Run dispatches args to the appropriate subcommand handler.
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

// requireAdmin returns ErrUnauthorized if the user does not have admin privileges.
func requireAdmin(user *model.User) error {
	if !auth.CanManageTickets(user) {
		return model.ErrUnauthorized
	}
	return nil
}
