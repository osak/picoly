package export

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/osak/picoly/internal/model"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

var funcMap = template.FuncMap{
	"formatTime": func(t time.Time) string {
		return t.Format(time.RFC3339)
	},
	"formatDate": func(t time.Time) string {
		return t.Format("2006-01-02")
	},
}

// boardData はboard.md.tmplに渡すデータ
type boardData struct {
	UpdatedAt time.Time
	Tickets   []model.ListItem
}

// Exporter はMarkdownエクスポートを行う
type Exporter struct {
	outputDir string
	board     *template.Template
	ticket    *template.Template
}

// NewExporter はExporterを作成する
func NewExporter(outputDir string) (*Exporter, error) {
	board, err := template.New("board.md.tmpl").Funcs(funcMap).ParseFS(templateFS, "templates/board.md.tmpl")
	if err != nil {
		return nil, fmt.Errorf("parse board template: %w", err)
	}
	ticket, err := template.New("ticket.md.tmpl").Funcs(funcMap).ParseFS(templateFS, "templates/ticket.md.tmpl")
	if err != nil {
		return nil, fmt.Errorf("parse ticket template: %w", err)
	}
	return &Exporter{
		outputDir: outputDir,
		board:     board,
		ticket:    ticket,
	}, nil
}

// ExportBoard はボード全体のboard.mdを書き出す
func (e *Exporter) ExportBoard(tickets []model.ListItem) error {
	if err := os.MkdirAll(e.outputDir, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", e.outputDir, err)
	}
	outPath := filepath.Join(e.outputDir, "board.md")
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create board.md: %w", err)
	}
	defer f.Close()

	data := boardData{
		UpdatedAt: time.Now().UTC(),
		Tickets:   tickets,
	}
	if err := e.board.Execute(f, data); err != nil {
		return fmt.Errorf("execute board template: %w", err)
	}
	return nil
}

// ExportTicket は個別チケットのMarkdownを書き出す
func (e *Exporter) ExportTicket(ticket *model.Ticket) error {
	ticketDir := filepath.Join(e.outputDir, "tickets")
	if err := os.MkdirAll(ticketDir, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", ticketDir, err)
	}
	outPath := filepath.Join(ticketDir, fmt.Sprintf("%d.md", ticket.ID))
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create ticket md: %w", err)
	}
	defer f.Close()

	if err := e.ticket.Execute(f, ticket); err != nil {
		return fmt.Errorf("execute ticket template: %w", err)
	}
	return nil
}

// ExportAll はボード全体と指定チケットのMarkdownを書き出す
func (e *Exporter) ExportAll(tickets []model.ListItem, changed *model.Ticket) error {
	if err := e.ExportBoard(tickets); err != nil {
		return err
	}
	if changed != nil {
		if err := e.ExportTicket(changed); err != nil {
			return err
		}
	}
	return nil
}
