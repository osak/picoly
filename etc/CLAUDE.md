# Picoly task management

If `picoly.db` exists in the current working directory, manage tasks using the `picoly` CLI.

## When picoly.db is present

All agents (orchestrator and subagents alike) must follow this protocol.

### Orchestrator: delegating work to a subagent

1. Create a ticket before spawning the subagent:
   ```bash
   picoly add --title "Task summary" --desc "Detailed instructions"
   ```
2. Pass the ticket ID to the subagent in its prompt, along with the agent name to use as `PICOLY_USER_ID`.
3. After the subagent reports back, read the ticket to determine the next action:
   ```bash
   picoly read <id>
   ```
   - Status `done` → proceed to the next task
   - Status `cancelled` → read the comments and decide on recovery

### Subagent: when given a ticket ID

1. Read the ticket to get the full task description:
   ```bash
   picoly read <id>
   ```
2. Mark work as started (use the `updated_at` from the read output as `--since`):
   ```bash
   picoly work <id> --status in_progress --since "<updated_at>"
   ```
3. Report progress periodically via comments:
   ```bash
   picoly work <id> --comment "Done X, moving on to Y" --since "<updated_at>"
   ```
4. On completion or failure, record the final result and report back to the orchestrator:
   ```bash
   # success
   picoly work <id> --status done --comment "Completed. <summary>" --since "<updated_at>"
   # failure
   picoly work <id> --status cancelled --comment "Problem: <detail>. Attempted: <what was tried>" --since "<updated_at>"
   ```

### Passing --since correctly

Always take `updated_at` from the most recent `picoly read` or `picoly work` output and pass it to the next write command:

```bash
OUT=$(picoly read <id>)
TS=$(echo "$OUT" | python3 -c "import sys,json; print(json.load(sys.stdin)['updated_at'])")
picoly work <id> --status in_progress --since "$TS"

OUT=$(picoly work <id> --status in_progress --since "$TS")
TS=$(echo "$OUT" | python3 -c "import sys,json; print(json.load(sys.stdin)['updated_at'])")
# use the new TS for subsequent commands
```

### Environment variables

| Variable | Required | Description |
|---|---|---|
| `PICOLY_USER_ID` | **Yes** | Agent name to record as the author of operations |
| `PICOLY_DB` | No | Path to the DB file (default: `./picoly.db`) |
| `PICOLY_EXPORT` | No | Markdown output directory (default: `./picoly/`) |

When spawning a subagent, instruct it to set `PICOLY_USER_ID` to that agent's name.
