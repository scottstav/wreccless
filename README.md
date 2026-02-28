# wre(**CCL**)ess Claude Code Launcher

Fire-and-forget Claude Code workers from the command line.

`ccl` launches background [Claude Code](https://docs.anthropic.com/en/docs/claude-code) sessions, tracks their lifecycle, and gets out of the way. No TUI required — just a single binary that scripts, pickers, and status bars can talk to.

## Install

```sh
go install github.com/scottstav/wreccless/cmd/ccl@latest
```

## Usage

```sh
# Launch a worker (starts immediately in background)
ccl new --dir ~/projects/myapp --task "Add rate limiting to the API"

# Launch as pending (requires manual approval)
ccl new --dir ~/projects/myapp --task "Refactor auth" --pending

# See what's running
ccl list
ccl list --json              # machine-readable
ccl list --status working    # filter by status

# Detailed info
ccl status <id>

# Manage pending workers
ccl approve <id>
ccl deny <id>

# Interrupt a running worker
ccl kill <id>

# Resume interactively (execs into claude --resume)
ccl resume <id>

# Tail worker output
ccl logs <id>
ccl logs <id> -f             # follow
ccl logs <id> --json         # raw NDJSON stream

# Clean up finished workers
ccl clean                    # removes done/error
ccl clean --all              # removes everything
```

## Configuration

`~/.config/ccl/config.toml`

```toml
[claude]
skip_permissions = true
system_prompt = "Complete the task. Don't ask questions."
extra_flags = []

[hooks]
on_start   = ["pkill -SIGRTMIN+12 waybar"]
on_done    = ["notify-send 'Done' '{{.Task}}'"]
on_pending = ["notify-send 'Pending' '{{.Task}}'"]
on_error   = ["notify-send -u critical 'Failed' '{{.Task}}'"]
on_kill    = []
```

Hook templates have access to `{{.ID}}`, `{{.Task}}`, `{{.Dir}}`, `{{.Status}}`, and `{{.SessionID}}`.

## How it works

`ccl new` writes a state file, then spawns a detached `ccl run <id>` process that calls `claude -p --output-format stream-json --verbose`. Output goes to a log file. When claude exits, the state transitions to `done` (or `error`) and hooks fire.

Workers live in `~/.local/state/ccl/` as JSON files — one per worker. No daemon, no database.
