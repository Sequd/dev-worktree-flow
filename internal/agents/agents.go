package agents

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/ekoru/dev-flow-tui/internal/process"
)

// Agent defines a launchable coding agent.
type Agent struct {
	Name    string
	Command string       // command to run
	Kind    process.Kind // process tracker key
	WSL     bool         // run through WSL (for tools that need Linux)
}

var (
	Codex = Agent{
		Name:    "Codex",
		Command: "codex",
		Kind:    process.KindCodex,
		WSL:     true,
	}
	ClaudeCode = Agent{
		Name:    "Claude Code",
		Command: "claude",
		Kind:    "claude",
		WSL:     false,
	}
	QwenAgent = Agent{
		Name:    "Qwen",
		Command: "qwen",
		Kind:    "qwen",
		WSL:     false,
	}

	// All available agents
	All = []Agent{Codex, ClaudeCode, QwenAgent}
)

// Launch starts an agent in a new terminal window in the given worktree directory.
func Launch(pm *process.Manager, agent Agent, wtPath string) error {
	if pm.IsRunning(agent.Kind, wtPath) {
		return fmt.Errorf("%s is already running for %s", agent.Name, wtPath)
	}

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		if agent.WSL {
			wslPath := winToWSLPath(wtPath)
			cmd = exec.Command("cmd", "/c", "start",
				fmt.Sprintf("%s - %s", agent.Name, wtPath),
				"wsl", "--", "bash", "-ic",
				fmt.Sprintf("cd '%s' && %s", wslPath, agent.Command))
		} else {
			// Native Windows: open new terminal, cd, run agent
			cmd = exec.Command("cmd", "/c", "start",
				fmt.Sprintf("%s - %s", agent.Name, wtPath),
				"cmd", "/k",
				fmt.Sprintf("cd /d \"%s\" && %s", wtPath, agent.Command))
		}
	default:
		cmd = exec.Command("sh", "-c",
			fmt.Sprintf("cd %q && %s", wtPath, agent.Command))
	}

	return pm.Start(agent.Kind, agent.Name, wtPath, cmd)
}

func winToWSLPath(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	if len(p) >= 2 && p[1] == ':' {
		drive := strings.ToLower(string(p[0]))
		rest := p[2:]
		return "/mnt/" + drive + rest
	}
	return p
}
