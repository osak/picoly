// integration-test is a standalone integration test runner for the picoly CLI.
// It requires a pre-built picoly binary and runs end-to-end scenarios against it.
//
// Usage:
//
//	go run ./cmd/integration-test [path/to/picoly]
//
// The binary path defaults to ./bin/picoly.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/osak/picoly/internal/db"
	"github.com/osak/picoly/internal/model"
)

var binaryPath string

func main() {
	binaryPath = "./bin/picoly"
	if len(os.Args) > 1 {
		binaryPath = os.Args[1]
	}

	if _, err := os.Stat(binaryPath); err != nil {
		fatalf("binary not found at %s: %v\nRun 'just build' first.\n", binaryPath, err)
	}

	tests := []struct {
		name string
		fn   func(*testEnv) error
	}{
		{"AddAndRead", testAddAndRead},
		{"AddWithJSON", testAddWithJSON},
		{"WorkStatusAndList", testWorkStatusAndList},
		{"WorkComment", testWorkComment},
		{"RaceCondition", testRaceCondition},
		{"WorkerCannotAdd", testWorkerCannotAdd},
		{"NonGodRequiresSince", testNonGodRequiresSince},
		{"FirstUserBecomesGod", testFirstUserBecomesGod},
		{"MarkdownExport", testMarkdownExport},
	}

	failed := 0
	for _, tt := range tests {
		env := newTestEnv("alice")
		err := tt.fn(env)
		env.cleanup()
		if err != nil {
			fmt.Printf("FAIL  %s: %v\n", tt.name, err)
			failed++
		} else {
			fmt.Printf("PASS  %s\n", tt.name)
		}
	}

	fmt.Printf("\n%d tests, %d failed\n", len(tests), failed)
	if failed > 0 {
		os.Exit(1)
	}
}

// --- test environment ---

type testEnv struct {
	dir       string
	dbPath    string
	exportDir string
	userID    string
}

func newTestEnv(userID string) *testEnv {
	dir, err := os.MkdirTemp("", "picoly-itest-*")
	if err != nil {
		fatalf("mkdirtemp: %v", err)
	}
	return &testEnv{
		dir:       dir,
		dbPath:    filepath.Join(dir, "picoly.db"),
		exportDir: filepath.Join(dir, "export"),
		userID:    userID,
	}
}

func (e *testEnv) cleanup() {
	os.RemoveAll(e.dir)
}

func (e *testEnv) withUser(userID string) *testEnv {
	return &testEnv{
		dir:       e.dir,
		dbPath:    e.dbPath,
		exportDir: e.exportDir,
		userID:    userID,
	}
}

func (e *testEnv) osEnv() []string {
	return []string{
		"PICOLY_USER_ID=" + e.userID,
		"PICOLY_DB=" + e.dbPath,
		"PICOLY_EXPORT=" + e.exportDir,
		"PATH=" + os.Getenv("PATH"),
	}
}

// run executes picoly with the given args and returns stdout and exit code.
func (e *testEnv) run(args ...string) (string, int) {
	cmd := exec.Command(binaryPath, args...)
	cmd.Env = e.osEnv()
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return string(out) + string(exitErr.Stderr), exitErr.ExitCode()
		}
		fatalf("exec: %v", err)
	}
	return string(out), 0
}

func (e *testEnv) mustRun(args ...string) string {
	out, code := e.run(args...)
	if code != 0 {
		fatalf("command %v failed (exit %d):\n%s", args, code, out)
	}
	return out
}

func (e *testEnv) openDB() *db.DB {
	d, err := db.Open(e.dbPath)
	if err != nil {
		fatalf("open DB: %v", err)
	}
	return d
}

// --- helpers ---

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "fatal: "+format+"\n", args...)
	os.Exit(2)
}

func check(condition bool, format string, args ...any) error {
	if !condition {
		return fmt.Errorf(format, args...)
	}
	return nil
}

func parseTicket(s string) (model.Ticket, error) {
	var t model.Ticket
	if err := json.Unmarshal([]byte(strings.TrimSpace(s)), &t); err != nil {
		return t, fmt.Errorf("parse ticket JSON: %w\nraw: %s", err, s)
	}
	return t, nil
}

// --- test cases ---

// testAddAndRead verifies that a ticket created with add can be retrieved with read.
func testAddAndRead(e *testEnv) error {
	out := e.mustRun("add", "--title", "Hello", "--desc", "World")
	created, err := parseTicket(out)
	if err != nil {
		return err
	}
	if err := check(created.ID != 0, "expected non-zero ID"); err != nil {
		return err
	}
	if err := check(created.Title == "Hello", "Title = %q, want %q", created.Title, "Hello"); err != nil {
		return err
	}
	if err := check(created.Status == model.StatusTodo, "Status = %q, want %q", created.Status, model.StatusTodo); err != nil {
		return err
	}

	out = e.mustRun("read", "1")
	read, err := parseTicket(out)
	if err != nil {
		return err
	}
	if err := check(read.Title == created.Title && read.Description == created.Description,
		"read ticket differs from created: %+v vs %+v", read, created); err != nil {
		return err
	}
	return nil
}

// testAddWithJSON verifies the --json flag on add.
func testAddWithJSON(e *testEnv) error {
	out := e.mustRun("add", "--json", `{"title":"JSON title","description":"JSON desc"}`)
	ticket, err := parseTicket(out)
	if err != nil {
		return err
	}
	if err := check(ticket.Title == "JSON title", "Title = %q, want %q", ticket.Title, "JSON title"); err != nil {
		return err
	}
	return check(ticket.Description == "JSON desc", "Description = %q, want %q", ticket.Description, "JSON desc")
}

// testWorkStatusAndList verifies that work --status updates the ticket and list --status filters correctly.
func testWorkStatusAndList(e *testEnv) error {
	e.mustRun("add", "--title", "T1")
	e.mustRun("add", "--title", "T2")
	e.mustRun("work", "1", "--status", "in_progress") // alice is god; --since not required

	out := e.mustRun("list", "--status", "in_progress")
	var resp struct {
		Tickets []model.ListItem `json:"tickets"`
		Total   int              `json:"total"`
	}
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		return fmt.Errorf("parse list JSON: %w", err)
	}
	if err := check(resp.Total == 1, "Total = %d, want 1", resp.Total); err != nil {
		return err
	}
	return check(resp.Tickets[0].ID == 1, "Tickets[0].ID = %d, want 1", resp.Tickets[0].ID)
}

// testWorkComment verifies that a comment is appended and visible in read.
func testWorkComment(e *testEnv) error {
	e.mustRun("add", "--title", "T1")
	e.mustRun("work", "1", "--comment", "progress note")

	out := e.mustRun("read", "1")
	ticket, err := parseTicket(out)
	if err != nil {
		return err
	}
	if err := check(len(ticket.Comments) == 1, "len(Comments) = %d, want 1", len(ticket.Comments)); err != nil {
		return err
	}
	return check(ticket.Comments[0].Body == "progress note",
		"Comment body = %q, want %q", ticket.Comments[0].Body, "progress note")
}

// testRaceCondition verifies that --since detects concurrent updates.
func testRaceCondition(e *testEnv) error {
	e.mustRun("add", "--title", "T1")

	out := e.mustRun("read", "1")
	ticket, err := parseTicket(out)
	if err != nil {
		return err
	}
	oldTS := ticket.UpdatedAt.UTC().Format(time.RFC3339Nano)

	e.mustRun("work", "1", "--status", "in_progress", "--since", oldTS)

	// second update with the same (now stale) timestamp must fail
	out, code := e.run("work", "1", "--status", "done", "--since", oldTS)
	if err := check(code != 0, "expected failure for stale --since, got exit 0"); err != nil {
		return err
	}
	return check(strings.Contains(out, "race condition"),
		"expected race condition error, got: %s", out)
}

// testWorkerCannotAdd verifies that a worker-role user cannot create tickets.
func testWorkerCannotAdd(e *testEnv) error {
	e.mustRun("add", "--title", "seed") // alice creates DB and becomes god

	bob := e.withUser("bob")
	bob.mustRun("list") // register bob as worker

	out, code := bob.run("add", "--title", "Should fail")
	if err := check(code != 0, "expected worker add to fail, got exit 0"); err != nil {
		return err
	}
	return check(strings.Contains(out, "unauthorized"),
		"expected unauthorized error, got: %s", out)
}

// testNonGodRequiresSince verifies that non-god users must supply --since for work.
func testNonGodRequiresSince(e *testEnv) error {
	e.mustRun("add", "--title", "T1")

	bob := e.withUser("bob")
	bob.mustRun("list") // register bob as worker

	// work without --since must fail for bob
	out, code := bob.run("work", "1", "--status", "in_progress")
	if err := check(code != 0, "expected worker work without --since to fail, got exit 0"); err != nil {
		return err
	}
	if err := check(strings.Contains(out, "--since is required"),
		"expected --since required error, got: %s", out); err != nil {
		return err
	}

	// alice (god) can work without --since
	_, code = e.run("work", "1", "--status", "in_progress")
	return check(code == 0, "expected god to work without --since, got exit %d", code)
}

// testFirstUserBecomesGod verifies that the first user gets god role and subsequent users get worker,
// reading the roles directly from the database.
func testFirstUserBecomesGod(e *testEnv) error {
	e.mustRun("add", "--title", "seed") // alice becomes god

	bob := e.withUser("bob")
	bob.mustRun("list") // register bob as worker

	d := e.openDB()
	defer d.Close()

	rows, err := d.Query("SELECT id, role FROM users ORDER BY id")
	if err != nil {
		return fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	type userRow struct{ id, role string }
	var users []userRow
	for rows.Next() {
		var u userRow
		if err := rows.Scan(&u.id, &u.role); err != nil {
			return fmt.Errorf("scan: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	if err := check(len(users) == 2, "expected 2 users, got %d: %v", len(users), users); err != nil {
		return err
	}
	if err := check(users[0].id == "alice" && users[0].role == string(model.RoleGod),
		"alice: got role %q, want %q", users[0].role, model.RoleGod); err != nil {
		return err
	}
	return check(users[1].id == "bob" && users[1].role == string(model.RoleWorker),
		"bob: got role %q, want %q", users[1].role, model.RoleWorker)
}

// testMarkdownExport verifies that board.md and tickets/<id>.md are generated after add.
func testMarkdownExport(e *testEnv) error {
	e.mustRun("add", "--title", "Export test", "--desc", "Check files")

	boardPath := filepath.Join(e.exportDir, "board.md")
	if _, err := os.Stat(boardPath); err != nil {
		return fmt.Errorf("board.md not found: %v", err)
	}
	ticketPath := filepath.Join(e.exportDir, "tickets", "1.md")
	if _, err := os.Stat(ticketPath); err != nil {
		return fmt.Errorf("tickets/1.md not found: %v", err)
	}

	content, err := os.ReadFile(boardPath)
	if err != nil {
		return err
	}
	return check(strings.Contains(string(content), "Export test"),
		"board.md does not contain ticket title")
}
