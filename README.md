# dev-flow-tui

Terminal UI for managing Git Worktrees and dev workflows. Built with Go + Bubble Tea.

## Features

- **Worktree management**: list, create, delete git worktrees
- **IDE integration**: open worktree in JetBrains Rider or VS Code
- **Codex CLI**: launch Codex in any worktree (opens new terminal window)
- **Docker Compose**: start/stop docker compose from the TUI
- **Process tracking**: see which worktrees have running processes

## Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- [Git](https://git-scm.com/)
- Optional: JetBrains Rider, VS Code, Docker, Codex CLI

## Build & Run

```bash
# Clone and build
cd dev-flow-tui
go mod tidy
go build -o dev-flow.exe ./cmd/dev-flow

# Run from inside any git repository
cd /path/to/your/repo
/path/to/dev-flow.exe
```

Or install directly:

```bash
go install github.com/ekoru/dev-flow-tui/cmd/dev-flow@latest
```

## Configuration

| Env Variable | Description | Default |
|---|---|---|
| `DEVFLOW_WORKTREE_BASE` | Directory where new worktrees are created | `../<repo-name>-worktrees` |

## Keyboard Shortcuts

| Key | Action |
|---|---|
| `↑`/`k` | Move up |
| `↓`/`j` | Move down |
| `c` | Create new worktree |
| `d` | Delete selected worktree |
| `o` | Open in Rider |
| `v` | Open in VS Code |
| `x` | Launch Codex CLI |
| `u` | Docker Compose up |
| `s` | Docker Compose down |
| `r` | Refresh worktree list |
| `q` | Quit |

### Create Worktree Dialog

| Key | Action |
|---|---|
| `Tab` | Toggle new branch / existing branch |
| `Enter` | Confirm |
| `Esc` | Cancel |

### Delete Confirmation

| Key | Action |
|---|---|
| `Y` | Confirm delete |
| `N`/`Esc` | Cancel |

## Project Structure

```
dev-flow-tui/
├── cmd/dev-flow/main.go       # Entry point
├── internal/
│   ├── git/worktree.go        # Git worktree operations
│   ├── process/manager.go     # Child process tracking
│   ├── ide/launcher.go        # Rider / VS Code launcher
│   ├── codex/codex.go         # Codex CLI launcher
│   ├── docker/compose.go      # Docker Compose operations
│   └── ui/
│       ├── model.go           # Bubble Tea model + update logic
│       ├── view.go            # Rendering
│       ├── keys.go            # Key bindings
│       └── styles.go          # Lip Gloss styles
├── go.mod
└── README.md
```

## How It Works

1. Run `dev-flow` from inside any git repository
2. It detects the repo root and lists all worktrees
3. New worktrees are created in a sibling directory (`<repo>-worktrees/`)
4. External processes (Codex, Docker) are tracked and shown in the UI
5. IDE launches are fire-and-forget (detached processes)

## Future Plans

- Agent scenario templates
- Quick command runner per worktree
- Git status details (ahead/behind, stash count)
- Health checks for running services
- Multi-agent workflow orchestration
