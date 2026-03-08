package cmd_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/osak/picoly/internal/model"
)

var binaryPath string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "picoly-build-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmp)

	binaryPath = filepath.Join(tmp, "picoly")
	cmd := exec.Command("go", "build", "-o", binaryPath, "../../cmd/picoly")
	cmd.Env = append(os.Environ(), "GOEXPERIMENT=jsonv2")
	if out, err := cmd.CombinedOutput(); err != nil {
		panic("failed to build picoly: " + string(out))
	}

	os.Exit(m.Run())
}

// testEnv sets up an isolated DB and export directory for a single test.
func testEnv(t *testing.T, userID string) (env []string, dbPath, exportDir string) {
	t.Helper()
	dir := t.TempDir()
	dbPath = filepath.Join(dir, "picoly.db")
	exportDir = filepath.Join(dir, "export")
	env = []string{
		"PICOLY_USER_ID=" + userID,
		"PICOLY_DB=" + dbPath,
		"PICOLY_EXPORT=" + exportDir,
		"PATH=" + os.Getenv("PATH"),
	}
	return
}

func run(t *testing.T, env []string, args ...string) (stdout string, exitCode int) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	cmd.Env = env
	out, err := cmd.Output()
	stdout = string(out)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return stdout + string(exitErr.Stderr), exitErr.ExitCode()
		}
		t.Fatalf("exec error: %v", err)
	}
	return stdout, 0
}

func mustRun(t *testing.T, env []string, args ...string) string {
	t.Helper()
	out, code := run(t, env, args...)
	if code != 0 {
		t.Fatalf("command %v failed (exit %d):\n%s", args, code, out)
	}
	return out
}

func mustFail(t *testing.T, env []string, args ...string) string {
	t.Helper()
	out, code := run(t, env, args...)
	if code == 0 {
		t.Fatalf("command %v expected to fail but succeeded:\n%s", args, out)
	}
	return out
}

func parseTicket(t *testing.T, s string) model.Ticket {
	t.Helper()
	var ticket model.Ticket
	if err := json.Unmarshal([]byte(strings.TrimSpace(s)), &ticket); err != nil {
		t.Fatalf("parse ticket JSON: %v\nraw: %s", err, s)
	}
	return ticket
}

// TestAddAndRead verifies that a ticket created with add can be retrieved with read.
func TestAddAndRead(t *testing.T) {
	env, _, _ := testEnv(t, "alice")

	out := mustRun(t, env, "add", "--title", "Hello", "--desc", "World")
	created := parseTicket(t, out)

	if created.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if created.Title != "Hello" {
		t.Errorf("Title = %q, want %q", created.Title, "Hello")
	}
	if created.Status != model.StatusTodo {
		t.Errorf("Status = %q, want %q", created.Status, model.StatusTodo)
	}

	out = mustRun(t, env, "read", "1")
	read := parseTicket(t, out)

	if read.Title != created.Title || read.Description != created.Description {
		t.Errorf("read ticket differs from created: %+v vs %+v", read, created)
	}
}

// TestAddWithJSON verifies the --json flag on add.
func TestAddWithJSON(t *testing.T) {
	env, _, _ := testEnv(t, "alice")

	out := mustRun(t, env, "add", "--json", `{"title":"JSON title","description":"JSON desc"}`)
	ticket := parseTicket(t, out)

	if ticket.Title != "JSON title" {
		t.Errorf("Title = %q, want %q", ticket.Title, "JSON title")
	}
	if ticket.Description != "JSON desc" {
		t.Errorf("Description = %q, want %q", ticket.Description, "JSON desc")
	}
}

// TestWorkStatusAndList verifies that work --status updates the ticket and list --status filters correctly.
func TestWorkStatusAndList(t *testing.T) {
	env, _, _ := testEnv(t, "alice")

	mustRun(t, env, "add", "--title", "T1")
	mustRun(t, env, "add", "--title", "T2")

	// alice is god, so --since is not required
	mustRun(t, env, "work", "1", "--status", "in_progress")

	out := mustRun(t, env, "list", "--status", "in_progress")
	var resp struct {
		Tickets []model.ListItem `json:"tickets"`
		Total   int              `json:"total"`
	}
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("parse list JSON: %v", err)
	}
	if resp.Total != 1 {
		t.Errorf("Total = %d, want 1", resp.Total)
	}
	if resp.Tickets[0].ID != 1 {
		t.Errorf("Tickets[0].ID = %d, want 1", resp.Tickets[0].ID)
	}
}

// TestWorkComment verifies that a comment is appended and visible in read.
func TestWorkComment(t *testing.T) {
	env, _, _ := testEnv(t, "alice")

	mustRun(t, env, "add", "--title", "T1")
	mustRun(t, env, "work", "1", "--comment", "progress note")

	out := mustRun(t, env, "read", "1")
	ticket := parseTicket(t, out)

	if len(ticket.Comments) != 1 {
		t.Fatalf("len(Comments) = %d, want 1", len(ticket.Comments))
	}
	if ticket.Comments[0].Body != "progress note" {
		t.Errorf("Comment body = %q, want %q", ticket.Comments[0].Body, "progress note")
	}
}

// TestRaceCondition verifies that --since detects concurrent updates.
func TestRaceCondition(t *testing.T) {
	env, _, _ := testEnv(t, "alice")

	mustRun(t, env, "add", "--title", "T1")

	// capture updated_at before the first update
	out := mustRun(t, env, "read", "1")
	ticket := parseTicket(t, out)
	oldTS := ticket.UpdatedAt.UTC().Format(time.RFC3339Nano)

	// first update succeeds
	mustRun(t, env, "work", "1", "--status", "in_progress", "--since", oldTS)

	// second update with the same (now stale) timestamp must fail
	out = mustFail(t, env, "work", "1", "--status", "done", "--since", oldTS)
	if !strings.Contains(out, "race condition") {
		t.Errorf("expected race condition error, got: %s", out)
	}
}

// TestWorkerCannotAdd verifies that a worker-role user cannot create tickets.
func TestWorkerCannotAdd(t *testing.T) {
	envAlice, _, _ := testEnv(t, "alice")

	// register bob as worker by having him list (alice's DB)
	envBob := append([]string(nil), envAlice...)
	for i, e := range envBob {
		if strings.HasPrefix(e, "PICOLY_USER_ID=") {
			envBob[i] = "PICOLY_USER_ID=bob"
		}
	}
	mustRun(t, envAlice, "add", "--title", "seed") // ensure DB exists with alice as god
	mustRun(t, envBob, "list")                     // register bob as worker

	out := mustFail(t, envBob, "add", "--title", "Should fail")
	if !strings.Contains(out, "unauthorized") {
		t.Errorf("expected unauthorized error, got: %s", out)
	}
}

// TestNonGodRequiresSince verifies that non-god users must supply --since for work and edit.
func TestNonGodRequiresSince(t *testing.T) {
	envAlice, _, _ := testEnv(t, "alice")
	envBob := append([]string(nil), envAlice...)
	for i, e := range envBob {
		if strings.HasPrefix(e, "PICOLY_USER_ID=") {
			envBob[i] = "PICOLY_USER_ID=bob"
		}
	}

	mustRun(t, envAlice, "add", "--title", "T1")
	mustRun(t, envBob, "list") // register bob as worker

	// work without --since must fail for bob (worker)
	out := mustFail(t, envBob, "work", "1", "--status", "in_progress")
	if !strings.Contains(out, "--since is required") {
		t.Errorf("expected --since required error, got: %s", out)
	}

	// god (alice) can work without --since
	mustRun(t, envAlice, "work", "1", "--status", "in_progress")
}

// TestFirstUserBecomesGod verifies that the first user gets the god role and subsequent users get worker.
func TestFirstUserBecomesGod(t *testing.T) {
	envAlice, dbPath, _ := testEnv(t, "alice")
	mustRun(t, envAlice, "add", "--title", "seed")

	envBob := append([]string(nil), envAlice...)
	for i, e := range envBob {
		if strings.HasPrefix(e, "PICOLY_USER_ID=") {
			envBob[i] = "PICOLY_USER_ID=bob"
		}
	}
	mustRun(t, envBob, "list") // register bob

	out, err := exec.Command("sqlite3", dbPath, "SELECT id, role FROM users ORDER BY id;").Output()
	if err != nil {
		t.Fatalf("sqlite3: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 users, got: %s", out)
	}
	if lines[0] != "alice|god" {
		t.Errorf("alice role: got %q, want %q", lines[0], "alice|god")
	}
	if lines[1] != "bob|worker" {
		t.Errorf("bob role: got %q, want %q", lines[1], "bob|worker")
	}
}

// TestMarkdownExport verifies that board.md and tickets/<id>.md are generated after add.
func TestMarkdownExport(t *testing.T) {
	env, _, exportDir := testEnv(t, "alice")

	mustRun(t, env, "add", "--title", "Export test", "--desc", "Check files")

	boardPath := filepath.Join(exportDir, "board.md")
	if _, err := os.Stat(boardPath); err != nil {
		t.Errorf("board.md not found: %v", err)
	}
	ticketPath := filepath.Join(exportDir, "tickets", "1.md")
	if _, err := os.Stat(ticketPath); err != nil {
		t.Errorf("tickets/1.md not found: %v", err)
	}

	content, _ := os.ReadFile(boardPath)
	if !strings.Contains(string(content), "Export test") {
		t.Error("board.md does not contain ticket title")
	}
}
