package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/ekoru/dev-flow-tui/internal/docker"
	"github.com/ekoru/dev-flow-tui/internal/git"
)

func (m Model) View() string {
	if m.loadErr != nil {
		return errorStyle.Render(fmt.Sprintf("\n  Error: %v\n\n  Make sure you're in a git repository.\n", m.loadErr))
	}

	switch m.mode {
	case modeCreate:
		return m.viewCreate()
	case modeConfirmDelete:
		return m.viewConfirmDelete()
	case modeActions:
		return m.viewActions()
	default:
		return m.viewList()
	}
}

func (m Model) viewList() string {
	w := m.width
	if w < 40 {
		w = 80
	}
	h := m.height
	if h < 10 {
		h = 24
	}

	leftW := w*2/5 - 1
	bodyH := h - 4

	// Header
	header := titleStyle.Render("dev-flow") + titleDimStyle.Render(" worktree manager")

	// Left column: list
	var leftLines []string
	for i, wt := range m.worktrees {
		leftLines = append(leftLines, m.renderRow(i, wt))
	}
	if len(leftLines) == 0 {
		leftLines = append(leftLines, helpDescStyle.Render("  no worktrees"))
	}
	for len(leftLines) < bodyH {
		leftLines = append(leftLines, "")
	}

	// Right column: details
	rightLines := m.renderDetail(m.selected())
	for len(rightLines) < bodyH {
		rightLines = append(rightLines, "")
	}

	// Compose columns
	sep := lipgloss.NewStyle().Foreground(colorBorder).Render("│")
	var bodyRows []string
	for i := 0; i < bodyH; i++ {
		left := padRight(leftLines[i], leftW)
		right := ""
		if i < len(rightLines) {
			right = rightLines[i]
		}
		bodyRows = append(bodyRows, left+" "+sep+" "+right)
	}
	body := strings.Join(bodyRows, "\n")

	// Status
	status := ""
	if m.statusMsg != "" {
		if m.statusErr {
			status = statusBarStyle.Render(errorStyle.Render(m.statusMsg))
		} else {
			status = statusBarStyle.Render(successStyle.Render(m.statusMsg))
		}
	}

	// Help
	help := m.renderHelp()

	return header + "\n\n" + body + "\n" + status + "\n" + help
}

// ─── Row ─────────────────────────────────────────────────

func (m Model) renderRow(idx int, wt worktreeView) string {
	name := wt.Branch
	if name == "" {
		name = filepath.Base(wt.Path)
	}

	selected := idx == m.cursor
	prefix := "  "
	if selected {
		prefix = "> "
	}

	var namePart string
	if selected {
		if wt.IsMain {
			namePart = rowMainSelectedStyle.Render(prefix + name)
		} else {
			namePart = rowSelectedStyle.Render(prefix + name)
		}
	} else {
		if wt.IsMain {
			namePart = rowMainNormalStyle.Render(prefix + name)
		} else {
			namePart = rowNormalStyle.Render(prefix + name)
		}
	}

	// Sync indicator
	if sync, ok := m.syncCache[wt.Path]; ok && !sync.NoRemote {
		short := sync.Short()
		if short == "✓" {
			namePart += " " + statusCleanStyle.Render(short)
		} else if short != "" {
			namePart += " " + statusDirtyStyle.Render(short)
		}
	}

	return namePart
}

// worktreeView is a type alias to avoid import in signature
type worktreeView = git.Worktree

// ─── Detail ──────────────────────────────────────────────

func (m Model) renderDetail(wt *git.Worktree) []string {
	if wt == nil {
		return []string{helpDescStyle.Render(" select a worktree")}
	}

	branch := wt.Branch
	if branch == "" {
		branch = "(detached)"
	}
	head := wt.HEAD
	if len(head) > 8 {
		head = head[:8]
	}
	displayPath := wt.Path
	if rel, err := filepath.Rel(filepath.Dir(m.repoRoot), wt.Path); err == nil {
		displayPath = filepath.ToSlash(rel)
	}

	var lines []string

	lines = append(lines, row("branch", branch))
	lines = append(lines, row("commit", head))
	lines = append(lines, row("path", displayPath))
	lines = append(lines, "")

	// Cached git status
	if s, ok := m.statusCache[wt.Path]; ok {
		if s == "clean" {
			lines = append(lines, row("git", statusCleanStyle.Render("clean")))
		} else if s == "error" {
			lines = append(lines, row("git", errorStyle.Render("error")))
		} else {
			lines = append(lines, row("git", statusDirtyStyle.Render(s)))
		}
	}

	// Sync status
	if sync, ok := m.syncCache[wt.Path]; ok {
		if sync.NoRemote {
			lines = append(lines, row("sync", helpDescStyle.Render("no remote tracking")))
		} else {
			summary := sync.Summary()
			if sync.Ahead == 0 && sync.Behind == 0 {
				lines = append(lines, row("sync", statusCleanStyle.Render(summary)))
			} else {
				lines = append(lines, row("sync", statusDirtyStyle.Render(summary)))
			}
			lines = append(lines, row("remote", helpDescStyle.Render(sync.Upstream)))
		}
	}

	if docker.HasComposeFile(wt.Path) {
		lines = append(lines, row("docker", detailValueStyle.Render("compose found")))
	}

	return lines
}

func row(label, value string) string {
	return " " + detailLabelStyle.Render(label) + detailValueStyle.Render(value)
}

func padRight(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

// ─── Help ────────────────────────────────────────────────

func (m Model) renderHelp() string {
	entries := [][2]string{
		{"↑↓", "navigate"},
		{"enter", "actions"},
		{"c", "create"},
		{"d", "delete"},
		{"g", "pull"},
		{"f", "fetch"},
		{"r", "refresh"},
		{"q", "quit"},
	}

	var parts []string
	for _, e := range entries {
		parts = append(parts, helpKeyStyle.Render(e[0])+" "+helpDescStyle.Render(e[1]))
	}

	sep := helpSepStyle.Render(" │ ")
	return "  " + strings.Join(parts, sep)
}

// ─── Actions Menu ────────────────────────────────────────

func (m Model) viewActions() string {
	wt := m.selected()
	if wt == nil {
		return ""
	}

	var b strings.Builder

	name := wt.Branch
	if name == "" {
		name = filepath.Base(wt.Path)
	}
	b.WriteString(titleStyle.Render("actions") + titleDimStyle.Render(" "+name))
	b.WriteString("\n\n")

	for i, a := range m.actions {
		prefix := "  "
		if i == m.actionCursor {
			prefix = "> "
		}

		keyPart := helpKeyStyle.Render(fmt.Sprintf("%-2s", a.key))
		if i == m.actionCursor {
			b.WriteString(keyPart + " " + rowSelectedStyle.Render(prefix+a.label))
		} else {
			b.WriteString(keyPart + " " + rowNormalStyle.Render(prefix+a.label))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString("  " + helpDescStyle.Render("enter select  /  esc back  /  shortcut key"))

	return b.String()
}

// ─── Dialogs ─────────────────────────────────────────────

func (m Model) viewCreate() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("create worktree"))
	b.WriteString("\n\n")

	modeLabel := "new branch"
	if !m.newBranch {
		modeLabel = "existing branch"
	}

	var lines []string
	lines = append(lines, inputLabelStyle.Render("mode")+"  "+detailValueStyle.Render(modeLabel))
	lines = append(lines, "")
	lines = append(lines, inputLabelStyle.Render("branch name"))
	lines = append(lines, m.branchInput.View())
	lines = append(lines, "")
	lines = append(lines, inputLabelStyle.Render("from")+"  "+helpDescStyle.Render("default: "+m.cfg.BaseBranch))
	lines = append(lines, m.baseBranchInput.View())

	// Suggestions list — fixed height so dialog doesn't jump
	const maxShow = 12
	lines = append(lines, "")

	if m.createFocus == 1 && m.showSuggestions && len(m.filteredBranches) > 0 {
		total := len(m.filteredBranches)

		// Scroll window around cursor
		visible := maxShow
		if total < visible {
			visible = total
		}
		start := 0
		if m.suggestCursor >= visible {
			start = m.suggestCursor - visible + 1
		}
		end := start + visible
		if end > total {
			end = total
			start = end - visible
			if start < 0 {
				start = 0
			}
		}

		// Counter
		lines = append(lines, helpDescStyle.Render(fmt.Sprintf("  %d branches", total)))

		rendered := 0
		for i := start; i < end; i++ {
			bi := m.filteredBranches[i]
			var tag string
			switch bi.Origin {
			case git.BranchLocal:
				tag = branchTagLocal.Render("L")
			case git.BranchRemote:
				tag = branchTagRemote.Render("R")
			case git.BranchBoth:
				tag = branchTagBoth.Render("LR")
			}
			if i == m.suggestCursor {
				lines = append(lines, tag+" "+rowSelectedStyle.Render(bi.Name))
			} else {
				lines = append(lines, tag+" "+helpDescStyle.Render(bi.Name))
			}
			rendered++
		}

		// Pad remaining rows to keep height fixed
		for rendered < maxShow {
			lines = append(lines, "")
			rendered++
		}
	} else {
		// Empty state — still reserve space
		for i := 0; i < maxShow+1; i++ {
			lines = append(lines, "")
		}
	}

	lines = append(lines, "")
	lines = append(lines, helpDescStyle.Render("tab next  /  ↑↓ select  /  enter confirm  /  esc cancel"))

	b.WriteString(dialogStyle.Render(strings.Join(lines, "\n")))

	if m.statusMsg != "" {
		b.WriteString("\n\n")
		if m.statusErr {
			b.WriteString(statusBarStyle.Render(errorStyle.Render(m.statusMsg)))
		} else {
			b.WriteString(statusBarStyle.Render(successStyle.Render(m.statusMsg)))
		}
	}

	return b.String()
}

func (m Model) viewConfirmDelete() string {
	wt := m.selected()
	if wt == nil {
		return ""
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("delete worktree"))
	b.WriteString("\n\n")

	warning := "worktree and directory will be removed"
	if wt.Branch != "" && wt.Branch != "main" && wt.Branch != "master" {
		warning = fmt.Sprintf("worktree, directory and branch '%s' will be deleted", wt.Branch)
	}

	// Warn if dirty
	dirty := ""
	if s, ok := m.statusCache[wt.Path]; ok && s != "clean" && s != "error" {
		dirty = "\n" + warningStyle.Render("this worktree has uncommitted changes!")
	}

	content := fmt.Sprintf(
		"%s%s\n\n%s\n%s\n\n%s",
		errorStyle.Render(warning),
		dirty,
		row("branch", wt.Branch),
		row("path", wt.Path),
		helpKeyStyle.Render("y")+" confirm  /  "+helpKeyStyle.Render("n")+" cancel",
	)

	b.WriteString(dialogStyle.Render(content))
	return b.String()
}
