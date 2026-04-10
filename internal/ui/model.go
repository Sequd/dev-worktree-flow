package ui

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ekoru/dev-flow-tui/internal/agents"
	"github.com/ekoru/dev-flow-tui/internal/config"
	"github.com/ekoru/dev-flow-tui/internal/docker"
	"github.com/ekoru/dev-flow-tui/internal/git"
	"github.com/ekoru/dev-flow-tui/internal/ide"
	"github.com/ekoru/dev-flow-tui/internal/process"
)

type mode int

const (
	modeList mode = iota
	modeCreate
	modeConfirmDelete
	modeActions // quick actions menu
)

// Model is the main Bubble Tea model.
type Model struct {
	worktrees    []git.Worktree
	cursor       int
	mode         mode
	width        int
	height       int
	repoRoot     string
	worktreeBase string

	pm  *process.Manager
	cfg config.Config

	// Create form
	branchInput     textinput.Model
	baseBranchInput textinput.Model
	newBranch       bool
	createFocus     int // 0 = branch name, 1 = base branch

	// Branch suggestions
	allBranches      []git.BranchInfo
	filteredBranches []git.BranchInfo
	suggestCursor    int
	showSuggestions  bool

	// Status messages with auto-hide
	statusMsg   string
	statusErr   bool
	statusTimer int // countdown ticks

	// Cached git status per worktree path
	statusCache map[string]string
	syncCache   map[string]git.SyncStatus

	// Actions menu
	actionCursor int
	actions      []action

	loadErr error
}

type action struct {
	key   string
	label string
	fn    func(m *Model) tea.Cmd
}

// Messages
type worktreeLoadedMsg struct {
	worktrees []git.Worktree
	err       error
}

type actionResultMsg struct {
	msg string
	err bool
}

type clearStatusMsg struct{}

type statusCacheMsg struct {
	cache     map[string]string
	syncCache map[string]git.SyncStatus
}

func NewModel(repoRoot, worktreeBase string, pm *process.Manager, cfg config.Config) Model {
	ti := textinput.New()
	ti.Placeholder = "feature/my-task"
	ti.CharLimit = 128
	ti.Width = 40

	bi := textinput.New()
	bi.Placeholder = cfg.BaseBranch
	bi.CharLimit = 128
	bi.Width = 40

	return Model{
		repoRoot:        repoRoot,
		worktreeBase:    worktreeBase,
		pm:              pm,
		cfg:             cfg,
		branchInput:     ti,
		baseBranchInput: bi,
		newBranch:       true,
		statusCache:     make(map[string]string),
		syncCache:       make(map[string]git.SyncStatus),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(loadWorktrees, refreshStatusCache, tickAutoRefresh())
}

// Auto-refresh every 30 seconds
type autoRefreshMsg struct{}

func tickAutoRefresh() tea.Cmd {
	return tea.Tick(30*time.Second, func(time.Time) tea.Msg {
		return autoRefreshMsg{}
	})
}

func loadWorktrees() tea.Msg {
	wts, err := git.List()
	return worktreeLoadedMsg{worktrees: wts, err: err}
}

func refreshStatusCache() tea.Msg {
	wts, err := git.List()
	if err != nil {
		return statusCacheMsg{cache: nil}
	}
	cache := make(map[string]string)
	sync := make(map[string]git.SyncStatus)
	for _, wt := range wts {
		s, err := git.ShortStatus(wt.Path)
		if err != nil {
			cache[wt.Path] = "error"
		} else {
			cache[wt.Path] = s
		}
		sync[wt.Path] = git.GetSyncStatus(wt.Path)
	}
	return statusCacheMsg{cache: cache, syncCache: sync}
}

func clearStatusAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

func (m *Model) setStatus(msg string, isErr bool) tea.Cmd {
	m.statusMsg = msg
	m.statusErr = isErr
	return clearStatusAfter(3 * time.Second)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case worktreeLoadedMsg:
		if msg.err != nil {
			m.loadErr = msg.err
			return m, nil
		}
		m.worktrees = msg.worktrees
		m.loadErr = nil
		if m.cursor >= len(m.worktrees) {
			m.cursor = max(0, len(m.worktrees)-1)
		}
		return m, nil

	case statusCacheMsg:
		if msg.cache != nil {
			m.statusCache = msg.cache
		}
		if msg.syncCache != nil {
			m.syncCache = msg.syncCache
		}
		return m, nil

	case actionResultMsg:
		cmd := m.setStatus(msg.msg, msg.err)
		return m, tea.Batch(cmd, refreshStatusCache)

	case clearStatusMsg:
		m.statusMsg = ""
		return m, nil

	case autoRefreshMsg:
		return m, tea.Batch(loadWorktrees, refreshStatusCache, tickAutoRefresh())

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	if m.mode == modeCreate {
		var cmd tea.Cmd
		m.branchInput, cmd = m.branchInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, keys.Quit) && m.mode == modeList {
		m.pm.StopAll()
		return m, tea.Quit
	}
	if key.Matches(msg, keys.Cancel) {
		m.mode = modeList
		m.statusMsg = ""
		return m, nil
	}

	switch m.mode {
	case modeList:
		return m.handleListKey(msg)
	case modeCreate:
		return m.handleCreateKey(msg)
	case modeConfirmDelete:
		return m.handleDeleteKey(msg)
	case modeActions:
		return m.handleActionsKey(msg)
	}
	return m, nil
}

func (m Model) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(msg, keys.Down):
		if m.cursor < len(m.worktrees)-1 {
			m.cursor++
		}
	case key.Matches(msg, keys.Actions):
		if wt := m.selected(); wt != nil {
			m.openActions(wt)
		}
	case key.Matches(msg, keys.Create):
		m.mode = modeCreate
		m.branchInput.SetValue("")
		m.baseBranchInput.SetValue("")
		m.newBranch = true
		m.createFocus = 0
		m.statusMsg = ""
		// Fetch + load branches (fetch to get latest remote refs)
		_ = exec.Command("git", "fetch", "--prune").Run()
		m.allBranches = git.AllBranches()
		m.filteredBranches = nil
		m.suggestCursor = 0
		m.showSuggestions = false
		m.baseBranchInput.Blur()
		return m, m.branchInput.Focus()
	case key.Matches(msg, keys.Delete):
		if wt := m.selected(); wt != nil && !wt.IsBare && !wt.IsMain {
			m.mode = modeConfirmDelete
			m.statusMsg = ""
		}
	case key.Matches(msg, keys.Rider):
		return m.doAction(func(wt *git.Worktree) (string, error) {
			return "Opened in Rider", ide.OpenInRider(wt.Path)
		})
	case key.Matches(msg, keys.VSCode):
		return m.doAction(func(wt *git.Worktree) (string, error) {
			return "Opened in VS Code", ide.OpenInVSCode(wt.Path)
		})
	case key.Matches(msg, keys.Explorer):
		return m.doAction(func(wt *git.Worktree) (string, error) {
			return "Opened in Explorer", ide.OpenInExplorer(wt.Path)
		})
	case key.Matches(msg, keys.Codex):
		return m.doAction(func(wt *git.Worktree) (string, error) {
			return "Codex launched", agents.Launch(m.pm, agents.Codex, wt.Path)
		})
	case key.Matches(msg, keys.DockerUp):
		return m.doAction(func(wt *git.Worktree) (string, error) {
			return "Docker Compose starting...", docker.ComposeUp(m.pm, wt.Path)
		})
	case key.Matches(msg, keys.DockerDown):
		if wt := m.selected(); wt != nil {
			return m, func() tea.Msg {
				if err := docker.ComposeDown(wt.Path); err != nil {
					return actionResultMsg{msg: "Docker: " + err.Error(), err: true}
				}
				return actionResultMsg{msg: "Docker Compose stopped", err: false}
			}
		}
	case key.Matches(msg, keys.GitPull):
		if wt := m.selected(); wt != nil {
			cmd := m.setStatus("Pulling...", false)
			return m, tea.Batch(cmd, func() tea.Msg {
				if err := git.Pull(wt.Path); err != nil {
					return actionResultMsg{msg: "Pull: " + err.Error(), err: true}
				}
				return actionResultMsg{msg: "Pull complete", err: false}
			})
		}
	case key.Matches(msg, keys.GitFetch):
		if wt := m.selected(); wt != nil {
			cmd := m.setStatus("Fetching...", false)
			return m, tea.Batch(cmd, func() tea.Msg {
				if err := git.Fetch(wt.Path); err != nil {
					return actionResultMsg{msg: "Fetch: " + err.Error(), err: true}
				}
				return actionResultMsg{msg: "Fetch complete", err: false}
			})
		}
	case key.Matches(msg, keys.Refresh):
		return m, tea.Batch(loadWorktrees, refreshStatusCache)
	}
	return m, nil
}

// doAction is a helper for simple sync actions on the selected worktree.
func (m Model) doAction(fn func(*git.Worktree) (string, error)) (tea.Model, tea.Cmd) {
	wt := m.selected()
	if wt == nil {
		return m, nil
	}
	msg, err := fn(wt)
	if err != nil {
		cmd := m.setStatus(err.Error(), true)
		return m, cmd
	}
	cmd := m.setStatus(msg, false)
	return m, cmd
}

// ─── Actions Menu ────────────────────────────────────────

func (m *Model) openActions(wt *git.Worktree) {
	m.mode = modeActions
	m.actionCursor = 0
	m.actions = []action{
		{"o", "Open in Rider", func(m *Model) tea.Cmd {
			if err := ide.OpenInRider(wt.Path); err != nil {
				return m.setStatus("Rider: "+err.Error(), true)
			}
			return m.setStatus("Opened in Rider", false)
		}},
		{"v", "Open in VS Code", func(m *Model) tea.Cmd {
			if err := ide.OpenInVSCode(wt.Path); err != nil {
				return m.setStatus("VS Code: "+err.Error(), true)
			}
			return m.setStatus("Opened in VS Code", false)
		}},
		{"e", "Open in Explorer", func(m *Model) tea.Cmd {
			if err := ide.OpenInExplorer(wt.Path); err != nil {
				return m.setStatus("Explorer: "+err.Error(), true)
			}
			return m.setStatus("Opened in Explorer", false)
		}},
	}

	// Agents section
	agentKeys := map[string]string{
		"Codex":      "x",
		"Claude Code": "1",
		"Qwen":       "2",
	}
	for _, ag := range agents.All {
		ag := ag // capture
		k := agentKeys[ag.Name]
		m.actions = append(m.actions, action{k, ag.Name, func(m *Model) tea.Cmd {
			if err := agents.Launch(m.pm, ag, wt.Path); err != nil {
				return m.setStatus(ag.Name+": "+err.Error(), true)
			}
			return m.setStatus(ag.Name+" launched", false)
		}})
	}

	// Git section
	m.actions = append(m.actions,
		action{"g", "Git Pull", func(m *Model) tea.Cmd {
			return tea.Batch(m.setStatus("Pulling...", false), func() tea.Msg {
				if err := git.Pull(wt.Path); err != nil {
					return actionResultMsg{msg: "Pull: " + err.Error(), err: true}
				}
				return actionResultMsg{msg: "Pull complete", err: false}
			})
		}},
		action{"f", "Git Fetch", func(m *Model) tea.Cmd {
			return tea.Batch(m.setStatus("Fetching...", false), func() tea.Msg {
				if err := git.Fetch(wt.Path); err != nil {
					return actionResultMsg{msg: "Fetch: " + err.Error(), err: true}
				}
				return actionResultMsg{msg: "Fetch complete", err: false}
			})
		}},
	)

	if docker.HasComposeFile(wt.Path) {
		m.actions = append(m.actions,
			action{"u", "Docker Compose Up", func(m *Model) tea.Cmd {
				if err := docker.ComposeUp(m.pm, wt.Path); err != nil {
					return m.setStatus("Docker: "+err.Error(), true)
				}
				return m.setStatus("Docker Compose starting...", false)
			}},
			action{"s", "Docker Compose Down", func(m *Model) tea.Cmd {
				return tea.Batch(m.setStatus("Stopping...", false), func() tea.Msg {
					if err := docker.ComposeDown(wt.Path); err != nil {
						return actionResultMsg{msg: "Docker: " + err.Error(), err: true}
					}
					return actionResultMsg{msg: "Docker Compose stopped", err: false}
				})
			}},
		)
	}
}

func (m Model) handleActionsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Up):
		if m.actionCursor > 0 {
			m.actionCursor--
		}
	case key.Matches(msg, keys.Down):
		if m.actionCursor < len(m.actions)-1 {
			m.actionCursor++
		}
	case key.Matches(msg, keys.Confirm):
		if m.actionCursor < len(m.actions) {
			a := m.actions[m.actionCursor]
			m.mode = modeList
			cmd := a.fn(&m)
			return m, cmd
		}
	default:
		// Check if typed key matches an action shortcut
		k := msg.String()
		for _, a := range m.actions {
			if a.key == k {
				m.mode = modeList
				cmd := a.fn(&m)
				return m, cmd
			}
		}
	}
	return m, nil
}

// ─── Create ──────────────────────────────────────────────

func (m Model) handleCreateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Tab):
		if m.createFocus == 0 {
			// branch name → base branch
			m.createFocus = 1
			m.branchInput.Blur()
			m.showSuggestions = true
			m.filterBranches()
			return m, m.baseBranchInput.Focus()
		}
		// base branch → toggle mode → branch name
		m.createFocus = 0
		m.baseBranchInput.Blur()
		m.showSuggestions = false
		m.newBranch = !m.newBranch
		return m, m.branchInput.Focus()

	case key.Matches(msg, keys.Up):
		if m.createFocus == 1 && m.showSuggestions && len(m.filteredBranches) > 0 {
			if m.suggestCursor > 0 {
				m.suggestCursor--
			}
			return m, nil
		}

	case key.Matches(msg, keys.Down):
		if m.createFocus == 1 && m.showSuggestions && len(m.filteredBranches) > 0 {
			if m.suggestCursor < len(m.filteredBranches)-1 {
				m.suggestCursor++
			}
			return m, nil
		}

	case key.Matches(msg, keys.Confirm):
		// If in suggestions and have a selection, fill it in
		if m.createFocus == 1 && m.showSuggestions && len(m.filteredBranches) > 0 {
			m.baseBranchInput.SetValue(m.filteredBranches[m.suggestCursor].Name)
			m.showSuggestions = false
			return m, nil
		}

		name := strings.TrimSpace(m.branchInput.Value())
		if name == "" {
			cmd := m.setStatus("Branch name cannot be empty", true)
			return m, cmd
		}

		base := strings.TrimSpace(m.baseBranchInput.Value())
		if base == "" {
			base = m.cfg.BaseBranch
		}

		if err := git.Add(m.worktreeBase, name, base, m.newBranch); err != nil {
			cmd := m.setStatus(err.Error(), true)
			return m, cmd
		}

		m.mode = modeList
		cmd := m.setStatus(fmt.Sprintf("Created worktree '%s' from %s", name, base), false)
		return m, tea.Batch(cmd, loadWorktrees, refreshStatusCache)
	}

	// Pass to the focused input, then update suggestions
	var cmd tea.Cmd
	if m.createFocus == 1 {
		m.baseBranchInput, cmd = m.baseBranchInput.Update(msg)
		m.filterBranches()
	} else {
		m.branchInput, cmd = m.branchInput.Update(msg)
	}
	return m, cmd
}

func (m *Model) filterBranches() {
	query := strings.ToLower(strings.TrimSpace(m.baseBranchInput.Value()))
	m.filteredBranches = nil
	m.suggestCursor = 0

	if query == "" {
		m.filteredBranches = append(m.filteredBranches, m.allBranches...)
		return
	}

	for _, b := range m.allBranches {
		if strings.Contains(strings.ToLower(b.Name), query) {
			m.filteredBranches = append(m.filteredBranches, b)
		}
	}
}

// ─── Delete ──────────────────────────────────────────────

func (m Model) handleDeleteKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	wt := m.selected()

	switch msg.String() {
	case "y", "Y":
		if wt == nil {
			m.mode = modeList
			return m, nil
		}

		_ = m.pm.Stop(process.KindCodex, wt.Path)
		_ = m.pm.Stop(process.KindDocker, wt.Path)

		// If dirty and not force-confirmed, try normal remove first
		force := false
		if git.IsDirty(wt.Path) {
			force = true
		}

		if err := git.Remove(wt.Path, wt.Branch, force); err != nil {
			m.mode = modeList
			cmd := m.setStatus(err.Error(), true)
			return m, cmd
		}

		m.mode = modeList
		cmd := m.setStatus(fmt.Sprintf("Removed worktree '%s'", wt.Branch), false)
		return m, tea.Batch(cmd, loadWorktrees, refreshStatusCache)

	case "n", "N", "esc":
		m.mode = modeList
		m.statusMsg = ""
	}
	return m, nil
}

func (m Model) selected() *git.Worktree {
	if len(m.worktrees) == 0 || m.cursor >= len(m.worktrees) {
		return nil
	}
	return &m.worktrees[m.cursor]
}
