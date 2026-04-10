package codex

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/ekoru/dev-flow-tui/internal/process"
)

// Launch starts Codex CLI in the given worktree directory.
// On Windows: opens a new terminal with WSL, cd into the path, then runs codex.
func Launch(pm *process.Manager, wtPath string) error {
	if pm.IsRunning(process.KindCodex, wtPath) {
		return fmt.Errorf("codex is already running for %s", wtPath)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		wslPath := winToWSLPath(wtPath)
		// Open new Windows Terminal tab (or cmd window) → wsl → cd → codex
		cmd = exec.Command("cmd", "/c", "start",
			fmt.Sprintf("Codex - %s", wtPath),
			"wsl", "--", "bash", "-ic",
			fmt.Sprintf("cd '%s' && codex", wslPath))
	default:
		cmd = exec.Command("sh", "-c",
			fmt.Sprintf("cd %q && codex", wtPath))
	}

	return pm.Start(process.KindCodex, "codex", wtPath, cmd)
}

// winToWSLPath converts a Windows path like "F:\foo\bar" or "F:/foo/bar"
// to a WSL path like "/mnt/f/foo/bar".
func winToWSLPath(p string) string {
	// Normalize backslashes
	p = strings.ReplaceAll(p, "\\", "/")

	// Handle "X:/..." → "/mnt/x/..."
	if len(p) >= 2 && p[1] == ':' {
		drive := strings.ToLower(string(p[0]))
		rest := p[2:]
		return "/mnt/" + drive + rest
	}
	return p
}
