# Picoly

A lightweight Kanban board system for AI agents. Picoly lets coding agents (Claude Code, etc.) share and manage task lists through a simple CLI, while keeping human-readable Markdown files in sync.

## Requirements

- Go 1.25+
- Linux (the file lock implementation uses `syscall.Flock`)

## Setup

### 1. Build

Use the bundled `just` binary:

```sh
bin/just build
```

This produces `bin/picoly`. Alternatively, build directly with Go:

```sh
GOEXPERIMENT=jsonv2 go build -o bin/picoly ./cmd/picoly
```

### 2. Environment variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `PICOLY_USER_ID` | **Yes** | — | The user ID of the current agent or user |
| `PICOLY_DB` | No | `./picoly.db` | Path to the SQLite database file |
| `PICOLY_EXPORT` | No | `./picoly` | Directory where Markdown files are written |

The first user to run any `picoly` command against a new database is automatically assigned the **god** role. All subsequent new users receive the **worker** role.

### 3. Roles

| Role | Add / Edit tickets | Update status / Comment |
|---|---|---|
| `god` | ✓ | ✓ |
| `admin` | ✓ | ✓ |
| `worker` | ✗ | ✓ |

## Testing

### Unit tests

Standard Go unit tests covering individual packages (model, db, store, auth, export, lock):

```sh
bin/just test
```

### Integration tests

End-to-end tests that exercise the `picoly` binary as a black box. Because they require a pre-built binary, they live in `cmd/integration-test/` as a standalone Go program rather than in a `_test.go` file.

```sh
bin/just e2e        # builds bin/picoly first, then runs all integration tests
```

To run the integration tests against an already-built binary:

```sh
GOEXPERIMENT=jsonv2 go run ./cmd/integration-test [path/to/picoly]
```

Each test case gets its own isolated temporary directory (DB + export dir) so tests do not interfere with each other. Database assertions read directly from SQLite using the internal `db` package instead of shelling out to an external tool.

## Commands

All commands print a single JSON object to stdout. On error, they exit with code 1 and print `{"error": "..."}` to stdout.

---

### `picoly add` — Add a ticket

Requires admin role or above.

```
picoly add --title <title> [--desc <description>]
picoly add --json '{"title":"...","description":"..."}'
```

| Flag | Description |
|---|---|
| `--title` | Ticket title (required unless `--json` is used) |
| `--desc` | Ticket description |
| `--json` | JSON object with `title` and/or `description` fields; overrides the other flags |

**Example:**

```sh
export PICOLY_USER_ID=alice

picoly add --title "Implement login page" --desc "OAuth2 + session cookie"
```

```json
{
  "id": 1,
  "title": "Implement login page",
  "description": "OAuth2 + session cookie",
  "status": "todo",
  "created_by": "alice",
  "created_at": "2026-03-08T09:00:00Z",
  "updated_at": "2026-03-08T09:00:00Z"
}
```

---

### `picoly edit` — Edit a ticket

Requires admin role or above.

```
picoly edit <id> [--title <title>] [--desc <description>] [--since <RFC3339>]
picoly edit <id> --json '{"title":"...","description":"..."}' [--since <RFC3339>]
```

| Flag | Description |
|---|---|
| `--title` | New title |
| `--desc` | New description |
| `--json` | JSON object with `title` and/or `description` fields |
| `--since` | Expected `updated_at` value (RFC3339). If the ticket has been updated since this timestamp, the command fails with a race condition error |

**Example:**

```sh
picoly edit 1 --title "Implement login page (v2)"
```

---

### `picoly work` — Update status or add a comment

Available to all roles. At least one of `--status` or `--comment` must be provided.

```
picoly work <id> [--status <status>] [--comment <text>] [--since <RFC3339>]
```

| Flag | Description |
|---|---|
| `--status` | New status: `todo`, `in_progress`, `done`, or `cancelled` |
| `--comment` | Comment text to append |
| `--since` | Expected `updated_at` value (RFC3339) for optimistic locking on status changes |

**Example:**

```sh
# Start working on a ticket
picoly work 1 --status in_progress

# Add a progress note
picoly work 1 --comment "Finished the OAuth flow, moving on to session handling"

# Do both at once with optimistic locking
TS=$(picoly read 1 | python3 -c "import sys,json; print(json.load(sys.stdin)['updated_at'])")
picoly work 1 --status done --comment "All done" --since "$TS"
```

---

### `picoly read` — Read a ticket

Available to all roles.

```
picoly read <id>
```

Returns the full ticket including all comments.

**Example:**

```sh
picoly read 1
```

```json
{
  "id": 1,
  "title": "Implement login page",
  "description": "OAuth2 + session cookie",
  "status": "in_progress",
  "created_by": "alice",
  "created_at": "2026-03-08T09:00:00Z",
  "updated_at": "2026-03-08T09:05:00Z",
  "comments": [
    {
      "id": 1,
      "ticket_id": 1,
      "body": "Started the OAuth flow",
      "author": "alice",
      "created_at": "2026-03-08T09:05:00Z"
    }
  ]
}
```

---

### `picoly list` — List tickets

Available to all roles.

```
picoly list [--status <status>] [--sort id|updated_at|status] [--asc]
```

| Flag | Description |
|---|---|
| `--status` | Filter by status: `todo`, `in_progress`, `done`, or `cancelled` |
| `--sort` | Sort field: `id` (default), `updated_at`, or `status` |
| `--asc` | Sort ascending (default is descending when `--sort` is specified) |

**Example:**

```sh
# All open tickets, sorted by last update
picoly list --status in_progress --sort updated_at
```

```json
{
  "tickets": [
    {
      "id": 1,
      "title": "Implement login page",
      "status": "in_progress",
      "comment_count": 1,
      "updated_at": "2026-03-08T09:05:00Z"
    }
  ],
  "total": 1
}
```

---

## Markdown output

After every write operation (`add`, `edit`, `work`), Picoly regenerates two Markdown files inside `PICOLY_EXPORT`:

- `board.md` — overview table of all tickets
- `tickets/<id>.md` — full detail of the changed ticket

These files are intended for humans to review progress at a glance.

## Concurrency

Write commands (`add`, `edit`, `work`) acquire an exclusive file lock on `<PICOLY_DB>.lock` before touching the database, so concurrent agents serialize automatically.

For fine-grained race detection, pass `--since <updated_at>` with the timestamp you last read. The command will fail if another agent has modified the ticket in the meantime, letting you retry with fresh data.
