package process

import (
	"fmt"
	"os/exec"
	"sync"
)

// Kind identifies the type of a managed process.
type Kind string

const (
	KindCodex  Kind = "codex"
	KindDocker Kind = "docker"
)

// Entry represents a running process tracked by the manager.
type Entry struct {
	Kind    Kind
	Label   string
	WTPath  string
	Cmd     *exec.Cmd
	Done    chan struct{}
	Err     error
}

// Manager tracks child processes spawned from the TUI.
type Manager struct {
	mu      sync.Mutex
	entries map[string]*Entry // key: "kind:wtPath"
}

// New creates a new process manager.
func New() *Manager {
	return &Manager{entries: make(map[string]*Entry)}
}

func key(kind Kind, wtPath string) string {
	return fmt.Sprintf("%s:%s", kind, wtPath)
}

// Start launches a command and tracks it.
func (m *Manager) Start(kind Kind, label, wtPath string, cmd *exec.Cmd) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	k := key(kind, wtPath)
	if e, ok := m.entries[k]; ok {
		select {
		case <-e.Done:
			// previous process finished, allow restart
		default:
			return fmt.Errorf("%s is already running for %s", kind, wtPath)
		}
	}

	done := make(chan struct{})
	entry := &Entry{
		Kind:   kind,
		Label:  label,
		WTPath: wtPath,
		Cmd:    cmd,
		Done:   done,
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start %s: %w", kind, err)
	}

	go func() {
		entry.Err = cmd.Wait()
		close(done)
	}()

	m.entries[k] = entry
	return nil
}

// IsRunning checks whether a process of the given kind is active for a worktree.
func (m *Manager) IsRunning(kind Kind, wtPath string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	e, ok := m.entries[key(kind, wtPath)]
	if !ok {
		return false
	}
	select {
	case <-e.Done:
		return false
	default:
		return true
	}
}

// Stop terminates a tracked process.
func (m *Manager) Stop(kind Kind, wtPath string) error {
	m.mu.Lock()
	e, ok := m.entries[key(kind, wtPath)]
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("no %s process for %s", kind, wtPath)
	}

	select {
	case <-e.Done:
		return nil // already done
	default:
	}

	if e.Cmd.Process != nil {
		return e.Cmd.Process.Kill()
	}
	return nil
}

// Running returns all currently running entries.
func (m *Manager) Running() []*Entry {
	m.mu.Lock()
	defer m.mu.Unlock()

	var result []*Entry
	for _, e := range m.entries {
		select {
		case <-e.Done:
			continue
		default:
			result = append(result, e)
		}
	}
	return result
}

// RunningFor returns running processes for a specific worktree.
func (m *Manager) RunningFor(wtPath string) []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	var labels []string
	for _, e := range m.entries {
		if e.WTPath != wtPath {
			continue
		}
		select {
		case <-e.Done:
			continue
		default:
			labels = append(labels, string(e.Kind))
		}
	}
	return labels
}

// StopAll kills all tracked processes.
func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, e := range m.entries {
		select {
		case <-e.Done:
		default:
			if e.Cmd.Process != nil {
				_ = e.Cmd.Process.Kill()
			}
		}
	}
}
