# wre(**CCL**)ess

Fire-and-forget [Claude Code](https://docs.anthropic.com/en/docs/claude-code) workers. Launch background sessions, track them, get out of the way.

![Dashboard](docs/gifs/dashboard.gif)

Single binary. No daemon, no database. State lives as flat JSON files in `~/.local/state/ccl/`. Scripts, pickers, and status bars can talk to it. Ships with a TUI if you want one.

## Install

```sh
go install github.com/scottstav/wreccless/cmd/ccl@latest
```

## Quick start

```sh
# fire off a worker
ccl new --dir ~/projects/myapp --task "Add rate limiting to the API"

# or create one that waits for approval first
ccl new --dir ~/projects/myapp --task "Refactor auth" --pending

# see what's going on
ccl list

# open the TUI
ccl ui
```

## TUI

`ccl ui` gives you a dashboard with all your workers, color-coded statuses, and a live log preview. Navigate with `j`/`k`, hit `Enter` to open full logs, `n` to create a new worker. Press `?` for all keybindings.

### Create, finish, resume

![Create and resume](docs/gifs/create-and-resume.gif)

Workers fire hooks on state changes — the `on_done` hook above sends a desktop notification. Hit `r` to resume any worker interactively (drops you into `claude --resume`).

## CLI

```sh
ccl new --dir <path> --task "..."   # launch a worker
ccl new ... --pending               # launch pending (needs approval)
ccl list                            # list workers (--json, --status <s>)
ccl status <id>                     # detailed info (--json)
ccl approve <id>                    # start a pending worker
ccl deny <id>                       # reject a pending worker
ccl kill <id>                       # stop a running worker
ccl resume <id>                     # drop into claude --resume
ccl logs <id>                       # rendered output (-f to follow, --json for raw)
ccl clean                           # remove done/error workers (--all for everything)
ccl ui                              # TUI dashboard
```

`ccl list --json` and `ccl status --json` make it trivial to wire into waybar, polybar, or whatever status bar you're running.

## Config

`~/.config/ccl/config.toml` — ships with sane defaults, everything's optional.

```toml
[claude]
skip_permissions = true   # hence the repo name
system_prompt = "Complete the task. Don't ask questions."
extra_flags = []

[hooks]
on_start   = ["pkill -SIGRTMIN+12 waybar"]
on_done    = ["notify-send 'Done' '{{.Task}}'"]
on_pending = ["notify-send 'Pending' '{{.Task}}'"]
on_error   = ["notify-send -u critical 'Failed' '{{.Task}}'"]
on_kill    = []
```

### Claude args

`skip_permissions` runs claude with `--dangerously-skip-permissions` — fully autonomous, no confirmation prompts. `system_prompt` overrides the default. `extra_flags` passes anything else through to the claude CLI.

### Hooks

Shell commands that fire on worker state transitions. Templates have access to `{{.ID}}`, `{{.Task}}`, `{{.Dir}}`, `{{.Status}}`, and `{{.SessionID}}`. Good for notifications, status bar refreshes, or chaining workflows.

## How it works

`ccl new` writes a state file and spawns a detached `ccl run <id>` process that calls `claude -p --output-format stream-json`. Output streams to a log file. When claude exits, state transitions to `done` or `error` and hooks fire. That's it.
