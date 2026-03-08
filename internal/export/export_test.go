package export

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/osak/picoly/internal/model"
)

func TestExportBoard(t *testing.T) {
	dir := t.TempDir()
	e, err := NewExporter(dir)
	if err != nil {
		t.Fatalf("NewExporter: %v", err)
	}

	tickets := []model.ListItem{
		{ID: 1, Title: "Test Ticket", Status: model.StatusTodo, CommentCount: 0, UpdatedAt: time.Now()},
	}
	if err := e.ExportBoard(tickets); err != nil {
		t.Fatalf("ExportBoard: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "board.md"))
	if err != nil {
		t.Fatalf("read board.md: %v", err)
	}

	if !strings.Contains(string(content), "Test Ticket") {
		t.Error("board.md does not contain ticket title")
	}
	if !strings.Contains(string(content), "Picoly Board") {
		t.Error("board.md does not contain header")
	}
}

func TestExportTicket(t *testing.T) {
	dir := t.TempDir()
	e, err := NewExporter(dir)
	if err != nil {
		t.Fatalf("NewExporter: %v", err)
	}

	now := time.Now()
	ticket := &model.Ticket{
		ID:          1,
		Title:       "My Ticket",
		Description: "Some description",
		Status:      model.StatusInProgress,
		CreatedBy:   "alice",
		CreatedAt:   now,
		UpdatedAt:   now,
		Comments: []model.Comment{
			{ID: 1, TicketID: 1, Body: "A comment", Author: "bob", CreatedAt: now},
		},
	}
	if err := e.ExportTicket(ticket); err != nil {
		t.Fatalf("ExportTicket: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "tickets", "1.md"))
	if err != nil {
		t.Fatalf("read ticket md: %v", err)
	}

	str := string(content)
	if !strings.Contains(str, "My Ticket") {
		t.Error("ticket.md does not contain title")
	}
	if !strings.Contains(str, "A comment") {
		t.Error("ticket.md does not contain comment")
	}
	if !strings.Contains(str, "bob") {
		t.Error("ticket.md does not contain comment author")
	}
}
