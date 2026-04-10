package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ekoru/dev-flow-tui/internal/config"
	"github.com/ekoru/dev-flow-tui/internal/git"
	"github.com/ekoru/dev-flow-tui/internal/process"
	"github.com/ekoru/dev-flow-tui/internal/ui"
)

func main() {
	repoRoot, err := git.FindRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\nRun this command from inside a git repository.\n", err)
		os.Exit(1)
	}

	// Default worktree base: sibling directory to repo root named "<repo>-worktrees"
	repoName := filepath.Base(repoRoot)
	worktreeBase := filepath.Join(filepath.Dir(repoRoot), repoName+"-worktrees")

	// Load per-repo config
	cfg := config.Load(repoRoot)

	// Config can override worktree base
	if cfg.WorktreeBase != "" {
		worktreeBase = cfg.WorktreeBase
	}
	// Env var has highest priority
	if base := os.Getenv("DEVFLOW_WORKTREE_BASE"); base != "" {
		worktreeBase = base
	}

	// Ensure worktree base directory exists
	if err := os.MkdirAll(worktreeBase, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating worktree base dir: %v\n", err)
		os.Exit(1)
	}

	pm := process.New()
	model := ui.NewModel(repoRoot, worktreeBase, pm, cfg)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		pm.StopAll()
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
